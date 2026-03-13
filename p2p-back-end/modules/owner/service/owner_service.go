package service

import (
	"strings"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

type ownerService struct {
	repo     models.OwnerRepository
	authSrv  models.AuthService
	capexSrv models.CapexService
}

func NewOwnerService(repo models.OwnerRepository, authSrv models.AuthService, capexSrv models.CapexService) models.OwnerService {
	return &ownerService{repo: repo, authSrv: authSrv, capexSrv: capexSrv}
}

func (s *ownerService) GetDashboardSummary(user *models.UserInfo, filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	logs.Infof("[DEBUG] OwnerService: GetDashboardSummary START for User=%s", user.UserName)

	filter = s.injectPermissions(user, filter)

	// 🛠️ Key Mapping: Frontend Owner uses 'conso_gls', Backend Repo uses 'budget_gls'
	if gls, ok := filter["conso_gls"]; ok {
		filter["budget_gls"] = gls
	}

	// 1. Fetch PL Summary from own OwnerRepository (Pure SRP)
	summary, err := s.repo.GetDashboardAggregates(filter)
	if err != nil {
		logs.Errorf("[ERROR] OwnerService: DashboardSummary Failed: %v", err)
		return nil, err
	}

	// 2. Fetch Capex Summary from CapexService
	capexSummary, err := s.capexSrv.GetCapexDashboardSummary(filter)
	capexBudget := 0.0
	capexActual := 0.0
	if err == nil && capexSummary != nil {
		capexBudget = capexSummary.TotalBudget
		capexActual = capexSummary.TotalActual
	}

	// 3. Top Expenses Logic (By Department for now)
	topExp := []models.TopExpenseDTO{}
	for i, d := range summary.DepartmentData {
		if i >= 5 {
			break
		}
		if d.Actual > 0 {
			topExp = append(topExp, models.TopExpenseDTO{
				Name:   d.Department,
				Amount: d.Actual,
			})
		}
	}

	logs.Infof("[DEBUG] OwnerService: Result - Budget=%.2f, Actual=%.2f, DeptsCount=%d", summary.TotalBudget, summary.TotalActual, len(summary.DepartmentData))

	return &models.OwnerDashboardSummaryDTO{
		DashboardSummaryDTO: *summary,
		TopExpenses:         topExp,
		CapexBudget:         capexBudget,
		CapexActual:         capexActual,
	}, nil
}

func (s *ownerService) GetActualTransactions(user *models.UserInfo, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetActualTransactions(filter)
}

func (s *ownerService) GetActualDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetActualDetails(filter)
}

func (s *ownerService) GetBudgetDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	filter = s.injectPermissions(user, filter)
	return s.repo.GetBudgetDetails(filter)
}

func (s *ownerService) GetFilterOptions(user *models.UserInfo) (interface{}, error) {
	filter := s.injectPermissions(user, nil)
	allowedDepts, _ := filter["departments"].([]string)

	allFacts, err := s.repo.GetBudgetFilterOptions()
	if err != nil {
		return nil, err
	}

	isAdmin := false
	if restricted, ok := filter["is_restricted"].(bool); !ok || !restricted {
		isAdmin = true
	}

	if isAdmin {
		return allFacts, nil
	}

	deptMap := make(map[string]bool)
	for _, d := range allowedDepts {
		deptMap[d] = true
	}

	var filtered []models.BudgetFactEntity
	for _, fact := range allFacts {
		if deptMap[fact.Department] {
			filtered = append(filtered, fact)
		}
	}

	return filtered, nil
}

func (s *ownerService) GetOrganizationStructure(user *models.UserInfo) ([]models.OrganizationDTO, error) {
	filter := s.injectPermissions(user, nil)
	allowedDepts, _ := filter["departments"].([]string)

	facts, err := s.repo.GetOrganizationStructure()
	if err != nil {
		return nil, err
	}

	isAdmin := false
	if restricted, ok := filter["is_restricted"].(bool); !ok || !restricted {
		isAdmin = true
	}

	// 1. Build the full structure from facts
	structure := make(map[string]map[string][]string)
	for _, f := range facts {
		if f.Entity == "" {
			continue
		}
		if structure[f.Entity] == nil {
			structure[f.Entity] = make(map[string][]string)
		}
		if f.Branch != "" {
			if _, exists := structure[f.Entity][f.Branch]; !exists {
				structure[f.Entity][f.Branch] = []string{}
			}
			if f.Department != "" {
				structure[f.Entity][f.Branch] = append(structure[f.Entity][f.Branch], f.Department)
			}
		}
	}

	// 2. Convert to DTOs and filter if not admin
	deptMap := make(map[string]bool)
	for _, d := range allowedDepts {
		deptMap[d] = true
	}

	var results []models.OrganizationDTO
	for entName, branchMap := range structure {
		var branches []models.BranchDTO
		for brName, depts := range branchMap {
			var filteredDepts []string
			if isAdmin {
				filteredDepts = depts
			} else {
				for _, d := range depts {
					if deptMap[d] {
						filteredDepts = append(filteredDepts, d)
					}
				}
			}

			if len(filteredDepts) > 0 {
				branches = append(branches, models.BranchDTO{
					Name:        brName,
					Departments: filteredDepts,
				})
			}
		}

		if len(branches) > 0 {
			results = append(results, models.OrganizationDTO{
				Entity:   entName,
				Branches: branches,
			})
		}
	}

	return results, nil
}

func (s *ownerService) GetOwnerFilterLists(user *models.UserInfo) (*models.OwnerFilterListsDTO, error) {
	years, err := s.repo.GetActualYears()
	if err != nil {
		logs.Errorf("[ERROR] OwnerService: Failed to get actual years: %v", err)
	}

	lists := &models.OwnerFilterListsDTO{
		Companies: []string{"HMW", "ACG", "CLIK"}, // Hardcoded defaults for now
		Branches:  []string{},
		Years:     years,
	}
	return lists, nil
}

func (s *ownerService) injectPermissions(user *models.UserInfo, filter map[string]interface{}) map[string]interface{} {
	if filter == nil {
		filter = make(map[string]interface{})
	}

	// 🛡️ Proactively Fetch User Permissions if missing from context
	if len(user.Permissions) == 0 && user.UserId != "" {
		dbPerms, err := s.authSrv.GetUserPermissions(user.UserId)
		if err == nil {
			user.Permissions = dbPerms
		}
	}

	roles := user.Roles

	// 🕵️ Determine if user is Admin or Restricted Owner
	isAdmin := false

	// Check 1: Role-based
	for _, r := range roles {
		rUpper := strings.ToUpper(strings.TrimSpace(r))
		if rUpper == "ADMIN" || rUpper == "ADMINISTRATOR" || strings.Contains(rUpper, "ADMIN") {
			isAdmin = true
			break
		}
	}

	// Check 2: Username-based (safety fallback)
	userNameLower := strings.ToLower(user.UserName)
	if userNameLower == "admin" || userNameLower == "administrator" || strings.Contains(userNameLower, "admin") {
		isAdmin = true
	}

	// 🛠️ Logic Refinement: If you are an Admin, you are NEVER restricted by specific owner permissions.
	// This addresses the user's concern: only NON-ADMIN owners will be restricted by their permissions list.
	logs.Infof("[DEBUG] injectPermissions: User=%s, Roles=%v, Final isAdmin=%v, permsCount=%d", user.UserName, roles, isAdmin, len(user.Permissions))

	if !isAdmin {
		// 🛠️ Construct allowed departments list
		allowedDepts := make([]string, 0)
		for _, p := range user.Permissions {
			if p.IsActive && p.DepartmentCode != "" {
				allowedDepts = append(allowedDepts, strings.TrimSpace(p.DepartmentCode))
			}
		}



		logs.Infof("[DEBUG] injectPermissions: Non-Admin Owner detected. AllowedDepts: %v", allowedDepts)

		if len(allowedDepts) == 0 {
			// If not admin and NO permissions assigned, they should see NOTHING
			logs.Warnf("[WARN] injectPermissions: User %s has no Department Permissions assigned!", user.UserName)
			filter["departments"] = []string{"__RESTRICTED__"}
		} else {
			// If frontend already provided departments, intersect them with allowed ones
			if val, ok := filter["departments"]; ok {
				var finalDepts []string
				var chosenDepts []string

				// Safe conversion
				if strs, ok := val.([]string); ok {
					chosenDepts = strs
				} else if interfaces, ok := val.([]interface{}); ok {
					for _, i := range interfaces {
						if s, ok := i.(string); ok {
							chosenDepts = append(chosenDepts, s)
						}
					}
				}

				if len(chosenDepts) > 0 {
					for _, c := range chosenDepts {
						cTrim := strings.TrimSpace(c)
						for _, a := range allowedDepts {
							if cTrim == a {
								finalDepts = append(finalDepts, c)
								break
							}
						}
					}
					if len(finalDepts) == 0 {
						finalDepts = []string{"__RESTRICTED__"}
					}
					filter["departments"] = finalDepts
				} else {
					filter["departments"] = allowedDepts
				}
			} else {
				filter["departments"] = allowedDepts
			}
		}
		filter["is_restricted"] = true
		logs.Debugf("[DEBUG] injectPermissions: Final Departments Filter: %v", filter["departments"])
	}

	return filter
}
