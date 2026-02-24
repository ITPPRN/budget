package service

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"strings"
	"sync"
)

type ownerService struct {
	repo      models.OwnerRepository
	syncMutex sync.Mutex
	isSyncing bool
}

func NewOwnerService(repo models.OwnerRepository) models.OwnerService {
	return &ownerService{repo: repo}
}

func (s *ownerService) GetDashboardSummary(user *models.UserInfo, filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetDashboardAggregates(filter)
}

func (s *ownerService) GetActualTransactions(user *models.UserInfo, filter map[string]interface{}) ([]models.ActualTransactionDTO, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetActualTransactions(filter)
}

func (s *ownerService) GetActualDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.OwnerActualFactEntity, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetActualDetails(filter)
}

func (s *ownerService) GetBudgetDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetBudgetDetails(filter)
}

func (s *ownerService) GetFilterOptions(user *models.UserInfo) ([]models.FilterOptionDTO, error) {
	filter := s.injectPermissions(user, nil)
	rawOptions, err := s.repo.GetBudgetFilterOptions(filter)
	if err != nil {
		return nil, err
	}

	return s.mapFactsToFilterOptions(rawOptions), nil
}

func (s *ownerService) GetOrganizationStructure(user *models.UserInfo) ([]models.OrganizationDTO, error) {
	filter := s.injectPermissions(user, nil)
	facts, err := s.repo.GetBudgetFilterOptions(filter)
	if err != nil {
		return nil, err
	}

	// Map Entity -> Map Branch -> []Departments
	structure := make(map[string]map[string][]string)
	for _, f := range facts {
		if f.Entity == "" {
			continue
		}

		entityName := f.Entity
		if structure[entityName] == nil {
			structure[entityName] = make(map[string][]string)
		}

		if f.Branch != "" {
			branchName := f.Branch
			if _, exists := structure[entityName][branchName]; !exists {
				structure[entityName][branchName] = []string{}
			}

			if f.Department != "" {
				deptName := f.Department
				found := false
				for _, d := range structure[entityName][branchName] {
					if d == deptName {
						found = true
						break
					}
				}
				if !found {
					structure[entityName][branchName] = append(structure[entityName][branchName], deptName)
				}
			}
		}
	}

	var result []models.OrganizationDTO
	for entity, branchesMap := range structure {
		var branchDTOs []models.BranchDTO
		for branch, depts := range branchesMap {
			branchDTOs = append(branchDTOs, models.BranchDTO{
				Name:        branch,
				Departments: depts,
			})
		}
		result = append(result, models.OrganizationDTO{
			Entity:   entity,
			Branches: branchDTOs,
		})
	}
	return result, nil
}

func (s *ownerService) GetOwnerFilterLists(user *models.UserInfo) (*models.OwnerFilterListsDTO, error) {
	filter := s.injectPermissions(user, nil)
	return s.repo.GetOwnerFilterLists(filter)
}

func (s *ownerService) AutoSyncOwnerActuals() error {
	s.syncMutex.Lock()
	if s.isSyncing {
		s.syncMutex.Unlock()
		fmt.Println("[DEBUG] Owner Actuals Sync skipped: already in progress")
		return nil
	}
	s.isSyncing = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		s.isSyncing = false
		s.syncMutex.Unlock()
	}()

	fmt.Println("[DEBUG] Starting Owner Actuals Sync...")
	return s.repo.AutoSyncOwnerActuals()
}

func (s *ownerService) injectPermissions(user *models.UserInfo, filter map[string]interface{}) map[string]interface{} {
	if filter == nil {
		filter = make(map[string]interface{})
	}

	isAdmin := false
	for _, role := range user.Roles {
		roleUpper := strings.ToUpper(role)
		// Robust Admin Check: Support "ADMIN", "admin", and any role containing "ADMIN"
		if roleUpper == models.RoleAdmin || strings.Contains(roleUpper, "ADMIN") {
			isAdmin = true
			break
		}
	}

	// 🔍 Debug Log (Visible in server terminal)
	fmt.Printf("[DEBUG] injectPermissions: UserName=%s, Roles=%v, isAdmin=%v\n", user.UserName, user.Roles, isAdmin)

	if !isAdmin {
		// 🛡️ Load Missing Permissions Proactively
		if len(user.Permissions) == 0 && user.UserId != "" {
			dbPerms, err := s.repo.GetUserPermissions(user.UserId)
			if err == nil && len(dbPerms) > 0 {
				fmt.Printf("[DEBUG] injectPermissions: Loaded %d permissions from DB for UserID=%s\n", len(dbPerms), user.UserId)
				for _, p := range dbPerms {
					user.Permissions = append(user.Permissions, models.UserPermissionInfo{
						DepartmentCode: p.DepartmentCode,
						Role:           p.Role,
						IsActive:       p.IsActive != nil && *p.IsActive,
					})
				}
			}
		}

		var allowedDepts []string
		for _, p := range user.Permissions {
			if p.IsActive {
				allowedDepts = append(allowedDepts, p.DepartmentCode)
			}
		}
		// If user is not Admin, and has NO permissions, results should be restricted to nothing.
		filter["allowed_departments"] = allowedDepts
		filter["is_restricted"] = true
		fmt.Printf("[DEBUG] injectPermissions: Restricting to Depts: %v\n", allowedDepts)
	}

	return filter
}

// Helper: Map Facts to Tree (Entity -> Branch -> Department)
func (s *ownerService) mapFactsToFilterOptions(facts []models.BudgetFactEntity) []models.FilterOptionDTO {
	// Root: Entities
	entityMap := make(map[string]*models.FilterOptionDTO)

	for _, f := range facts {
		entName := f.Entity
		if entName == "" {
			continue
		}

		// 1. Entity Level
		if _, ok := entityMap[entName]; !ok {
			entityMap[entName] = &models.FilterOptionDTO{
				ID:       entName,
				Name:     entName,
				Level:    1,
				Children: []models.FilterOptionDTO{},
			}
		}
		entNode := entityMap[entName]

		// 2. Branch Level
		branchName := f.Branch
		if branchName == "" {
			branchName = "Head Office"
		}

		var branchNode *models.FilterOptionDTO
		// Find existing branch child
		for i := range entNode.Children {
			if entNode.Children[i].ID == branchName {
				branchNode = &entNode.Children[i]
				break
			}
		}
		if branchNode == nil {
			// Create new branch node
			newBranch := models.FilterOptionDTO{
				ID:       branchName,
				Name:     branchName,
				Level:    2,
				Children: []models.FilterOptionDTO{},
			}
			entNode.Children = append(entNode.Children, newBranch)
			branchNode = &entNode.Children[len(entNode.Children)-1]
		}

		// 3. Department Level
		deptName := f.Department
		if deptName == "" {
			continue
		}

		var deptNode *models.FilterOptionDTO
		for i := range branchNode.Children {
			if branchNode.Children[i].ID == deptName {
				deptNode = &branchNode.Children[i]
				break
			}
		}
		if deptNode == nil {
			newDept := models.FilterOptionDTO{
				ID:       deptName,
				Name:     deptName,
				Level:    3,
				Children: []models.FilterOptionDTO{},
			}
			branchNode.Children = append(branchNode.Children, newDept)
			deptNode = &branchNode.Children[len(branchNode.Children)-1]
		}

		// 4. NavCode Level (New)
		navName := f.NavCode
		if navName == "" {
			navName = "Unspecified"
		}

		navExists := false
		for _, child := range deptNode.Children {
			if child.ID == navName {
				navExists = true
				break
			}
		}
		if !navExists {
			deptNode.Children = append(deptNode.Children, models.FilterOptionDTO{
				ID:    navName,
				Name:  navName,
				Level: 4,
			})
		}
	}

	// Convert Map to Slice
	var results []models.FilterOptionDTO
	for _, v := range entityMap {
		results = append(results, *v)
	}
	return results
}
