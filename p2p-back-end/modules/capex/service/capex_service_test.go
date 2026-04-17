package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"

	"p2p-back-end/modules/entities/models"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type MockCapexRepository struct {
	mock.Mock
}

func (m *MockCapexRepository) WithTrx(trxHandle func(repo models.CapexRepository) error) error {
	m.Called(trxHandle)
	return trxHandle(m)
}

func (m *MockCapexRepository) CreateFileCapexBudget(ctx context.Context, file *models.FileCapexBudgetEntity) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockCapexRepository) CreateFileCapexActual(ctx context.Context, file *models.FileCapexActualEntity) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockCapexRepository) CreateCapexBudgetFacts(ctx context.Context, facts []models.CapexBudgetFactEntity) error {
	args := m.Called(ctx, facts)
	return args.Error(0)
}

func (m *MockCapexRepository) CreateCapexActualFacts(ctx context.Context, facts []models.CapexActualFactEntity) error {
	args := m.Called(ctx, facts)
	return args.Error(0)
}

func (m *MockCapexRepository) ListFileCapexBudgets(ctx context.Context) ([]models.FileCapexBudgetEntity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.FileCapexBudgetEntity), args.Error(1)
}

func (m *MockCapexRepository) ListFileCapexActuals(ctx context.Context) ([]models.FileCapexActualEntity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.FileCapexActualEntity), args.Error(1)
}

func (m *MockCapexRepository) GetFileCapexBudget(ctx context.Context, id string) (*models.FileCapexBudgetEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileCapexBudgetEntity), args.Error(1)
}

func (m *MockCapexRepository) GetFileCapexActual(ctx context.Context, id string) (*models.FileCapexActualEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileCapexActualEntity), args.Error(1)
}

func (m *MockCapexRepository) DeleteFileCapexBudget(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexRepository) DeleteFileCapexActual(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexRepository) DeleteAllCapexBudgetFacts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapexRepository) DeleteAllCapexActualFacts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapexRepository) DeleteCapexBudgetFactsByFileID(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockCapexRepository) DeleteCapexActualFactsByFileID(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockCapexRepository) UpdateFileCapexBudget(ctx context.Context, id string, filename string) error {
	args := m.Called(ctx, id, filename)
	return args.Error(0)
}

func (m *MockCapexRepository) UpdateFileCapexActual(ctx context.Context, id string, filename string) error {
	args := m.Called(ctx, id, filename)
	return args.Error(0)
}

func (m *MockCapexRepository) GetCapexDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardSummaryDTO), args.Error(1)
}

// ---------------------------------------------------------------------------
// Header helpers (new format with Branch)
// ---------------------------------------------------------------------------

var testHeaders = []string{"Entity", "Branch", "Department", "CAPEX No.", "CAPEX Name", "CAPEX Category", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC", "YEARTOTAL"}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// 1. Constructor
func TestNewCapexService(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	assert.NotNil(t, svc)
}

// 2-3. ListCapexBudgetFiles
func TestListCapexBudgetFiles_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	expected := []models.FileCapexBudgetEntity{
		{ID: uuid.New(), FileName: "budget_2025.xlsx", Year: "2025", UploadAt: time.Now()},
	}
	repo.On("ListFileCapexBudgets", ctx).Return(expected, nil)

	result, err := svc.ListCapexBudgetFiles(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expected[0].FileName, result[0].FileName)
	repo.AssertExpectations(t)
}

func TestListCapexBudgetFiles_Error(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	repo.On("ListFileCapexBudgets", ctx).Return([]models.FileCapexBudgetEntity{}, errors.New("db error"))

	result, err := svc.ListCapexBudgetFiles(ctx)
	assert.Error(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

// 4. ListCapexActualFiles
func TestListCapexActualFiles_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	expected := []models.FileCapexActualEntity{
		{ID: uuid.New(), FileName: "actual_2025.xlsx", Year: "2025", UploadAt: time.Now()},
	}
	repo.On("ListFileCapexActuals", ctx).Return(expected, nil)

	result, err := svc.ListCapexActualFiles(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expected[0].FileName, result[0].FileName)
	repo.AssertExpectations(t)
}

// 5-6. DeleteCapexBudgetFile
func TestDeleteCapexBudgetFile_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	id := uuid.New().String()

	repo.On("DeleteFileCapexBudget", ctx, id).Return(nil)

	err := svc.DeleteCapexBudgetFile(ctx, id)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteCapexBudgetFile_Error(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	id := uuid.New().String()

	repo.On("DeleteFileCapexBudget", ctx, id).Return(errors.New("not found"))

	err := svc.DeleteCapexBudgetFile(ctx, id)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// 7. DeleteCapexActualFile
func TestDeleteCapexActualFile_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	id := uuid.New().String()

	repo.On("DeleteFileCapexActual", ctx, id).Return(nil)

	err := svc.DeleteCapexActualFile(ctx, id)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 8. RenameCapexBudgetFile
func TestRenameCapexBudgetFile_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	id := uuid.New().String()

	repo.On("UpdateFileCapexBudget", ctx, id, "new_budget_name.xlsx").Return(nil)

	err := svc.RenameCapexBudgetFile(ctx, id, "new_budget_name.xlsx")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 9. RenameCapexActualFile
func TestRenameCapexActualFile_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	id := uuid.New().String()

	repo.On("UpdateFileCapexActual", ctx, id, "new_actual_name.xlsx").Return(nil)

	err := svc.RenameCapexActualFile(ctx, id, "new_actual_name.xlsx")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 10-11. GetCapexDashboardSummary
func TestGetCapexDashboardSummary_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	filter := map[string]interface{}{"year": "2025"}

	expected := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(100000),
		TotalActual: decimal.NewFromInt(50000),
	}
	repo.On("GetCapexDashboardAggregates", ctx, filter).Return(expected, nil)

	result, err := svc.GetCapexDashboardSummary(ctx, filter)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, expected.TotalBudget.Equal(result.TotalBudget))
	assert.True(t, expected.TotalActual.Equal(result.TotalActual))
	repo.AssertExpectations(t)
}

func TestGetCapexDashboardSummary_Error(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	filter := map[string]interface{}{"year": "2025"}

	repo.On("GetCapexDashboardAggregates", ctx, filter).Return(nil, errors.New("aggregate error"))

	result, err := svc.GetCapexDashboardSummary(ctx, filter)
	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

// 12. ClearCapexBudget
func TestClearCapexBudget_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	repo.On("WithTrx", mock.AnythingOfType("func(models.CapexRepository) error")).Return(nil)
	repo.On("DeleteAllCapexBudgetFacts", ctx).Return(nil)

	err := svc.ClearCapexBudget(ctx)
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteAllCapexBudgetFacts", ctx)
}

// 13. ClearCapexActual
func TestClearCapexActual_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	repo.On("WithTrx", mock.AnythingOfType("func(models.CapexRepository) error")).Return(nil)
	repo.On("DeleteAllCapexActualFacts", ctx).Return(nil)

	err := svc.ClearCapexActual(ctx)
	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteAllCapexActualFacts", ctx)
}

// 14. SyncCapexBudget_Success (with Branch)
func TestSyncCapexBudget_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	fileID := uuid.New()
	fileIDStr := fileID.String()

	jsonData, _ := json.Marshal([][]string{
		testHeaders,
		{"ACG", "HQ", "IT", "CX001", "Server", "Hardware", "100", "200", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "300"},
	})

	fileEntity := &models.FileCapexBudgetEntity{
		ID:       fileID,
		FileName: "capex_budget_2025.xlsx",
		Year:     "2025",
		UploadAt: time.Now(),
		Data:     datatypes.JSON(jsonData),
	}

	repo.On("GetFileCapexBudget", ctx, fileIDStr).Return(fileEntity, nil)
	repo.On("WithTrx", mock.AnythingOfType("func(models.CapexRepository) error")).Return(nil)
	repo.On("DeleteAllCapexBudgetFacts", ctx).Return(nil)
	repo.On("CreateCapexBudgetFacts", ctx, mock.MatchedBy(func(facts []models.CapexBudgetFactEntity) bool {
		if len(facts) != 1 {
			return false
		}
		f := facts[0]
		return f.Entity == "ACG" &&
			f.Branch == "HQ" &&
			f.Department == "IT" &&
			f.CapexNo == "CX001" &&
			f.CapexName == "Server" &&
			f.CapexCategory == "Hardware" &&
			f.Year == "2025" &&
			f.YearTotal.Equal(decimal.NewFromInt(300)) &&
			len(f.CapexBudgetAmounts) == 12
	})).Return(nil)

	err := svc.SyncCapexBudget(ctx, fileIDStr)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 15. SyncCapexBudget_FileNotFound
func TestSyncCapexBudget_FileNotFound(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	fileID := uuid.New().String()

	repo.On("GetFileCapexBudget", ctx, fileID).Return(nil, errors.New("record not found"))

	err := svc.SyncCapexBudget(ctx, fileID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FetchFile")
	repo.AssertExpectations(t)
}

// 16. SyncCapexBudget_InvalidJSON
func TestSyncCapexBudget_InvalidJSON(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	fileID := uuid.New()
	fileIDStr := fileID.String()

	fileEntity := &models.FileCapexBudgetEntity{
		ID:       fileID,
		FileName: "bad_2025.xlsx",
		Year:     "2025",
		UploadAt: time.Now(),
		Data:     datatypes.JSON([]byte(`not valid json`)),
	}

	repo.On("GetFileCapexBudget", ctx, fileIDStr).Return(fileEntity, nil)

	err := svc.SyncCapexBudget(ctx, fileIDStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unmarshal")
	repo.AssertExpectations(t)
}

// 17. SyncCapexActual_Success (with Branch)
func TestSyncCapexActual_Success(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	fileID := uuid.New()
	fileIDStr := fileID.String()

	jsonData, _ := json.Marshal([][]string{
		testHeaders,
		{"ACG", "BKK", "Finance", "CX002", "Laptop", "Equipment", "50", "50", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "100"},
	})

	fileEntity := &models.FileCapexActualEntity{
		ID:       fileID,
		FileName: "capex_actual_2025.xlsx",
		Year:     "2025",
		UploadAt: time.Now(),
		Data:     datatypes.JSON(jsonData),
	}

	repo.On("GetFileCapexActual", ctx, fileIDStr).Return(fileEntity, nil)
	repo.On("WithTrx", mock.AnythingOfType("func(models.CapexRepository) error")).Return(nil)
	repo.On("DeleteAllCapexActualFacts", ctx).Return(nil)
	repo.On("CreateCapexActualFacts", ctx, mock.MatchedBy(func(facts []models.CapexActualFactEntity) bool {
		if len(facts) != 1 {
			return false
		}
		f := facts[0]
		return f.Entity == "ACG" &&
			f.Branch == "BKK" &&
			f.Department == "Finance" &&
			f.CapexNo == "CX002" &&
			f.CapexName == "Laptop" &&
			f.CapexCategory == "Equipment" &&
			f.Year == "2025" &&
			f.YearTotal.Equal(decimal.NewFromInt(100)) &&
			len(f.CapexActualAmounts) == 12
	})).Return(nil)

	err := svc.SyncCapexActual(ctx, fileIDStr)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 18. SyncCapexActual_FileNotFound
func TestSyncCapexActual_FileNotFound(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()
	fileID := uuid.New().String()

	repo.On("GetFileCapexActual", ctx, fileID).Return(nil, errors.New("record not found"))

	err := svc.SyncCapexActual(ctx, fileID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FetchFile")
	repo.AssertExpectations(t)
}

// 19. processCapexBudgetFact with empty/insufficient rows
func TestProcessCapexBudgetFact_EmptyRows(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo).(*capexService)

	fileID := uuid.New()

	// No rows at all
	result, err := svc.processCapexBudgetFact([][]string{}, fileID, "2025")
	assert.NoError(t, err)
	assert.Empty(t, result)

	// Only header row (less than 2 rows)
	result, err = svc.processCapexBudgetFact([][]string{testHeaders}, fileID, "2025")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// 20. processCapexBudgetFact with data rows (with Branch)
func TestProcessCapexBudgetFact_WithData(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo).(*capexService)

	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACG", "HQ", "IT", "CX001", "Server", "Hardware", "100", "200", "300", "0", "0", "0", "0", "0", "0", "0", "0", "0", "600"},
		{"ACG", "BKK", "HR", "CX002", "Chairs", "Furniture", "50", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "50"},
	}

	result, err := svc.processCapexBudgetFact(rows, fileID, "2025")
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// First row
	assert.Equal(t, "ACG", result[0].Entity)
	assert.Equal(t, "HQ", result[0].Branch)
	assert.Equal(t, "IT", result[0].Department)
	assert.Equal(t, "CX001", result[0].CapexNo)
	assert.Equal(t, "Server", result[0].CapexName)
	assert.Equal(t, "Hardware", result[0].CapexCategory)
	assert.Equal(t, "2025", result[0].Year)
	assert.Equal(t, fileID, result[0].FileCapexBudgetID)
	assert.Len(t, result[0].CapexBudgetAmounts, 12)
	assert.True(t, result[0].YearTotal.Equal(decimal.NewFromInt(600)))
	assert.True(t, result[0].CapexBudgetAmounts[0].Amount.Equal(decimal.NewFromInt(100))) // JAN
	assert.True(t, result[0].CapexBudgetAmounts[1].Amount.Equal(decimal.NewFromInt(200))) // FEB
	assert.True(t, result[0].CapexBudgetAmounts[2].Amount.Equal(decimal.NewFromInt(300))) // MAR
	assert.Equal(t, "JAN", result[0].CapexBudgetAmounts[0].Month)
	assert.Equal(t, "FEB", result[0].CapexBudgetAmounts[1].Month)
	assert.Equal(t, "DEC", result[0].CapexBudgetAmounts[11].Month)

	// Second row
	assert.Equal(t, "BKK", result[1].Branch)
	assert.Equal(t, "HR", result[1].Department)
	assert.Equal(t, "CX002", result[1].CapexNo)
	assert.True(t, result[1].YearTotal.Equal(decimal.NewFromInt(50)))
}

// 21. processCapexActualFact with empty rows
func TestProcessCapexActualFact_EmptyRows(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo).(*capexService)

	fileID := uuid.New()

	result, err := svc.processCapexActualFact([][]string{}, fileID, "2025")
	assert.NoError(t, err)
	assert.Empty(t, result)

	result, err = svc.processCapexActualFact([][]string{testHeaders}, fileID, "2025")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// 22. processCapexActualFact with data rows (with Branch)
func TestProcessCapexActualFact_WithData(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo).(*capexService)

	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACG", "CNX", "Finance", "CX010", "Laptop", "Equipment", "1000", "500", "0", "0", "0", "0", "0", "0", "0", "0", "0", "250", "1750"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2025")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	fact := result[0]
	assert.Equal(t, "ACG", fact.Entity)
	assert.Equal(t, "CNX", fact.Branch)
	assert.Equal(t, "Finance", fact.Department)
	assert.Equal(t, "CX010", fact.CapexNo)
	assert.Equal(t, "Laptop", fact.CapexName)
	assert.Equal(t, "Equipment", fact.CapexCategory)
	assert.Equal(t, "2025", fact.Year)
	assert.Equal(t, fileID, fact.FileCapexActualID)
	assert.Len(t, fact.CapexActualAmounts, 12)
	assert.True(t, fact.YearTotal.Equal(decimal.NewFromInt(1750)))
	assert.True(t, fact.CapexActualAmounts[0].Amount.Equal(decimal.NewFromInt(1000)))  // JAN
	assert.True(t, fact.CapexActualAmounts[1].Amount.Equal(decimal.NewFromInt(500)))   // FEB
	assert.True(t, fact.CapexActualAmounts[11].Amount.Equal(decimal.NewFromInt(250)))  // DEC
	assert.Equal(t, "JAN", fact.CapexActualAmounts[0].Month)
	assert.Equal(t, "DEC", fact.CapexActualAmounts[11].Month)
}

// ---------------------------------------------------------------------------
// NEW: Branch-specific tests
// ---------------------------------------------------------------------------

// 23. Branch is correctly extracted from different positions
func TestProcessCapexActualFact_BranchExtraction(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server", "IT Equipment", "100", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "100"},
		{"ACH", "BKK", "HR", "CAP-002", "Office", "Building", "200", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "200"},
		{"PVL", "CNX", "MKT", "CAP-003", "Platform", "Software", "300", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "300"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	assert.Equal(t, "HQ", result[0].Branch)
	assert.Equal(t, "IT", result[0].Department)

	assert.Equal(t, "BKK", result[1].Branch)
	assert.Equal(t, "HR", result[1].Department)

	assert.Equal(t, "CNX", result[2].Branch)
	assert.Equal(t, "MKT", result[2].Department)
}

// 24. Branch can be empty string (backward compatibility)
func TestProcessCapexActualFact_EmptyBranch(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "", "IT", "CAP-001", "Server", "IT Equipment", "100", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "100"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "", result[0].Branch)
	assert.Equal(t, "IT", result[0].Department)
	assert.Equal(t, "ACH", result[0].Entity)
}

// 25. Budget also extracts Branch correctly
func TestProcessCapexBudgetFact_BranchExtraction(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server", "IT Equipment", "500", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "500"},
		{"PVL", "BKK", "MKT", "CAP-002", "Marketing", "Software", "0", "300", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "300"},
	}

	result, err := svc.processCapexBudgetFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, "ACH", result[0].Entity)
	assert.Equal(t, "HQ", result[0].Branch)
	assert.Equal(t, "IT", result[0].Department)
	assert.True(t, result[0].YearTotal.Equal(decimal.NewFromInt(500)))

	assert.Equal(t, "PVL", result[1].Entity)
	assert.Equal(t, "BKK", result[1].Branch)
	assert.Equal(t, "MKT", result[1].Department)
	assert.True(t, result[1].YearTotal.Equal(decimal.NewFromInt(300)))
}

// 26. Multiple rows same branch - amounts calculated independently
func TestProcessCapexActualFact_MultipleSameBranch(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server", "Hardware", "1000", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "1000"},
		{"ACH", "HQ", "IT", "CAP-002", "Laptop", "Hardware", "500", "500", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "1000"},
		{"ACH", "BKK", "HR", "CAP-003", "Desk", "Furniture", "200", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "200"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Both HQ rows should have Branch = "HQ"
	assert.Equal(t, "HQ", result[0].Branch)
	assert.Equal(t, "HQ", result[1].Branch)
	assert.Equal(t, "BKK", result[2].Branch)

	// Each row has its own YearTotal
	assert.True(t, result[0].YearTotal.Equal(decimal.NewFromInt(1000)))
	assert.True(t, result[1].YearTotal.Equal(decimal.NewFromInt(1000)))
	assert.True(t, result[2].YearTotal.Equal(decimal.NewFromInt(200)))
}

// 27. SyncCapexActual end-to-end with Branch
func TestSyncCapexActual_WithBranch_EndToEnd(t *testing.T) {
	repo := new(MockCapexRepository)
	svc := NewCapexService(repo)
	ctx := context.Background()

	fileID := uuid.New()
	fileIDStr := fileID.String()

	jsonData, _ := json.Marshal([][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server Upgrade", "IT Equipment", "150000", "0", "200000", "0", "0", "50000", "0", "0", "100000", "0", "0", "0", "500000"},
		{"ACH", "BKK", "HR", "CAP-002", "Office Renovation", "Building", "0", "300000", "300000", "250000", "0", "0", "0", "0", "0", "0", "0", "0", "850000"},
	})

	fileEntity := &models.FileCapexActualEntity{
		ID:       fileID,
		FileName: "capex_actual_2026.xlsx",
		Year:     "2026",
		UploadAt: time.Now(),
		Data:     datatypes.JSON(jsonData),
	}

	repo.On("GetFileCapexActual", ctx, fileIDStr).Return(fileEntity, nil)
	repo.On("WithTrx", mock.AnythingOfType("func(models.CapexRepository) error")).Return(nil)
	repo.On("DeleteAllCapexActualFacts", ctx).Return(nil)
	repo.On("CreateCapexActualFacts", ctx, mock.MatchedBy(func(facts []models.CapexActualFactEntity) bool {
		if len(facts) != 2 {
			return false
		}

		// Row 1: ACH / HQ / IT
		f1 := facts[0]
		if f1.Entity != "ACH" || f1.Branch != "HQ" || f1.Department != "IT" {
			return false
		}
		if f1.CapexNo != "CAP-001" || f1.CapexName != "Server Upgrade" {
			return false
		}
		if !f1.YearTotal.Equal(decimal.NewFromInt(500000)) {
			return false
		}

		// Row 2: ACH / BKK / HR
		f2 := facts[1]
		if f2.Entity != "ACH" || f2.Branch != "BKK" || f2.Department != "HR" {
			return false
		}
		if !f2.YearTotal.Equal(decimal.NewFromInt(850000)) {
			return false
		}

		return true
	})).Return(nil)

	err := svc.SyncCapexActual(ctx, fileIDStr)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// 28. Verify comma-formatted amounts parse correctly
func TestProcessCapexActualFact_CommaAmounts(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server", "HW", "1,000", "2,500.50", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "3500.50"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	fact := result[0]
	assert.Equal(t, "HQ", fact.Branch)
	assert.True(t, fact.CapexActualAmounts[0].Amount.Equal(decimal.NewFromInt(1000)))                            // JAN: "1,000"
	assert.True(t, fact.CapexActualAmounts[1].Amount.Equal(decimal.RequireFromString("2500.50")))                // FEB: "2,500.50"
	assert.True(t, fact.YearTotal.Equal(decimal.RequireFromString("3500.50")))
}

// 29. Verify dash and empty values default to zero
func TestProcessCapexActualFact_DashAndEmpty(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT", "CAP-001", "Server", "HW", "-", "", "100", "0", "0", "0", "0", "0", "0", "0", "0", "0", "100"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	fact := result[0]
	assert.True(t, fact.CapexActualAmounts[0].Amount.Equal(decimal.Zero)) // JAN: "-" -> 0
	assert.True(t, fact.CapexActualAmounts[1].Amount.Equal(decimal.Zero)) // FEB: "" -> 0
	assert.True(t, fact.CapexActualAmounts[2].Amount.Equal(decimal.NewFromInt(100))) // MAR: "100"
	assert.True(t, fact.YearTotal.Equal(decimal.NewFromInt(100)))
}

// 30. Verify extractYear helper
func TestExtractYear(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"CAPEX_Actual_2026.xlsx", "2026"},
		{"budget_2025_v2.xlsx", "2025"},
		{"no_year.xlsx", ""}, // will fallback to current year
		{"2024_capex.xlsx", "2024"},
	}

	for _, tt := range tests {
		result := extractYear(tt.filename)
		if tt.expected != "" {
			assert.Equal(t, tt.expected, result, "filename: %s", tt.filename)
		}
	}
}

// 31. Case-insensitive header matching
func TestProcessCapexActualFact_CaseInsensitiveHeaders(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		{"entity", "branch", "department", "capex no.", "capex name", "capex category", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec", "yeartotal"},
		{"ACH", "HQ", "IT", "CAP-001", "Test", "HW", "100", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "100"},
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "ACH", result[0].Entity)
	assert.Equal(t, "HQ", result[0].Branch)
	assert.Equal(t, "IT", result[0].Department)
}

// 32. Short row (missing columns) should not panic
func TestProcessCapexActualFact_ShortRow(t *testing.T) {
	svc := NewCapexService(new(MockCapexRepository)).(*capexService)
	fileID := uuid.New()

	rows := [][]string{
		testHeaders,
		{"ACH", "HQ", "IT"}, // only 3 columns, rest missing
	}

	result, err := svc.processCapexActualFact(rows, fileID, "2026")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	fact := result[0]
	assert.Equal(t, "ACH", fact.Entity)
	assert.Equal(t, "HQ", fact.Branch)
	assert.Equal(t, "IT", fact.Department)
	assert.Equal(t, "", fact.CapexNo)   // missing -> empty
	assert.Equal(t, "", fact.CapexName) // missing -> empty
	assert.True(t, fact.YearTotal.Equal(decimal.Zero)) // all amounts zero
}
