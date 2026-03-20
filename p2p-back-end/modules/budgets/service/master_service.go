package service

import (
	"context"
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

func (s *masterDataService) ListGLMappings(ctx context.Context) ([]models.GlMappingEntity, error) {
	mappings, err := s.repo.ListGLMappings(ctx)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.ListGLMappings: %w", err)
	}

	sort.Slice(mappings, func(i, j int) bool {
		if mappings[i].Entity != mappings[j].Entity {
			return utils.NaturalLess(mappings[i].Entity, mappings[j].Entity)
		}
		return utils.NaturalLess(mappings[i].AccountName, mappings[j].AccountName)
	})

	return mappings, nil
}

func (s *masterDataService) GetGLMappingByID(ctx context.Context, id string) (*models.GlMappingEntity, error) {
	return s.repo.GetGLMappingByID(ctx, id)
}

func (s *masterDataService) CreateGLMapping(ctx context.Context, mapping *models.GlMappingEntity) error {
	mapping.ID = uuid.New()
	mapping.Entity = strings.ToUpper(strings.TrimSpace(mapping.Entity))
	return s.repo.CreateGLMapping(ctx, mapping)
}

func (s *masterDataService) UpdateGLMapping(ctx context.Context, mapping *models.GlMappingEntity) error {
	mapping.Entity = strings.ToUpper(strings.TrimSpace(mapping.Entity))
	return s.repo.UpdateGLMapping(ctx, mapping)
}

func (s *masterDataService) DeleteGLMapping(ctx context.Context, id string) error {
	return s.repo.DeleteGLMapping(ctx, id)
}

func (s *masterDataService) ImportGLMapping(ctx context.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, ".")+1:])
	if ext != "xlsx" {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: only .xlsx files are supported")
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: failed to read excel file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: excel file has no sheets")
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: failed to read rows: %v", err)
	}
	if len(rows) < 1 {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: excel file is empty")
	}

	header := rows[0]
	requiredHeaders := []string{"Entity", "Entity GL", "Conso GL", "Account Name"}
	if len(header) < len(requiredHeaders) {
		return fmt.Errorf("masterDataSrv.ImportGLMapping: invalid column count")
	}

	importCount := 0
	skipCount := 0

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			continue
		}

		padded := make([]string, 4)
		copy(padded, row)

		entity := strings.ToUpper(strings.TrimSpace(padded[0]))
		entityGL := strings.TrimSpace(padded[1])
		consoGL := strings.TrimSpace(padded[2])
		accountName := strings.TrimSpace(padded[3])

		if entity == "" || entityGL == "" || consoGL == "" {
			continue
		}

		exists, err := s.repo.CheckExactGLMapping(ctx, entity, entityGL, consoGL, accountName)
		if err == nil && exists {
			skipCount++
			continue
		}

		mapping := models.GlMappingEntity{
			ID:          uuid.New(),
			Entity:      entity,
			EntityGL:    entityGL,
			ConsoGL:     consoGL,
			AccountName: accountName,
			IsActive:    true,
		}

		if err := s.repo.CreateGLMapping(ctx, &mapping); err != nil {
			return fmt.Errorf("masterDataSrv.ImportGLMapping: failed row %d: %w", i+1, err)
		}
		importCount++
	}

	fmt.Printf("[Import GL Mapping] Imported: %d, Skipped: %d\n", importCount, skipCount)
	return nil
}

func (s *masterDataService) GetBudgetStructureTree(ctx context.Context) (interface{}, error) {
	entities, err := s.repo.GetBudgetStructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.GetBudgetStructureTree: %w", err)
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

func (s *masterDataService) ListBudgetStructure(ctx context.Context) ([]models.BudgetStructureEntity, error) {
	entities, err := s.repo.GetBudgetStructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.ListBudgetStructure: %w", err)
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	return entities, nil
}

func (s *masterDataService) GetBudgetStructureByID(ctx context.Context, id uint) (*models.BudgetStructureEntity, error) {
	return s.repo.GetBudgetStructureByID(ctx, id)
}

func (s *masterDataService) CreateBudgetStructure(ctx context.Context, entity *models.BudgetStructureEntity) error {
	return s.repo.CreateBudgetStructure(ctx, entity)
}

func (s *masterDataService) UpdateBudgetStructure(ctx context.Context, entity *models.BudgetStructureEntity) error {
	return s.repo.UpdateBudgetStructure(ctx, entity)
}

func (s *masterDataService) DeleteBudgetStructure(ctx context.Context, id uint) error {
	return s.repo.DeleteBudgetStructure(ctx, id)
}

func (s *masterDataService) GetUserConfigs(ctx context.Context, userID string) (map[string]string, error) {
	// Refactor: Make Data Management settings globally shared.
	// Ignore the requesting userID and use a fixed global identifier.
	globalID := "GLOBAL_ADMIN_SETTINGS"
	configs, err := s.repo.GetUserConfigs(ctx, globalID)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.GetUserConfigs: %w", err)
	}

	res := make(map[string]string)
	for _, c := range configs {
		res[c.ConfigKey] = c.Value
	}
	return res, nil
}

func (s *masterDataService) SetUserConfig(ctx context.Context, userID string, key string, value string) error {
	// Refactor: Make Data Management settings globally shared.
	// Ignore the requesting userID and use a fixed global identifier.
	globalID := "GLOBAL_ADMIN_SETTINGS"
	config := &models.UserConfigEntity{
		UserID:    globalID,
		ConfigKey: key,
		Value:     value,
	}
	return s.repo.UpdateUserConfig(ctx, config)
}


