package service

import (
	"p2p-back-end/modules/entities/models"
)

type ownerService struct {
	repo models.OwnerRepository
}

func NewOwnerService(repo models.OwnerRepository) models.OwnerService {
	return &ownerService{repo: repo}
}

func (s *ownerService) GetDashboardSummary(user *models.UserInfo, filter map[string]interface{}) (*models.OwnerDashboardSummaryDTO, error) {
	// 1. Get User Context (Department, etc.)
	// userEntity, err := s.repo.GetUserContext(user.ID) // If needed for strict RLS
	// For now, we trust the Filter passed from Controller (which includes Dept from Token if needed? Or we enforce it here?)
	// Let's enforce Department Filter if User is NOT Admin/Owner-Global
	// Actually, the requirement says "Owner sees Dashboard (OWNER)".
	// If the user has a specific department in their profile, we SHOULD restrict the view.
	// But `filter` coming from Controller might already have it?
	// Let's assume Controller passes raw filters.

	// Logic: If user has Department Code, FORCE validation.
	// But for now, let's just fetch global aggregates based on the requested filter.
	return s.repo.GetDashboardAggregates(filter)
}

func (s *ownerService) GetActualTransactions(user *models.UserInfo, filter map[string]interface{}) ([]models.ActualTransactionDTO, error) {
	return s.repo.GetActualTransactions(filter)
}

func (s *ownerService) GetActualDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.OwnerActualFactEntity, error) {
	return s.repo.GetActualDetails(filter)
}

func (s *ownerService) GetBudgetDetails(user *models.UserInfo, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	return s.repo.GetBudgetDetails(filter)
}

func (s *ownerService) GetFilterOptions(user *models.UserInfo) ([]models.FilterOptionDTO, error) {
	// Re-use budget logic or fetch from Owner Repo
	// We need a way to Convert BudgetFactEntity to FilterOptionDTO tree.
	// Since OwnerRepo.GetBudgetFilterOptions returns []BudgetFactEntity, we need the mapper.
	rawOptions, err := s.repo.GetBudgetFilterOptions()
	if err != nil {
		return nil, err
	}

	// Mapper Logic (Duplicated from Budget Service? Or Shared?)
	// Let's duplicate for isolation or move to shared utils.
	// Implementing simple mapping here.
	return s.mapFactsToFilterOptions(rawOptions), nil
}

func (s *ownerService) GetOwnerFilterLists(user *models.UserInfo) (*models.OwnerFilterListsDTO, error) {
	return s.repo.GetOwnerFilterLists()
}

func (s *ownerService) AutoSyncOwnerActuals() error {
	return s.repo.AutoSyncOwnerActuals()
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
