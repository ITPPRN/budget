package service

import (
	"context"
	"errors"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/organization/repository"
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
	type MappingRow struct {
		Master string
		ACG    []string
		HMW    []string
		CLIK   []string
	}

	// DATA MATRIX from User Request
	data := []MappingRow{
		{Master: "ACC",
			ACG:  []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC-MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"},
			HMW:  []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC-MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"},
			CLIK: []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC_MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"}},
		{Master: "BUDGET",
			ACG:  []string{"BUDGET"},
			HMW:  []string{"BUDGET"},
			CLIK: []string{"BUDGET"}},
		{Master: "DPO",
			ACG:  []string{"DM", "DPO"},
			HMW:  []string{"DM", "DPO"},
			CLIK: []string{"DM", "DPO"}},
		{Master: "FIN",
			ACG:  []string{"FIN", "FIN-CENTER", "FIN_CENTER"},
			HMW:  []string{"FIN", "FIN-CENTER", "FIN_CENTER"},
			CLIK: []string{"FIN", "FIN-CENTER", "FIN_CENTER"}},
		{Master: "G-CENTER",
			ACG:  []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"},
			HMW:  []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"},
			CLIK: []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"}},
		{Master: "G-CR",
			ACG:  []string{"G-CR", "G_CR"},
			HMW:  []string{"G-CR", "G_CR"},
			CLIK: []string{"G-CR", "G_CR"}},
		{Master: "G-HR",
			ACG:  []string{"G-HR", "G_HR"},
			HMW:  []string{"G-HR", "G_HR"},
			CLIK: []string{"G-HR", "G_HR"}},
		{Master: "G-MAINT",
			ACG:  []string{"G-MAINT", "G_MAINT"},
			HMW:  []string{"G-MAINT", "G_MAINT"},
			CLIK: []string{"G-MAINT", "G_MAINT"}},
		{Master: "G-PC",
			ACG:  []string{"G-PC", "G_PC"},
			HMW:  []string{"G-PC", "G_PC"},
			CLIK: []string{"G-PC", "G_PC"}},
		{Master: "G-SALARY",
			ACG:  []string{"G-SALARY", "G_SALARY"},
			HMW:  []string{"G-SALARY", "G_SALARY"},
			CLIK: []string{"G-SALARY", "G_SALARY"}},
		{Master: "IA",
			ACG:  []string{"IA", "G-IA"},
			HMW:  []string{"IA", "G-IA"},
			CLIK: []string{"IA", "G-IA"}},
		{Master: "IT-DEVELOP",
			ACG:  []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"},
			HMW:  []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"},
			CLIK: []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"}},
		{Master: "IT-SUPPORT",
			ACG:  []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"},
			HMW:  []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"},
			CLIK: []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"}},
		{Master: "MARKETING",
			ACG:  []string{"MARKETING", "MAREKTING"}, // Typos
			HMW:  []string{"MARKETING", "MAREKTING"},
			CLIK: []string{"MARKETING", "MAREKTING"}},
		{Master: "MGMT",
			ACG:  []string{"MGMT"},
			HMW:  []string{"MGMT"},
			CLIK: []string{"MGMT"}},
		{Master: "SALE-CENTER",
			ACG:  []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"},
			HMW:  []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"},
			CLIK: []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"}},
		{Master: "SALE-INSURE",
			ACG:  []string{"SALE-INSURE", "SALE_INSURE"},
			HMW:  []string{"SALE-INSURE", "SALE_INSURE"},
			CLIK: []string{"SALE-INSURE", "SALE_INSURE"}},
		{Master: "SECRETARY",
			ACG:  []string{"IR", "SECRETARY"},
			HMW:  []string{"IR", "SECRETARY"},
			CLIK: []string{"IR", "SECRETARY"}},
		{Master: "SERVICE",
			ACG:  []string{"BP", "G-PDI", "G_PDI", "GRPM", "G-STORE", "G_STORE", "SERVICE", "SERVICE-BP", "SERVICE_BP", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERVICE-CENTER", "SERVICE_CENTER"},
			HMW:  []string{"BP1", "G-PDI", "G_PDI", "GRPM", "G-STORE", "G_STORE", "SERVICE", "SERVICE-BP", "SERVICE_BP", "SERVICE-CENTER", "SERVICE_CENTER", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERIVCE-ADMIN", "SERVICE_ADMIN", "SERVICE-CASHIER", "SERVICE_CASHIER", "BP2"},
			CLIK: []string{"G-PDI", "G_PDI", "G-STORE", "G_STORE", "SERVICE-BP", "SERVICE_BP", "SERVICE-CENTER", "SERVICE_CENTER", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERVICE-ADMIN", "SERVICE_ADMIN", "SERVICE-CASHIER", "SERVICE_CASHIER"}},
		{Master: "SERVICE_CLIK",
			ACG:  []string{},
			HMW:  []string{},
			CLIK: []string{"SERVICE", "SERVICE_CLIK", "SERVICE-CLIK", "SERVICE CLIK"}},
		{Master: "STRATEGY",
			ACG:  []string{"STRATEGY"},
			HMW:  []string{"STRATEGY"},
			CLIK: []string{"STRATEGY"}},
		{Master: "TRAINING",
			ACG:  []string{"TRAINING"},
			HMW:  []string{"TRAINING"},
			CLIK: []string{"TRAINING"}},
		{Master: "None",
			ACG:  []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "G-SECURITY", "G_SECURITY", "INS", "IT", "IT-CENTER", "IT_CENTER", "MNG", "PC", "SALE", "STORE", "TECH", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE", "REG"},
			HMW:  []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "G-SECURITY", "G_SECURITY", "INS", "IT", "IT-CENTER", "IT_CENTER", "MNG", "PC", "SALE", "STORE", "TECH", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE", "REG"},
			CLIK: []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "IT", "IT-CENTER", "IT_CENTER", "PC", "PLANNING", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE"}},
	}

	// 3. Insert ONLY Missing Master Departments
	var newDepts []models.DepartmentEntity
	for _, row := range data {
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
	for _, row := range data {
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

	// 5. Repair existing users with broken or missing department links
	var usersToFix []models.UserEntity
	// Find users where department_id is NULL OR current link doesn't exist anymore
	s.repo.GetDB().Table("user_entities").
		Joins("LEFT JOIN department_entities ON department_entities.id = user_entities.department_id").
		Where("user_entities.department_id IS NULL OR department_entities.id IS NULL").
		Find(&usersToFix)

	if len(usersToFix) > 0 {
		logs.Infof("Repair: Checking %d users for missing or broken department links...", len(usersToFix))
		for _, u := range usersToFix {
			fixed := false
			// 1. Try to Fix based on ACTIVE permissions
			var perms []models.UserPermissionEntity
			s.repo.GetDB().Where("user_id = ? AND is_active = true", u.ID).Find(&perms)

			if len(perms) > 0 {
				masterDept, err := s.GetMasterDepartment(ctx, perms[0].DepartmentCode, "ACG")
				if err == nil && masterDept != nil {
					logs.Infof("   -> Fixing user %s: Linking to Master [%s] %s (from Permission Path %s)", u.Username, masterDept.Code, masterDept.Name, perms[0].DepartmentCode)
					s.repo.GetDB().Model(&u).Update("department_id", masterDept.ID)
					fixed = true
				}
			}

			// 2. Try to Fix based on SectionID mapping if not fixed by permissions
			if !fixed && u.SectionID != nil {
				var section models.Sections
				if err := s.repo.GetDB().Where("id = ?", u.SectionID).First(&section).Error; err == nil {
					// 2.1 Only try to fetch Child Dept if ID is valid (!= nil)
					if section.DepartmentID != nil {
						var childDept models.Departments
						if err := s.repo.GetDB().Where("id = ?", section.DepartmentID).First(&childDept).Error; err == nil {
							masterDept, err := s.GetMasterDepartment(ctx, childDept.Code, "ACG")
							if err == nil && masterDept != nil {
								logs.Infof("   -> Fixing user %s: Linking to Master [%s] %s (from Section %s -> Dept %s)", u.Username, masterDept.Code, masterDept.Name, section.Code, childDept.Code)
								s.repo.GetDB().Model(&u).Update("department_id", masterDept.ID)
								fixed = true
							} else {
								logs.Warnf("   -> Repair Note: Found Section %s for user %s, but Dept code %s has no Master mapping.", section.Code, u.Username, childDept.Code)
							}
						} else {
							logs.Warnf("   -> Repair Note: Found Section %s for user %s, but linked Department ID %d not found in local DB.", section.Code, u.Username, section.DepartmentID)
						}
					} else {
						logs.Warnf("   -> Repair Note: Section %s assigned to user %s has DepartmentID=0 (Master Data Gap).", section.Code, u.Username)
					}
				} else {
					logs.Warnf("   -> Repair Note: SectionID %d on user %s not found in local DB.", u.SectionID, u.Username)
				}
			}

			if !fixed {
				// We leave it as NULL as per user request.
				// This allows us to identify users with missing mapping (shows as '-' in UI)
				logs.Warnf("   -> Fix Failed: User %s (%s) has no valid Permission Path or Section/Dept mapping.", u.Username, u.NameTh)
			}
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
		return nil, errors.New("Mapping found but Department not loaded")
	}
	return mapping.Department, nil
}
