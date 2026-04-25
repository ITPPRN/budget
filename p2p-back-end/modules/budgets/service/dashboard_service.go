package service

import (
	"context"
	"fmt"
	"sort"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type dashboardService struct {
	repo   models.DashboardRepository
	depSrv models.DepartmentService
}

func NewDashboardService(repo models.DashboardRepository, depSrv models.DepartmentService) models.DashboardService {
	return &dashboardService{repo: repo, depSrv: depSrv}
}

func (s *dashboardService) GetFilterOptions(ctx context.Context) ([]models.FilterOptionDTO, error) {
	facts, err := s.repo.GetBudgetFilterOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboardSrv.GetFilterOptions: %w", err)
	}
	fmt.Printf("[DEBUG] Service: Got %d facts\n", len(facts))

	// Nested Map Structure: Group -> Dept -> EntityGL -> []Leaf
	// We use a map for the first 3 levels, and a slice for the 4th (or map if needing unique)
	// Tree: map[Group] -> map[Dept] -> map[EntityGL] -> []{Code, Name}
	type Leaf struct {
		Code string
		Name string
	}
	tree := make(map[string]map[string]map[string][]Leaf)

	for _, f := range facts {
		groupName := f.Group
		if groupName == "" {
			groupName = "(No Group)"
		}
		deptName := f.Department
		if deptName == "" {
			deptName = "(No Dept)"
		}
		entityGLName := f.EntityGL
		if entityGLName == "" {
			entityGLName = "(No Category)"
		}

		// L4 Name: "Code - Name"
		code := f.ConsoGL
		name := f.GLName
		// If code is missing, maybe rely on name?

		if tree[groupName] == nil {
			tree[groupName] = make(map[string]map[string][]Leaf)
		}
		if tree[groupName][deptName] == nil {
			tree[groupName][deptName] = make(map[string][]Leaf)
		}

		// Add Leaf if unique. Since we query distinct, it should be mostly unique,
		// but multiple rows might share same code/name if other fields differ?
		// The repo query distincts on group,dept,entity_gl,conso_gl,gl_name. So it is unique per branch.

		// Check duplicates in slice? (Inefficient but fine for small N)
		leaves := tree[groupName][deptName][entityGLName]
		exists := false
		for _, l := range leaves {
			if l.Code == code && l.Name == name {
				exists = true
				break
			}
		}
		if !exists {
			tree[groupName][deptName][entityGLName] = append(leaves, Leaf{Code: code, Name: name})
		}
	}

	// Convert Map to DTO Slice
	var rootNodes []models.FilterOptionDTO

	for grpName, deptMap := range tree {
		grpNode := models.FilterOptionDTO{
			ID:       "L1_" + grpName,
			Name:     grpName,
			Level:    1,
			Children: []models.FilterOptionDTO{},
		}

		for deptName, glMap := range deptMap {
			deptNode := models.FilterOptionDTO{
				ID:       "L2_" + grpName + "_" + deptName, // Include parent name for absolute uniqueness
				Name:     deptName,
				Level:    2,
				Children: []models.FilterOptionDTO{},
			}

			for glName, leaves := range glMap {
				glNode := models.FilterOptionDTO{
					ID:       "L3_" + grpName + "_" + deptName + "_" + glName, // Include ancestors for absolute uniqueness
					Name:     glName,
					Level:    3,
					Children: []models.FilterOptionDTO{},
				}

				// Level 4 Leaves
				for _, leaf := range leaves {
					displayName := leaf.Name
					if leaf.Code != "" {
						displayName = fmt.Sprintf("%s-%s", leaf.Code, leaf.Name)
					}

					uniqueSuffix := "L4_" + glNode.ID + "_" + leaf.Name
					leafNode := models.FilterOptionDTO{
						ID:    fmt.Sprintf("%s|%s", leaf.Code, uniqueSuffix),
						Name:  displayName,
						Level: 4,
					}

					glNode.Children = append(glNode.Children, leafNode)
				}

				deptNode.Children = append(deptNode.Children, glNode)
			}
			grpNode.Children = append(grpNode.Children, deptNode)
		}
		rootNodes = append(rootNodes, grpNode)
	}

	// Recursive sort helper for DTO
	var sortDTO func([]models.FilterOptionDTO)
	sortDTO = func(nodes []models.FilterOptionDTO) {
		sort.Slice(nodes, func(i, j int) bool {
			return utils.NaturalLess(nodes[i].Name, nodes[j].Name)
		})
		for k, node := range nodes {
			if len(node.Children) > 0 {
				sortDTO(nodes[k].Children)
			}
		}
	}

	sortDTO(rootNodes)

	return rootNodes, nil
}
func (s *dashboardService) GetRawFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	return s.repo.GetBudgetFilterOptions(ctx)
}

func (s *dashboardService) GetOrganizationStructure(ctx context.Context) ([]models.OrganizationDTO, error) {
	facts, err := s.repo.GetOrganizationStructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboardSrv.GetOrganizationStructure: %w", err)
	}

	// Return Raw Codes (HMW, BUR) as stored in DB.
	// User requested "Abbreviations" in filter.

	// Map Entity -> Map Branch -> []Departments
	structure := make(map[string]map[string][]string)
	for _, f := range facts {
		if f.Entity == "" {
			continue
		}

		// Use Raw Entity Code
		entityName := f.Entity
		if structure[entityName] == nil {
			structure[entityName] = make(map[string][]string)
		}

		// Use Raw Branch Code
		if f.Branch != "" {
			branchName := f.Branch
			if _, exists := structure[entityName][branchName]; !exists {
				structure[entityName][branchName] = []string{}
			}

			// Add Department if not exists
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

	// Final Sort
	sort.Slice(result, func(i, j int) bool {
		return utils.NaturalLess(result[i].Entity, result[j].Entity)
	})
	for i := range result {
		sort.Slice(result[i].Branches, func(j, k int) bool {
			return utils.NaturalLess(result[i].Branches[j].Name, result[i].Branches[k].Name)
		})
		for j := range result[i].Branches {
			sort.Slice(result[i].Branches[j].Departments, func(k, l int) bool {
				return utils.NaturalLess(result[i].Branches[j].Departments[k], result[i].Branches[j].Departments[l])
			})
		}
	}

	return result, nil
}

func (s *dashboardService) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetDetailDTO, error) {
	sanitizeFilter(filter)
	return s.repo.GetBudgetDetails(ctx, filter)
}

func (s *dashboardService) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	sanitizeFilter(filter)
	return s.repo.GetActualDetails(ctx, filter)
}

func (s *dashboardService) GetDashboardSummary(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	sanitizeFilter(filter)
	return s.repo.GetDashboardAggregates(ctx, filter)
}

func (s *dashboardService) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	sanitizeFilter(filter)
	return s.repo.GetActualTransactions(ctx, filter)
}

func (s *dashboardService) GetActualYears(ctx context.Context) ([]string, error) {
	return s.repo.GetActualYears(ctx)
}

func (s *dashboardService) GetAvailableMonths(ctx context.Context, year string) ([]string, error) {
	return s.repo.GetAvailableMonths(ctx, year)
}

func (s *dashboardService) GetAdminPermittedMonths(ctx context.Context) []string {
	return s.repo.GetAdminPermittedMonths(ctx)
}
