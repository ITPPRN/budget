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

// --- Unified GL Grouping & Filter Tree ---

func (s *masterDataService) GetBudgetStructureTree(ctx context.Context) (interface{}, error) {
	groupings, err := s.repo.ListGLGroupings(ctx)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.GetBudgetStructureTree: %w", err)
	}

	type TreeNode struct {
		ID       string      `json:"id"`
		Name     string      `json:"name"`
		Level    int         `json:"level"`
		Children []*TreeNode `json:"children,omitempty"`
	}

	findChild := func(parent *TreeNode, name string) *TreeNode {
		for _, child := range parent.Children {
			if child.Name == name {
				return child
			}
		}
		return nil
	}

	var roots []*TreeNode
	processedLeaves := make(map[string]bool)

	for _, e := range groupings {
		if e.Group1 == "" || e.ConsoGL == "" {
			continue
		}

		leafKey := fmt.Sprintf("%s|%s|%s|%s", e.Group1, e.Group2, e.Group3, e.ConsoGL)
		if processedLeaves[leafKey] {
			continue
		}

		g1 := findChild(&TreeNode{Children: roots}, e.Group1)
		if g1 == nil {
			g1 = &TreeNode{ID: "G1|" + e.Group1, Name: e.Group1, Level: 1, Children: []*TreeNode{}}
			roots = append(roots, g1)
		}

		g2 := findChild(g1, e.Group2)
		if g2 == nil {
			g2 = &TreeNode{ID: g1.ID + "|G2|" + e.Group2, Name: e.Group2, Level: 2, Children: []*TreeNode{}}
			g1.Children = append(g1.Children, g2)
		}

		g3 := findChild(g2, e.Group3)
		if g3 == nil {
			g3 = &TreeNode{ID: g2.ID + "|G3|" + e.Group3, Name: e.Group3, Level: 3, Children: []*TreeNode{}}
			g2.Children = append(g2.Children, g3)
		}

		leafID := fmt.Sprintf("%s|%s", e.ConsoGL, e.ID)
		leafName := fmt.Sprintf("%s - %s", e.ConsoGL, e.AccountName)
		leaf := &TreeNode{ID: leafID, Name: leafName, Level: 4}
		g3.Children = append(g3.Children, leaf)

		processedLeaves[leafKey] = true
	}

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

func (s *masterDataService) ListGLGroupings(ctx context.Context) ([]models.GlGroupingEntity, error) {
	groupings, err := s.repo.ListGLGroupings(ctx)
	if err != nil {
		return nil, fmt.Errorf("masterDataSrv.ListGLGroupings: %w", err)
	}

	sort.Slice(groupings, func(i, j int) bool {
		if groupings[i].Entity != groupings[j].Entity {
			return utils.NaturalLess(groupings[i].Entity, groupings[j].Entity)
		}
		if groupings[i].Group1 != groupings[j].Group1 {
			return utils.NaturalLess(groupings[i].Group1, groupings[j].Group1)
		}
		return utils.NaturalLess(groupings[i].AccountName, groupings[j].AccountName)
	})

	return groupings, nil
}

func (s *masterDataService) GetGLGroupingByID(ctx context.Context, id string) (*models.GlGroupingEntity, error) {
	return s.repo.GetGLGroupingByID(ctx, id)
}

func (s *masterDataService) CreateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	g.ID = uuid.New()
	g.Entity = strings.ToUpper(strings.TrimSpace(g.Entity))
	return s.repo.CreateGLGrouping(ctx, g)
}

func (s *masterDataService) UpdateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	g.Entity = strings.ToUpper(strings.TrimSpace(g.Entity))
	return s.repo.UpdateGLGrouping(ctx, g)
}

func (s *masterDataService) DeleteGLGrouping(ctx context.Context, id string) error {
	return s.repo.DeleteGLGrouping(ctx, id)
}

func (s *masterDataService) ImportGLGrouping(ctx context.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: %w", err)
	}
	// defer file.Close()
	defer func() { _ = file.Close() }()
	ext := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, ".")+1:])
	if ext != "xlsx" {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: only .xlsx files are supported")
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: failed to read excel file: %v", err)
	}
	// defer f.Close()
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: excel file has no sheets")
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: failed to read rows: %v", err)
	}
	if len(rows) < 2 {
		return fmt.Errorf("masterDataSrv.ImportGLGrouping: excel file is too short")
	}

	importCount := 0
	updateCount := 0

	for i := 2; i < len(rows); i++ {
		row := rows[i]
		padded := make([]string, 7)
		copy(padded, row)

		g1 := strings.TrimSpace(padded[0])
		g2 := strings.TrimSpace(padded[1])
		g3 := strings.TrimSpace(padded[2])
		entity := strings.ToUpper(strings.TrimSpace(padded[3]))
		entityGL := strings.TrimSpace(padded[4])
		consoGL := strings.TrimSpace(padded[5])
		accName := strings.TrimSpace(padded[6])

		if entity == "" || entityGL == "" || consoGL == "" {
			continue
		}

		var existing models.GlGroupingEntity
		err := s.repo.GetGLGroupingInfo(ctx, entity, entityGL, &existing)
		if err == nil {
			existing.Group1 = g1
			existing.Group2 = g2
			existing.Group3 = g3
			existing.ConsoGL = consoGL
			existing.AccountName = accName
			_ =s.repo.UpdateGLGrouping(ctx, &existing)
			updateCount++
			continue
		}

		newEntry := models.GlGroupingEntity{
			ID:          uuid.New(),
			Entity:      entity,
			EntityGL:    entityGL,
			ConsoGL:     consoGL,
			AccountName: accName,
			Group1:      g1,
			Group2:      g2,
			Group3:      g3,
			IsActive:    true,
		}
		_ =s.repo.CreateGLGrouping(ctx, &newEntry)
		importCount++
	}

	fmt.Printf("[Import GL Grouping] Imported: %d, Updated: %d\n", importCount, updateCount)
	return nil
}

// --- Admin Settings (Unified) ---

func (s *masterDataService) GetUserConfigs(ctx context.Context, userID string) (map[string]string, error) {
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
	globalID := "GLOBAL_ADMIN_SETTINGS"
	config := &models.UserConfigEntity{
		UserID:    globalID,
		ConfigKey: key,
		Value:     value,
	}
	return s.repo.UpdateUserConfig(ctx, config)
}


