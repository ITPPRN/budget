package service

import (
	"context"
	"errors"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/organization/repository"
	"p2p-back-end/pkg/utils"
)

type departmentService struct {
	repo repository.DepartmentRepositoryDB
}

func NewDepartmentService(repo repository.DepartmentRepositoryDB) models.DepartmentService {
	return &departmentService{repo: repo}
}

// ManageDepartments handles the incremental seeding of Department Data (Preserves IDs)
func (s *departmentService) ManageDepartments(ctx context.Context) error {
	logs.Info("Starting Department Sync (Preserving Existing IDs)...")

	// 1. Fetch Existing Departments to avoid duplicate inserts and ID changes
	existingDepts, err := s.repo.GetDepartmentMap()
	if err != nil {
		return err
	}

	// 2. Prepare Master Data & Mappings
	// type MappingRow struct {
	// 	Master string
	// 	ACG    []string
	// 	HMW    []string
	// 	CLIK   []string
	// }

	// --- Unified Mapping from utils ---
	mappingData := utils.DepartmentMappingData

	// 3. Insert ONLY Missing Master Departments
	var newDepts []models.DepartmentEntity
	for _, row := range mappingData {
		if _, exists := existingDepts[row.Master]; !exists {
			newDepts = append(newDepts, models.DepartmentEntity{
				Code: row.Master,
				Name: row.Master,
			})
		}
	}

	if len(newDepts) > 0 {
		logs.Infof("Seeding %d new master departments...", len(newDepts))
		if err := s.repo.CreateDepartmentsBatch(newDepts); err != nil {
			return err
		}
		// Refresh map to include new items
		existingDepts, err = s.repo.GetDepartmentMap()
		if err != nil {
			return err
		}
	}

	// 4. Refresh Mappings (Truncate Mappings table only)
	if err := s.repo.ClearMappings(); err != nil {
		return err
	}

	var mappings []models.DepartmentMappingEntity
	for _, row := range mappingData {
		dept, exists := existingDepts[row.Master]
		if !exists {
			continue
		}

		addMappings := func(codes []string, entity string) {
			for _, code := range codes {
				if code == "" {
					continue
				}
				mappings = append(mappings, models.DepartmentMappingEntity{
					DepartmentID: dept.ID,
					Entity:       entity,
					NavCode:      code,
				})
			}
		}

		addMappings(row.ACG, "ACG")
		addMappings(row.HMW, "HMW")
		addMappings(row.CLIK, "CLIK")
	}

	if len(mappings) > 0 {
		if err := s.repo.CreateMappingsBatch(mappings); err != nil {
			return err
		}
	}

	// 5. Populate CodeMap in the 'departments' table (Integrated Mapping)
	var allGranularDepts []models.Departments
	granularMap := make(map[string]models.Departments)
	if err := s.repo.GetDB().Find(&allGranularDepts).Error; err == nil {
		// Create a lookup map for faster processing: GranularCode -> MasterCode
		granularToMaster := make(map[string]string)
		for _, row := range mappingData {
			for _, c := range row.ACG {
				granularToMaster[c] = row.Master
			}
			for _, c := range row.HMW {
				granularToMaster[c] = row.Master
			}
			for _, c := range row.CLIK {
				granularToMaster[c] = row.Master
			}
		}

		logs.Infof("Updating CodeMap for %d granular departments...", len(allGranularDepts))
		for _, gd := range allGranularDepts {
			granularMap[gd.Code] = gd
			if master, ok := granularToMaster[gd.Code]; ok {
				if gd.CodeMap == nil || *gd.CodeMap != master {
					s.repo.GetDB().Model(&gd).Update("code_map", master)
				}
			} else if gd.CodeMap != nil {
				s.repo.GetDB().Model(&gd).Update("code_map", nil)
			}
		}
	}

	// 6. Repair existing users with broken or missing department links
	// Now linking to the 'departments' table instead of 'department_entities'
	var usersToFix []models.UserEntity
	s.repo.GetDB().Table("user_entities").
		Joins("LEFT JOIN departments ON departments.id = user_entities.department_id").
		Where("user_entities.department_id IS NULL OR departments.id IS NULL").
		Find(&usersToFix)

	if len(usersToFix) > 0 {
		logs.Infof("Repair: Checking %d users for missing or broken department links...", len(usersToFix))
		for _, u := range usersToFix {
			// fixed := false
			var targetDeptCode string

			// 1. Try to get code from ACTIVE permissions
			var perms []models.UserPermissionEntity
			s.repo.GetDB().Where("user_id = ? AND is_active = true", u.ID).Find(&perms)
			if len(perms) > 0 {
				targetDeptCode = perms[0].DepartmentCode
			}

			// 2. Fallback to Section -> Dept mapping if not found in perms
			if targetDeptCode == "" && u.SectionID != nil {
				var section models.Sections
				if err := s.repo.GetDB().Preload("Department").Where("id = ?", u.SectionID).First(&section).Error; err == nil {
					if section.Department != nil {
						targetDeptCode = section.Department.Code
					}
				}
			}

			if targetDeptCode != "" {
				if dept, ok := granularMap[targetDeptCode]; ok {
					logs.Infof("   -> Fixing user %s: Linking to Granular Dept [%s] (ID: %s)", u.Username, dept.Code, dept.ID)
					s.repo.GetDB().Model(&u).Update("department_id", dept.ID)
					// fixed = true
				}
			}

			// if !fixed {
			// 	// Comment out to reduce log spam as requested by user
			// 	// logs.Warnf("   -> Fix Failed: User %s (%s) has no valid granular department match for code '%s'.", u.Username, u.NameTh, targetDeptCode)
			// }
		}
	}

	logs.Info("Department Sync Completed Successfully (Preserved Existing IDs).")
	return nil
}

// GetMasterDepartment finds the Master Department using the mapping table
func (s *departmentService) GetMasterDepartment(ctx context.Context, navCode, entity string) (*models.DepartmentEntity, error) {
	if navCode == "" {
		// Return "None" directly to avoid "record not found" error on empty lookup
		// and to ensure data isn't lost/blank in dashboard
		noneDept, err := s.repo.FindDepartmentByCode("None")
		if err != nil {
			return nil, err
		}
		return noneDept, nil
	}
	mapping, err := s.repo.FindMappingByNavCode(entity, navCode)
	if err != nil {
		return nil, err
	}
	// Handle nil mapping (Not Found) -> Fallback to "None"
	if mapping == nil {
		noneDept, err := s.repo.FindDepartmentByCode("None")
		if err != nil {
			return nil, err // "None" dept missing? causing issues
		}
		if noneDept == nil {
			// If even "None" is missing, returns nil (should not happen if seeded)
			return nil, nil
		}
		return noneDept, nil
	}
	// Ensure Department is loaded
	if mapping.Department == nil {
		return nil, errors.New("mapping found but Department not loaded")
	}
	return mapping.Department, nil
}
