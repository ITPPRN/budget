package service

import (
	"fmt"
	"mime/multipart"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type masterDataService struct {
	repo models.MasterDataRepository
}

func NewMasterDataService(repo models.MasterDataRepository) models.MasterDataService {
	return &masterDataService{repo: repo}
}

func (s *masterDataService) ListGLMappings() ([]models.GlMappingEntity, error) {
	mappings, err := s.repo.ListGLMappings()
	if err != nil {
		return nil, err
	}

	sort.Slice(mappings, func(i, j int) bool {
		// Sort by Entity, then GL Name or Code? Let's use AccountName as primary
		if mappings[i].Entity != mappings[j].Entity {
			return utils.NaturalLess(mappings[i].Entity, mappings[j].Entity)
		}
		return utils.NaturalLess(mappings[i].AccountName, mappings[j].AccountName)
	})

	return mappings, nil
}

func (s *masterDataService) GetGLMappingByID(id string) (*models.GlMappingEntity, error) {
	return s.repo.GetGLMappingByID(id)
}

func (s *masterDataService) CreateGLMapping(mapping *models.GlMappingEntity) error {
	mapping.ID = uuid.New()
	mapping.Entity = strings.ToUpper(strings.TrimSpace(mapping.Entity))
	return s.repo.CreateGLMapping(mapping)
}

func (s *masterDataService) UpdateGLMapping(mapping *models.GlMappingEntity) error {
	mapping.Entity = strings.ToUpper(strings.TrimSpace(mapping.Entity))
	return s.repo.UpdateGLMapping(mapping)
}

func (s *masterDataService) DeleteGLMapping(id string) error {
	return s.repo.DeleteGLMapping(id)
}

func (s *masterDataService) ImportGLMapping(fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	ext := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, ".")+1:])
	if ext != "xlsx" {
		return fmt.Errorf("only .xlsx files are supported")
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("failed to read excel file: %v", err)
	}
	defer f.Close()

	// 1. Use the first visible sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("excel file has no sheets")
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read rows from sheet '%s': %v", sheetName, err)
	}

	if len(rows) < 1 {
		return fmt.Errorf("excel file is empty")
	}

	// 2. Validate Header (Row 1)
	header := rows[0]
	requiredHeaders := []string{"Entity", "Entity GL", "Conso GL", "Account Name"}
	if len(header) != len(requiredHeaders) {
		return fmt.Errorf("invalid column count: expected %d columns (Entity, Entity GL, Conso GL, Account Name)", len(requiredHeaders))
	}

	for i, h := range requiredHeaders {
		if strings.TrimSpace(header[i]) != h {
			return fmt.Errorf("invalid column at position %d: expected '%s', got '%s'", i+1, h, header[i])
		}
	}

	// 3. Process Data Rows (starting from index 1)
	importCount := 0
	skipCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		// Pad row to ensure we don't crash if Account Name is missing in raw row slice
		padded := make([]string, 4)
		copy(padded, row)

		entity := strings.ToUpper(strings.TrimSpace(padded[0]))
		entityGL := strings.TrimSpace(padded[1])
		consoGL := strings.TrimSpace(padded[2])
		accountName := strings.TrimSpace(padded[3])

		if entity == "" || entityGL == "" || consoGL == "" {
			continue // Skip incomplete major fields
		}

		// 4. Exact duplicate check (Check all 4 fields)
		exists, err := s.repo.CheckExactGLMapping(entity, entityGL, consoGL, accountName)
		if err == nil && exists {
			skipCount++
			continue // Skip perfect duplicate
		}

		mapping := models.GlMappingEntity{
			ID:          uuid.New(),
			Entity:      entity,
			EntityGL:    entityGL,
			ConsoGL:     consoGL,
			AccountName: accountName,
			IsActive:    true,
		}

		if err := s.repo.CreateGLMapping(&mapping); err != nil {
			return fmt.Errorf("failed to create mapping at row %d: %v", i+1, err)
		}
		importCount++
	}

	fmt.Printf("[Import GL Mapping] Imported: %d, Skipped (Duplicate): %d\n", importCount, skipCount)
	return nil
}

func (s *masterDataService) GetBudgetStructureTree() (interface{}, error) {
	entities, err := s.repo.GetBudgetStructure()
	if err != nil {
		return nil, err
	}

	// Build Tree from flat data (Group1 -> Group2 -> Group3 -> Leaf(ConsoGL))
	type TreeNode struct {
		ID       string      `json:"id"`
		Name     string      `json:"name"`
		Level    int         `json:"level"`
		Children []*TreeNode `json:"children,omitempty"`
	}

	// Helper to find child by name to prevent duplicates
	findChild := func(parent *TreeNode, name string) *TreeNode {
		for _, child := range parent.Children {
			if child.Name == name {
				return child
			}
		}
		return nil
	}

	var roots []*TreeNode

	for _, e := range entities {
		// Level 1: Group 1
		g1 := findChild(&TreeNode{Children: roots}, e.Group1)
		if g1 == nil {
			g1 = &TreeNode{ID: "G1|" + e.Group1, Name: e.Group1, Level: 1, Children: []*TreeNode{}}
			roots = append(roots, g1)
		}

		// Level 2: Group 2
		g2 := findChild(g1, e.Group2)
		if g2 == nil {
			// ID includes parent ID for uniqueness
			g2 = &TreeNode{ID: g1.ID + "|G2|" + e.Group2, Name: e.Group2, Level: 2, Children: []*TreeNode{}}
			g1.Children = append(g1.Children, g2)
		}

		// Level 3: Group 3
		g3 := findChild(g2, e.Group3)
		if g3 == nil {
			// ID includes parent path for uniqueness
			g3 = &TreeNode{ID: g2.ID + "|G3|" + e.Group3, Name: e.Group3, Level: 3, Children: []*TreeNode{}}
			g2.Children = append(g2.Children, g3)
		}

		// Leaf: ConsoGL + Account Name
		leafID := fmt.Sprintf("%s|%d", e.ConsoGL, e.ID)
		leafName := fmt.Sprintf("%s - %s", e.ConsoGL, e.AccountName)
		leaf := &TreeNode{ID: leafID, Name: leafName, Level: 4}
		g3.Children = append(g3.Children, leaf)
	}

	// Recursive sort helper
	var sortTree func([]*TreeNode)
	sortTree = func(nodes []*TreeNode) {
		sort.Slice(nodes, func(i, j int) bool {
			return utils.NaturalLess(nodes[i].Name, nodes[j].Name)
		})
		for _, node := range nodes {
			if len(node.Children) > 0 {
				sortTree(node.Children)
			}
		}
	}

	sortTree(roots)

	return roots, nil
}

func (s *masterDataService) ListBudgetStructure() ([]models.BudgetStructureEntity, error) {
	entities, err := s.repo.GetBudgetStructure()
	if err != nil {
		return nil, err
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	return entities, nil
}

func (s *masterDataService) GetBudgetStructureByID(id uint) (*models.BudgetStructureEntity, error) {
	return s.repo.GetBudgetStructureByID(id)
}

func (s *masterDataService) CreateBudgetStructure(entity *models.BudgetStructureEntity) error {
	return s.repo.CreateBudgetStructure(entity)
}

func (s *masterDataService) UpdateBudgetStructure(entity *models.BudgetStructureEntity) error {
	return s.repo.UpdateBudgetStructure(entity)
}

func (s *masterDataService) DeleteBudgetStructure(id uint) error {
	return s.repo.DeleteBudgetStructure(id)
}

func (s *masterDataService) GetUserConfigs(userID string) (map[string]string, error) {
	configs, err := s.repo.GetUserConfigs(userID)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	for _, c := range configs {
		res[c.ConfigKey] = c.Value
	}
	return res, nil
}

func (s *masterDataService) SetUserConfig(userID string, key string, value string) error {
	config := &models.UserConfigEntity{
		UserID:    userID,
		ConfigKey: key,
		Value:     value,
	}
	return s.repo.UpdateUserConfig(config)
}


