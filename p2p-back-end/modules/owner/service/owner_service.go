package service

import (
	"context"
	"fmt"
	"strings"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"

	"github.com/shopspring/decimal"
)

type ownerService struct {
	repo     models.OwnerRepository
	authSrv  models.AuthService
	capexSrv models.CapexService
}

func NewOwnerService(repo models.OwnerRepository, authSrv models.AuthService, capexSrv models.CapexService) models.OwnerService {
	return &ownerService{repo: repo, authSrv: authSrv, capexSrv: capexSrv}
}

func (s *ownerService) GetDashboardSummary(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	logs.Infof("[DEBUG] OwnerService: GetDashboardSummary START for User=%s", user.Username)

	filter = s.InjectPermissions(ctx, user, filter)

	// 🛠️ Key Mapping: Frontend Owner uses 'conso_gls', Backend Repo uses 'budget_gls'
	if gls, ok := filter["conso_gls"]; ok {
		filter["budget_gls"] = gls
	}

	// 1. Fetch PL Summary from own OwnerRepository (Pure SRP)
	summary, err := s.repo.GetDashboardAggregates(ctx, filter)
	if err != nil {
		logs.Errorf("[ERROR] OwnerService: DashboardSummary Failed: %v", err)
		return nil, fmt.Errorf("ownerSrv.GetDashboardSummary: %w", err)
	}

	// 🛠️ Special Logic: CAPEX Section (Both Budget & Actual) ignores the 'year' filter
	// It should still filter by entities and departments to show project-wide scope.
	capexFilter := make(map[string]interface{})
	for k, v := range filter {
		capexFilter[k] = v
	}
	capexFilter["year"] = "All"

	capexSummary, err := s.capexSrv.GetCapexDashboardSummary(ctx, capexFilter)
	capexBudget := decimal.Zero
	capexActual := decimal.Zero
	if err == nil && capexSummary != nil {
		capexBudget = capexSummary.TotalBudget
		capexActual = capexSummary.TotalActual
	}

	logs.Infof("[DEBUG] OwnerService: Result - Budget=%v, Actual=%v, DeptsCount=%d", summary.TotalBudget, summary.TotalActual, len(summary.DepartmentData))

	return &models.OwnerDashboardSummaryDTO{
		DashboardSummaryDTO: *summary,
		CapexBudget:         capexBudget,
		CapexActual:         capexActual,
	}, nil
}

func (s *ownerService) GetActualTransactions(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	filter = s.InjectPermissions(ctx, user, filter)
	return s.repo.GetActualTransactions(ctx, filter)
}

func (s *ownerService) GetActualDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	filter = s.InjectPermissions(ctx, user, filter)
	return s.repo.GetActualDetails(ctx, filter)
}

func (s *ownerService) GetBudgetDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	filter = s.InjectPermissions(ctx, user, filter)
	return s.repo.GetBudgetDetails(ctx, filter)
}

func (s *ownerService) GetFilterOptions(ctx context.Context, user *models.UserInfo) (interface{}, error) {
	filter := s.InjectPermissions(ctx, user, nil)
	allowedDepts, _ := filter["departments"].([]string)

	allFacts, err := s.repo.GetBudgetFilterOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("ownerSrv.GetFilterOptions: %w", err)
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

func (s *ownerService) GetOrganizationStructure(ctx context.Context, user *models.UserInfo) ([]models.OrganizationDTO, error) {
	filter := s.InjectPermissions(ctx, user, nil)
	allowedDepts, _ := filter["departments"].([]string)

	facts, err := s.repo.GetOrganizationStructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("ownerSrv.GetOrganizationStructure: %w", err)
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

func (s *ownerService) GetOwnerFilterLists(ctx context.Context, user *models.UserInfo) (*models.OwnerFilterListsDTO, error) {
	years, err := s.repo.GetActualYears(ctx)
	if err != nil {
		return nil, fmt.Errorf("ownerSrv.GetOwnerFilterLists: %w", err)
	}
	return &models.OwnerFilterListsDTO{
		Companies: []string{"HMW", "ACG", "CLIK"},
		Branches:  []string{},
		Years:     years,
	}, nil
}

func (s *ownerService) GetActualYears(ctx context.Context, user *models.UserInfo) ([]string, error) {
	return s.repo.GetActualYears(ctx)
}

func (s *ownerService) InjectPermissions(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) map[string]interface{} {
	if filter == nil {
		filter = make(map[string]interface{})
	}

	// 🛡️ Proactively Fetch User Permissions if missing from context
	if len(user.Permissions) == 0 && user.ID != "" {
		dbPerms, err := s.authSrv.GetUserPermissions(ctx, user.ID)
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
	userNameLower := strings.ToLower(user.Username)
	if userNameLower == "admin" || userNameLower == "administrator" || strings.Contains(userNameLower, "admin") {
		isAdmin = true
	}

	// 🛠️ Logic Refinement: If you are an Admin, you are NEVER restricted by specific owner permissions.
	// This addresses the user's concern: only NON-ADMIN owners will be restricted by their permissions list.
	logs.Infof("[DEBUG] injectPermissions: User=%s, Roles=%v, Final isAdmin=%v, permsCount=%d", user.Username, roles, isAdmin, len(user.Permissions))

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
			logs.Warnf("[WARN] injectPermissions: User %s has no Department Permissions assigned!", user.Username)
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
						cTrim := strings.ToUpper(strings.TrimSpace(c))
						for _, a := range allowedDepts {
							aTrim := strings.ToUpper(strings.TrimSpace(a))
							// 🛡️ Case-Insensitive Bi-directional Matching: "acc" matches "ACC - Accounting"
							if cTrim == aTrim || strings.HasPrefix(cTrim, aTrim+" - ") || strings.HasPrefix(aTrim, cTrim+" - ") {
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
