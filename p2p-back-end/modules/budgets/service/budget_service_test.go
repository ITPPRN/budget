package service

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

// ============================================================
// MOCK DEFINITIONS
// ============================================================

// --- MockActualRepository ---

type MockActualRepository struct {
	mock.Mock
}

func (m *MockActualRepository) WithTrx(trxHandle func(repo models.ActualRepository) error) error {
	m.Called(trxHandle)
	return trxHandle(m)
}

func (m *MockActualRepository) CreateActualFacts(ctx context.Context, facts []models.ActualFactEntity) error {
	args := m.Called(ctx, facts)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteAllActualFacts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteActualFactsByYear(ctx context.Context, year string) error {
	args := m.Called(ctx, year)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteActualFactsByMonth(ctx context.Context, year string, month string) error {
	args := m.Called(ctx, year, month)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteAllActualTransactions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteActualTransactionsByYear(ctx context.Context, year string) error {
	args := m.Called(ctx, year)
	return args.Error(0)
}

func (m *MockActualRepository) DeleteActualTransactionsByMonth(ctx context.Context, year string, month string) error {
	args := m.Called(ctx, year, month)
	return args.Error(0)
}

func (m *MockActualRepository) GetAllAchHmwGle(ctx context.Context) ([]models.AchHmwGleEntity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.AchHmwGleEntity), args.Error(1)
}

func (m *MockActualRepository) GetAggregatedHMW(ctx context.Context, year string, months []string) ([]models.ActualAggregatedDTO, error) {
	args := m.Called(ctx, year, months)
	return args.Get(0).([]models.ActualAggregatedDTO), args.Error(1)
}

func (m *MockActualRepository) GetRawTransactionsHMW(ctx context.Context, year string, months []string) ([]models.ActualTransactionDTO, error) {
	args := m.Called(ctx, year, months)
	return args.Get(0).([]models.ActualTransactionDTO), args.Error(1)
}

func (m *MockActualRepository) GetAllClikGle(ctx context.Context) ([]models.ClikGleEntity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.ClikGleEntity), args.Error(1)
}

func (m *MockActualRepository) GetAggregatedCLIK(ctx context.Context, year string, months []string) ([]models.ActualAggregatedDTO, error) {
	args := m.Called(ctx, year, months)
	return args.Get(0).([]models.ActualAggregatedDTO), args.Error(1)
}

func (m *MockActualRepository) GetRawTransactionsCLIK(ctx context.Context, year string, months []string) ([]models.ActualTransactionDTO, error) {
	args := m.Called(ctx, year, months)
	return args.Get(0).([]models.ActualTransactionDTO), args.Error(1)
}

func (m *MockActualRepository) CreateActualTransactions(ctx context.Context, txs []models.ActualTransactionEntity) error {
	args := m.Called(ctx, txs)
	return args.Error(0)
}

func (m *MockActualRepository) GetRawDate(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockActualRepository) RefreshDataInventory(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// --- MockDashboardRepository ---

type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetBudgetFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockDashboardRepository) GetOrganizationStructure(ctx context.Context) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockDashboardRepository) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetDetailDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetDetailDTO), args.Error(1)
}

func (m *MockDashboardRepository) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualFactEntity), args.Error(1)
}

func (m *MockDashboardRepository) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedActualTransactionDTO), args.Error(1)
}

func (m *MockDashboardRepository) GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardSummaryDTO), args.Error(1)
}

func (m *MockDashboardRepository) GetActualYears(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDashboardRepository) GetAvailableMonths(ctx context.Context, year string) ([]string, error) {
	args := m.Called(ctx, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// --- MockPLBudgetRepository ---

type MockPLBudgetRepository struct {
	mock.Mock
}

func (m *MockPLBudgetRepository) WithTrx(trxHandle func(repo models.PLBudgetRepository) error) error {
	m.Called(trxHandle)
	return trxHandle(m)
}

func (m *MockPLBudgetRepository) CreateFileBudget(ctx context.Context, file *models.FileBudgetEntity) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockPLBudgetRepository) CreateBudgetFacts(ctx context.Context, facts []models.BudgetFactEntity) error {
	args := m.Called(ctx, facts)
	return args.Error(0)
}

func (m *MockPLBudgetRepository) ListFileBudgets(ctx context.Context) ([]models.FileBudgetEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FileBudgetEntity), args.Error(1)
}

func (m *MockPLBudgetRepository) GetFileBudget(ctx context.Context, id string) (*models.FileBudgetEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileBudgetEntity), args.Error(1)
}

func (m *MockPLBudgetRepository) DeleteFileBudget(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPLBudgetRepository) DeleteAllBudgetFacts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPLBudgetRepository) DeleteBudgetFactsByFileID(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockPLBudgetRepository) UpdateFileBudget(ctx context.Context, id string, filename string) error {
	args := m.Called(ctx, id, filename)
	return args.Error(0)
}

// --- MockMasterDataRepository ---

type MockMasterDataRepository struct {
	mock.Mock
}

func (m *MockMasterDataRepository) WithTrx(trxHandle func(repo models.MasterDataRepository) error) error {
	m.Called(trxHandle)
	return trxHandle(m)
}

func (m *MockMasterDataRepository) ListGLGroupings(ctx context.Context) ([]models.GlGroupingEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GlGroupingEntity), args.Error(1)
}

func (m *MockMasterDataRepository) GetGLGroupingByID(ctx context.Context, id string) (*models.GlGroupingEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GlGroupingEntity), args.Error(1)
}

func (m *MockMasterDataRepository) CreateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockMasterDataRepository) UpdateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockMasterDataRepository) DeleteGLGrouping(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMasterDataRepository) GetGLGroupingInfo(ctx context.Context, entity string, entityGL string, target *models.GlGroupingEntity) error {
	args := m.Called(ctx, entity, entityGL, target)
	return args.Error(0)
}

func (m *MockMasterDataRepository) GetUserConfigs(ctx context.Context, userID string) ([]models.UserConfigEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserConfigEntity), args.Error(1)
}

func (m *MockMasterDataRepository) UpdateUserConfig(ctx context.Context, config *models.UserConfigEntity) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

// --- MockAuditRepository ---

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) WithTrx(trxHandle func(repo models.AuditRepository) error) error {
	m.Called(trxHandle)
	return trxHandle(m)
}

func (m *MockAuditRepository) SaveAuditLog(ctx context.Context, log *models.AuditLogEntity) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) AddToBasket(ctx context.Context, items []models.AuditRejectBasket) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockAuditRepository) GetBasketItems(ctx context.Context, userID string) ([]models.ActualTransactionEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualTransactionEntity), args.Error(1)
}

func (m *MockAuditRepository) RemoveFromBasket(ctx context.Context, userID string, transactionID string) error {
	args := m.Called(ctx, userID, transactionID)
	return args.Error(0)
}

func (m *MockAuditRepository) SaveRejectedItems(ctx context.Context, items []models.AuditLogRejectedItemEntity) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockAuditRepository) GetAuditLogs(ctx context.Context, filter map[string]interface{}) ([]models.AuditLogEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuditLogEntity), args.Error(1)
}

func (m *MockAuditRepository) GetRejectedItemsByLogID(ctx context.Context, logID string) ([]models.AuditLogRejectedItemEntity, error) {
	args := m.Called(ctx, logID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuditLogRejectedItemEntity), args.Error(1)
}

func (m *MockAuditRepository) GetTransactionsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.ActualTransactionEntity, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualTransactionEntity), args.Error(1)
}

func (m *MockAuditRepository) GetTransactionsByFilter(ctx context.Context, filter map[string]interface{}) ([]models.ActualTransactionEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualTransactionEntity), args.Error(1)
}

func (m *MockAuditRepository) UpdateTransactionsStatus(ctx context.Context, ids []uuid.UUID, status string) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

func (m *MockAuditRepository) MarkRestAsComplete(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID, targetStatus string) error {
	args := m.Called(ctx, department, year, month, excludedIDs, targetStatus)
	return args.Error(0)
}

func (m *MockAuditRepository) GetBasketTransactionIDs(ctx context.Context, userID string) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockAuditRepository) ClearBasket(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuditRepository) ValidateBasketScope(ctx context.Context, ids []uuid.UUID, year string, month string) (bool, error) {
	args := m.Called(ctx, ids, year, month)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuditRepository) ConfirmMonthTransactions(ctx context.Context, department, year, month string, excludedIDs []uuid.UUID) error {
	args := m.Called(ctx, department, year, month, excludedIDs)
	return args.Error(0)
}

func (m *MockAuditRepository) CountPendingByDepartments(ctx context.Context, year, month string, departments []string) (int64, error) {
	args := m.Called(ctx, year, month, departments)
	return args.Get(0).(int64), args.Error(1)
}

// --- MockDashboardService ---

type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetFilterOptions(ctx context.Context) ([]models.FilterOptionDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FilterOptionDTO), args.Error(1)
}

func (m *MockDashboardService) GetRawFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockDashboardService) GetOrganizationStructure(ctx context.Context) ([]models.OrganizationDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OrganizationDTO), args.Error(1)
}

func (m *MockDashboardService) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetDetailDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetDetailDTO), args.Error(1)
}

func (m *MockDashboardService) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualFactEntity), args.Error(1)
}

func (m *MockDashboardService) GetDashboardSummary(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardSummaryDTO), args.Error(1)
}

func (m *MockDashboardService) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedActualTransactionDTO), args.Error(1)
}

func (m *MockDashboardService) GetActualYears(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDashboardService) GetAvailableMonths(ctx context.Context, year string) ([]string, error) {
	args := m.Called(ctx, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// --- MockMasterDataService ---

type MockMasterDataService struct {
	mock.Mock
}

func (m *MockMasterDataService) GetBudgetStructureTree(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func (m *MockMasterDataService) ListGLGroupings(ctx context.Context) ([]models.GlGroupingEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GlGroupingEntity), args.Error(1)
}

func (m *MockMasterDataService) GetGLGroupingByID(ctx context.Context, id string) (*models.GlGroupingEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GlGroupingEntity), args.Error(1)
}

func (m *MockMasterDataService) CreateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockMasterDataService) UpdateGLGrouping(ctx context.Context, g *models.GlGroupingEntity) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockMasterDataService) DeleteGLGrouping(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMasterDataService) ImportGLGrouping(ctx context.Context, file *multipart.FileHeader) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockMasterDataService) GetUserConfigs(ctx context.Context, userID string) (map[string]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockMasterDataService) SetUserConfig(ctx context.Context, userID string, key string, value string) error {
	args := m.Called(ctx, userID, key, value)
	return args.Error(0)
}

// --- MockDepartmentService ---

type MockDepartmentService struct {
	mock.Mock
}

func (m *MockDepartmentService) ManageDepartments(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDepartmentService) GetMasterDepartment(ctx context.Context, navCode, entity string) (*models.DepartmentEntity, error) {
	args := m.Called(ctx, navCode, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DepartmentEntity), args.Error(1)
}

// --- MockUserRepository ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]models.UserEntity, int, error) {
	args := m.Called(optional, ctx, offset, size)
	return args.Get(0).([]models.UserEntity), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) IsUserExistByID(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.UserEntity) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.UserEntity) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) ReactivateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserContext(ctx context.Context, userID string) (*models.UserEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserPermissionEntity), args.Error(1)
}

func (m *MockUserRepository) UpdateUserPermissionsAndRoles(ctx context.Context, userID string, permissions []models.UserPermissionEntity, roles []string) error {
	args := m.Called(ctx, userID, permissions, roles)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserID(ctx context.Context, oldID, newID string) error {
	args := m.Called(ctx, oldID, newID)
	return args.Error(0)
}

func (m *MockUserRepository) ListDepartments(ctx context.Context) ([]models.Departments, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockUserRepository) ListMasterDepartments(ctx context.Context) ([]models.DepartmentEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DepartmentEntity), args.Error(1)
}

func (m *MockUserRepository) GetDepartmentByCode(ctx context.Context, code string) (*models.DepartmentEntity, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DepartmentEntity), args.Error(1)
}

func (m *MockUserRepository) GetDepartmentByNavCode(ctx context.Context, navCode string) (*models.DepartmentEntity, error) {
	args := m.Called(ctx, navCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DepartmentEntity), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*models.UserEntity, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) SyncUsers(ctx context.Context, users []models.UserEntity) ([]models.UserEntity, error) {
	args := m.Called(ctx, users)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetUsers(ctx context.Context, lastID uint, limit int) ([]models.UserEntity, error) {
	args := m.Called(ctx, lastID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserEntity), args.Error(1)
}

// ============================================================
// COMMON_SERVICE.GO HELPER TESTS
// ============================================================

func TestParseDecimal_ValidNumber(t *testing.T) {
	result := parseDecimal("1234.56")
	assert.True(t, result.Equal(decimal.NewFromFloat(1234.56)))
}

func TestParseDecimal_Empty(t *testing.T) {
	result := parseDecimal("")
	assert.True(t, result.Equal(decimal.Zero))
}

func TestParseDecimal_WithCommas(t *testing.T) {
	result := parseDecimal("1,234,567.89")
	expected := decimal.NewFromFloat(1234567.89)
	assert.True(t, result.Equal(expected))
}

func TestParseDecimal_NegativeParentheses(t *testing.T) {
	result := parseDecimal("(1,000.00)")
	expected := decimal.NewFromFloat(-1000.00)
	assert.True(t, result.Equal(expected))
}

func TestParseDecimal_Dash(t *testing.T) {
	result := parseDecimal("-")
	assert.True(t, result.Equal(decimal.Zero))
}

func TestParseDecimal_Invalid(t *testing.T) {
	result := parseDecimal("abc")
	assert.True(t, result.Equal(decimal.Zero))
}

func TestNormalizeEntityCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HONDA MALIWAN", "HMW"},
		{"honda maliwan", "HMW"},
		{"AUTOCORP HOLDING", "ACG"},
		{"CLIK", "CLIK"},
		{"AC", "ACG"},
		{"MCG", "ACG"},
		{"HMW", "HMW"},
		{"UNKNOWN_ENTITY", "UNKNOWN_ENTITY"},
		{"  AC  ", "ACG"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEntityCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetColSafe_ValidIndex(t *testing.T) {
	row := []string{"a", "b", "c"}
	assert.Equal(t, "b", getColSafe(row, 1))
}

func TestGetColSafe_OutOfBounds(t *testing.T) {
	row := []string{"a", "b"}
	assert.Equal(t, "", getColSafe(row, 5))
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Budget 2025", "2025"},
		{"FY2025", "2025"},
		{"noYear", ""},
		{"Data2030Report", "2030"},
		{"2099", "2099"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractYear(tt.input))
		})
	}
}

func TestSanitizeFilter(t *testing.T) {
	t.Run("normalizes entity to entities", func(t *testing.T) {
		filter := map[string]interface{}{
			"entity": "HMW",
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"HMW"}, filter["entities"])
	})

	t.Run("normalizes branch to branches", func(t *testing.T) {
		filter := map[string]interface{}{
			"branch": "HOF",
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"HOF"}, filter["branches"])
	})

	t.Run("normalizes department to departments", func(t *testing.T) {
		filter := map[string]interface{}{
			"department": "ACC",
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"ACC"}, filter["departments"])
	})

	t.Run("handles []string values", func(t *testing.T) {
		filter := map[string]interface{}{
			"entities": []string{"HMW", "CLIK"},
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"HMW", "CLIK"}, filter["entities"])
	})

	t.Run("handles []interface{} values", func(t *testing.T) {
		filter := map[string]interface{}{
			"entities": []interface{}{"HONDA MALIWAN", "CLIK"},
		}
		sanitizeFilter(filter)
		result := filter["entities"].([]string)
		assert.Equal(t, []string{"HMW", "CLIK"}, result)
	})

	t.Run("extracts code from dash-separated values", func(t *testing.T) {
		filter := map[string]interface{}{
			"departments": []string{"ACC - Accounting", "IT - Information Technology"},
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"ACC", "IT"}, filter["departments"])
	})

	t.Run("removes empty slices", func(t *testing.T) {
		filter := map[string]interface{}{
			"entities": "",
		}
		sanitizeFilter(filter)
		_, exists := filter["entities"]
		assert.False(t, exists)
	})

	t.Run("does not overwrite existing entities key", func(t *testing.T) {
		filter := map[string]interface{}{
			"entity":   "CLIK",
			"entities": []string{"HMW"},
		}
		sanitizeFilter(filter)
		assert.Equal(t, []string{"HMW"}, filter["entities"])
	})
}

func TestExtractCode(t *testing.T) {
	assert.Equal(t, "ACC", extractCode("ACC - Accounting"))
	assert.Equal(t, "ACC", extractCode("ACC"))
	assert.Equal(t, "IT", extractCode("IT - Information Technology"))
}

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric("2025"))
	assert.True(t, isNumeric("0"))
	assert.False(t, isNumeric("20a5"))
	// Note: isNumeric("") returns true because the loop body never executes for empty strings
	assert.True(t, isNumeric(""))
}

// ============================================================
// ACTUAL SERVICE TESTS
// ============================================================

func TestNewActualService(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)

	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)
	assert.NotNil(t, svc)
}

func TestDeleteActualFacts_Success(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	repo.On("DeleteActualFactsByYear", mock.Anything, "2025").Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.DeleteActualFacts(context.Background(), "2025")

	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteActualFactsByYear", mock.Anything, "2025")
	repo.AssertCalled(t, "RefreshDataInventory", mock.Anything)
}

func TestDeleteActualFacts_EmptyYear(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	err := svc.DeleteActualFacts(context.Background(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "year is required")
}

func TestDeleteActualFacts_RepoError(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	repo.On("DeleteActualFactsByYear", mock.Anything, "2025").Return(errors.New("db error"))

	err := svc.DeleteActualFacts(context.Background(), "2025")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetRawDate_Success(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	repo.On("GetRawDate", mock.Anything).Return("2025-03-15", nil)

	result, err := svc.GetRawDate(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "2025-03-15", result)
}

func TestRefreshDataInventory_Success(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.RefreshDataInventory(context.Background())

	assert.NoError(t, err)
	repo.AssertCalled(t, "RefreshDataInventory", mock.Anything)
}

// ============================================================
// DASHBOARD SERVICE TESTS
// ============================================================

func TestNewDashboardService(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)
	assert.NotNil(t, svc)
}

func TestGetRawFilterOptions_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := []models.BudgetFactEntity{
		{Entity: "HMW", Department: "ACC", Group: "SGA"},
	}
	repo.On("GetBudgetFilterOptions", mock.Anything).Return(expected, nil)

	result, err := svc.GetRawFilterOptions(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "HMW", result[0].Entity)
}

func TestGetRawFilterOptions_Error(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	repo.On("GetBudgetFilterOptions", mock.Anything).Return(nil, errors.New("db error"))

	result, err := svc.GetRawFilterOptions(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetBudgetDetails_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := []models.BudgetDetailDTO{
		{ConsoGL: "5100", GLName: "Salary", YearTotal: decimal.NewFromInt(100000)},
	}
	repo.On("GetBudgetDetails", mock.Anything, mock.Anything).Return(expected, nil)

	filter := map[string]interface{}{"entity": "HMW"}
	result, err := svc.GetBudgetDetails(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "5100", result[0].ConsoGL)
}

func TestGetActualDetails_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := []models.ActualFactEntity{
		{Entity: "HMW", Department: "ACC"},
	}
	repo.On("GetActualDetails", mock.Anything, mock.Anything).Return(expected, nil)

	filter := map[string]interface{}{"entity": "HMW"}
	result, err := svc.GetActualDetails(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetDashboardSummary_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(500000),
		TotalActual: decimal.NewFromInt(300000),
	}
	repo.On("GetDashboardAggregates", mock.Anything, mock.Anything).Return(expected, nil)

	filter := map[string]interface{}{"entity": "HMW", "year": "2025"}
	result, err := svc.GetDashboardSummary(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalBudget.Equal(decimal.NewFromInt(500000)))
}

func TestGetActualTransactions_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := &models.PaginatedActualTransactionDTO{
		Data: []models.ActualTransactionDTO{
			{DocNo: "DOC001", Amount: decimal.NewFromInt(1000)},
		},
		TotalCount: 1,
		Page:       1,
		Limit:      10,
	}
	repo.On("GetActualTransactions", mock.Anything, mock.Anything).Return(expected, nil)

	filter := map[string]interface{}{"entity": "HMW"}
	result, err := svc.GetActualTransactions(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalCount)
}

func TestGetActualYears_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := []string{"2024", "2025"}
	repo.On("GetActualYears", mock.Anything).Return(expected, nil)

	result, err := svc.GetActualYears(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{"2024", "2025"}, result)
}

func TestGetAvailableMonths_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	expected := []string{"JAN", "FEB", "MAR"}
	repo.On("GetAvailableMonths", mock.Anything, "2025").Return(expected, nil)

	result, err := svc.GetAvailableMonths(context.Background(), "2025")

	assert.NoError(t, err)
	assert.Equal(t, []string{"JAN", "FEB", "MAR"}, result)
}

func TestGetOrganizationStructure_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	facts := []models.BudgetFactEntity{
		{Entity: "HMW", Branch: "HOF", Department: "ACC"},
		{Entity: "HMW", Branch: "HOF", Department: "IT"},
		{Entity: "CLIK", Branch: "HQ", Department: "FIN"},
	}
	repo.On("GetOrganizationStructure", mock.Anything).Return(facts, nil)

	result, err := svc.GetOrganizationStructure(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify entities are present
	entityNames := make(map[string]bool)
	for _, org := range result {
		entityNames[org.Entity] = true
	}
	assert.True(t, entityNames["HMW"])
	assert.True(t, entityNames["CLIK"])
}

func TestGetOrganizationStructure_EmptyEntity(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	facts := []models.BudgetFactEntity{
		{Entity: "", Branch: "HOF", Department: "ACC"},
		{Entity: "HMW", Branch: "HOF", Department: "IT"},
	}
	repo.On("GetOrganizationStructure", mock.Anything).Return(facts, nil)

	result, err := svc.GetOrganizationStructure(context.Background())

	assert.NoError(t, err)
	// Empty entity should be skipped
	for _, org := range result {
		assert.NotEqual(t, "", org.Entity)
	}
}

func TestGetOrganizationStructure_Error(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	repo.On("GetOrganizationStructure", mock.Anything).Return(nil, errors.New("db error"))

	result, err := svc.GetOrganizationStructure(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetFilterOptions_Success(t *testing.T) {
	repo := new(MockDashboardRepository)
	depSrv := new(MockDepartmentService)
	svc := NewDashboardService(repo, depSrv)

	facts := []models.BudgetFactEntity{
		{Group: "SGA", Department: "ACC", EntityGL: "5100", ConsoGL: "C5100", GLName: "Salary"},
		{Group: "COGS", Department: "IT", EntityGL: "6100", ConsoGL: "C6100", GLName: "Software"},
	}
	repo.On("GetBudgetFilterOptions", mock.Anything).Return(facts, nil)

	result, err := svc.GetFilterOptions(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Root nodes should be the Group level (Level 1)
	for _, node := range result {
		assert.Equal(t, 1, node.Level)
	}
}

// ============================================================
// PL BUDGET SERVICE TESTS
// ============================================================

func TestNewPLBudgetService(t *testing.T) {
	repo := new(MockPLBudgetRepository)
	depSrv := new(MockDepartmentService)
	svc := NewPLBudgetService(repo, depSrv)
	assert.NotNil(t, svc)
}

func TestListBudgetFiles_Success(t *testing.T) {
	repo := new(MockPLBudgetRepository)
	depSrv := new(MockDepartmentService)
	svc := NewPLBudgetService(repo, depSrv)

	expected := []models.FileBudgetEntity{
		{ID: uuid.New(), FileName: "Budget2025.xlsx"},
		{ID: uuid.New(), FileName: "Budget2024.xlsx"},
	}
	repo.On("ListFileBudgets", mock.Anything).Return(expected, nil)

	result, err := svc.ListBudgetFiles(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Budget2025.xlsx", result[0].FileName)
}

func TestDeleteBudgetFile_Success(t *testing.T) {
	repo := new(MockPLBudgetRepository)
	depSrv := new(MockDepartmentService)
	svc := NewPLBudgetService(repo, depSrv)

	testID := uuid.New().String()
	repo.On("DeleteFileBudget", mock.Anything, testID).Return(nil)

	err := svc.DeleteBudgetFile(context.Background(), testID)

	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteFileBudget", mock.Anything, testID)
}

func TestRenameBudgetFile_Success(t *testing.T) {
	repo := new(MockPLBudgetRepository)
	depSrv := new(MockDepartmentService)
	svc := NewPLBudgetService(repo, depSrv)

	testID := uuid.New().String()
	repo.On("UpdateFileBudget", mock.Anything, testID, "NewName.xlsx").Return(nil)

	err := svc.RenameBudgetFile(context.Background(), testID, "NewName.xlsx")

	assert.NoError(t, err)
	repo.AssertCalled(t, "UpdateFileBudget", mock.Anything, testID, "NewName.xlsx")
}

func TestClearBudget_Success(t *testing.T) {
	repo := new(MockPLBudgetRepository)
	depSrv := new(MockDepartmentService)
	svc := NewPLBudgetService(repo, depSrv)

	repo.On("WithTrx", mock.AnythingOfType("func(models.PLBudgetRepository) error")).Return(nil)
	repo.On("DeleteAllBudgetFacts", mock.Anything).Return(nil)

	err := svc.ClearBudget(context.Background())

	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteAllBudgetFacts", mock.Anything)
}

// ============================================================
// MASTER DATA SERVICE TESTS
// ============================================================

func TestNewMasterDataService(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)
	assert.NotNil(t, svc)
}

func TestListGLGroupings_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	groupings := []models.GlGroupingEntity{
		{ID: uuid.New(), Entity: "HMW", EntityGL: "5100", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
		{ID: uuid.New(), Entity: "CLIK", EntityGL: "6100", ConsoGL: "C6100", AccountName: "Software", Group1: "COGS", IsActive: true},
	}
	repo.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	result, err := svc.ListGLGroupings(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Verify sorted by Entity first
	assert.Equal(t, "CLIK", result[0].Entity)
	assert.Equal(t, "HMW", result[1].Entity)
}

func TestListGLGroupings_Error(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	repo.On("ListGLGroupings", mock.Anything).Return(nil, errors.New("db error"))

	result, err := svc.ListGLGroupings(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetGLGroupingByID_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	testID := uuid.New()
	expected := &models.GlGroupingEntity{
		ID: testID, Entity: "HMW", EntityGL: "5100", ConsoGL: "C5100",
	}
	repo.On("GetGLGroupingByID", mock.Anything, testID.String()).Return(expected, nil)

	result, err := svc.GetGLGroupingByID(context.Background(), testID.String())

	assert.NoError(t, err)
	assert.Equal(t, testID, result.ID)
}

func TestCreateGLGrouping_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	g := &models.GlGroupingEntity{
		Entity:   "  hmw  ",
		EntityGL: "5100",
		ConsoGL:  "C5100",
	}
	repo.On("CreateGLGrouping", mock.Anything, mock.AnythingOfType("*models.GlGroupingEntity")).Return(nil)

	err := svc.CreateGLGrouping(context.Background(), g)

	assert.NoError(t, err)
	// Verify entity was normalized to uppercase and trimmed
	assert.Equal(t, "HMW", g.Entity)
	// Verify UUID was set
	assert.NotEqual(t, uuid.Nil, g.ID)
}

func TestUpdateGLGrouping_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	g := &models.GlGroupingEntity{
		ID:       uuid.New(),
		Entity:   " clik ",
		EntityGL: "6100",
	}
	repo.On("UpdateGLGrouping", mock.Anything, mock.AnythingOfType("*models.GlGroupingEntity")).Return(nil)

	err := svc.UpdateGLGrouping(context.Background(), g)

	assert.NoError(t, err)
	assert.Equal(t, "CLIK", g.Entity)
}

func TestDeleteGLGrouping_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	testID := uuid.New().String()
	repo.On("DeleteGLGrouping", mock.Anything, testID).Return(nil)

	err := svc.DeleteGLGrouping(context.Background(), testID)

	assert.NoError(t, err)
	repo.AssertCalled(t, "DeleteGLGrouping", mock.Anything, testID)
}

func TestGetUserConfigs_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	configs := []models.UserConfigEntity{
		{UserID: "GLOBAL_ADMIN_SETTINGS", ConfigKey: "theme", Value: "dark"},
		{UserID: "GLOBAL_ADMIN_SETTINGS", ConfigKey: "lang", Value: "en"},
	}
	// Service always uses "GLOBAL_ADMIN_SETTINGS" regardless of userID passed
	repo.On("GetUserConfigs", mock.Anything, "GLOBAL_ADMIN_SETTINGS").Return(configs, nil)

	result, err := svc.GetUserConfigs(context.Background(), "any-user-id")

	assert.NoError(t, err)
	assert.Equal(t, "dark", result["theme"])
	assert.Equal(t, "en", result["lang"])
	assert.Len(t, result, 2)
}

func TestSetUserConfig_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	repo.On("UpdateUserConfig", mock.Anything, mock.MatchedBy(func(c *models.UserConfigEntity) bool {
		return c.UserID == "GLOBAL_ADMIN_SETTINGS" && c.ConfigKey == "theme" && c.Value == "light"
	})).Return(nil)

	err := svc.SetUserConfig(context.Background(), "any-user-id", "theme", "light")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetBudgetStructureTree_Success(t *testing.T) {
	repo := new(MockMasterDataRepository)
	svc := NewMasterDataService(repo)

	groupings := []models.GlGroupingEntity{
		{
			ID: uuid.New(), Entity: "HMW", EntityGL: "5100", ConsoGL: "C5100",
			AccountName: "Salary", Group1: "SGA", Group2: "Personnel", Group3: "Compensation",
			IsActive: true,
		},
		{
			ID: uuid.New(), Entity: "CLIK", EntityGL: "6100", ConsoGL: "C6100",
			AccountName: "Software", Group1: "COGS", Group2: "Technology", Group3: "License",
			IsActive: true,
		},
	}
	repo.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	result, err := svc.GetBudgetStructureTree(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Result should be a slice of tree nodes
	roots, ok := result.([]*struct {
		ID       string      `json:"id"`
		Name     string      `json:"name"`
		Level    int         `json:"level"`
		Children interface{} `json:"children,omitempty"`
	})
	// The actual type is an unexported TreeNode, so we just verify it is not nil
	if !ok {
		// It is an internal type, just ensure result is not nil
		assert.NotNil(t, result)
	} else {
		assert.NotEmpty(t, roots)
	}
}

// ============================================================
// AUDIT SERVICE TESTS
// ============================================================

func TestNewAuditService(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)

	svc := NewAuditService(auditRepo, dashRepo, userRepo)
	assert.NotNil(t, svc)
}

func TestListLogs_Success(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	expected := []models.AuditLogEntity{
		{ID: uuid.New(), Department: "ACC", Status: "CONFIRMED", Year: "2025", Month: "JAN"},
		{ID: uuid.New(), Department: "IT", Status: "REJECTED", Year: "2025", Month: "FEB"},
	}
	filter := map[string]interface{}{"year": "2025"}
	auditRepo.On("GetAuditLogs", mock.Anything, filter).Return(expected, nil)

	result, err := svc.ListLogs(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetRejectedItemDetails_Success(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	logID := uuid.New().String()
	expected := []models.AuditLogRejectedItemEntity{
		{ID: uuid.New(), AuditLogID: uuid.MustParse(logID), ConsoGL: "5100", Amount: decimal.NewFromInt(500)},
	}
	auditRepo.On("GetRejectedItemsByLogID", mock.Anything, logID).Return(expected, nil)

	result, err := svc.GetRejectedItemDetails(context.Background(), logID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "5100", result[0].ConsoGL)
}

func TestApprove_WithDepartment_Success(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"department": "ACC",
		"year":       "2025",
		"month":      "JAN",
		"entity":     "HMW",
		"branch":     "HOF",
	}

	// checkOwnerPermission will call GetUserPermissions
	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "OWNER"},
	}, nil)

	auditRepo.On("WithTrx", mock.AnythingOfType("func(models.AuditRepository) error")).Return(nil)
	auditRepo.On("GetBasketTransactionIDs", mock.Anything, "user-1").Return([]uuid.UUID(nil), nil)
	auditRepo.On("MarkRestAsComplete", mock.Anything, "ACC", "2025", "JAN", []uuid.UUID(nil), models.TxStatusComplete).Return(nil)
	auditRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLogEntity")).Return(nil)
	auditRepo.On("ClearBasket", mock.Anything, "user-1").Return(nil)

	err := svc.Approve(context.Background(), user, payload)

	assert.NoError(t, err)
	auditRepo.AssertCalled(t, "SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLogEntity"))
}

func TestApprove_NoDepartment_UsesPermissions(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"year":   "2025",
		"month":  "FEB",
		"entity": "HMW",
		"branch": "HOF",
	}

	// No department in payload -> look up all OWNER permissions
	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "OWNER"},
		{DepartmentCode: "IT", Role: "OWNER"},
	}, nil)

	auditRepo.On("WithTrx", mock.AnythingOfType("func(models.AuditRepository) error")).Return(nil)
	auditRepo.On("GetBasketTransactionIDs", mock.Anything, "user-1").Return([]uuid.UUID(nil), nil)
	auditRepo.On("MarkRestAsComplete", mock.Anything, mock.Anything, "2025", "FEB", []uuid.UUID(nil), models.TxStatusComplete).Return(nil)
	auditRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLogEntity")).Return(nil)
	auditRepo.On("ClearBasket", mock.Anything, "user-1").Return(nil)

	err := svc.Approve(context.Background(), user, payload)

	assert.NoError(t, err)
}

func TestApprove_NoTargets_Error(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"year":  "2025",
		"month": "JAN",
	}

	// No department in payload and user has no OWNER permissions
	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "VIEWER"},
	}, nil)

	err := svc.Approve(context.Background(), user, payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorized departments found")
}

func TestApprove_PermissionDenied(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"department": "FIN",
		"year":       "2025",
		"month":      "JAN",
	}

	// User does not have OWNER role for FIN
	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "OWNER"},
	}, nil)

	err := svc.Approve(context.Background(), user, payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission Denied")
}

func TestReport_MissingRejectedIDs_Error(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"year":  "2025",
		"month": "JAN",
	}

	err := svc.Report(context.Background(), user, payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rejected_item_ids is required")
}

func TestGetReportableTransactions_WithDepartment(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"department": "ACC",
		"year":       "2025",
		"month":      "JAN",
	}

	// checkOwnerPermission
	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "OWNER"},
	}, nil)

	expectedTxs := []models.ActualTransactionEntity{
		{ID: uuid.New(), Department: "ACC", DocNo: "DOC001", Amount: decimal.NewFromInt(1000)},
	}
	auditRepo.On("GetTransactionsByFilter", mock.Anything, mock.Anything).Return(expectedTxs, nil)

	result, err := svc.GetReportableTransactions(context.Background(), user, payload)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "DOC001", result[0].DocNo)
}

func TestGetReportableTransactions_NoDepartment_UsesPermissions(t *testing.T) {
	auditRepo := new(MockAuditRepository)
	dashRepo := new(MockDashboardRepository)
	userRepo := new(MockUserRepository)
	svc := NewAuditService(auditRepo, dashRepo, userRepo)

	user := &models.UserInfo{ID: "user-1", Name: "Test User"}
	payload := map[string]interface{}{
		"year":  "2025",
		"month": "FEB",
	}

	userRepo.On("GetUserPermissions", mock.Anything, "user-1").Return([]models.UserPermissionEntity{
		{DepartmentCode: "ACC", Role: "OWNER"},
		{DepartmentCode: "IT", Role: "OWNER"},
		{DepartmentCode: "FIN", Role: "VIEWER"},
	}, nil)

	expectedTxs := []models.ActualTransactionEntity{
		{ID: uuid.New(), Department: "ACC", DocNo: "DOC001"},
		{ID: uuid.New(), Department: "IT", DocNo: "DOC002"},
	}
	auditRepo.On("GetTransactionsByFilter", mock.Anything, mock.MatchedBy(func(f map[string]interface{}) bool {
		deps, ok := f["departments"].([]string)
		if !ok {
			return false
		}
		// Should contain ACC and IT but not FIN
		return len(deps) == 2
	})).Return(expectedTxs, nil)

	result, err := svc.GetReportableTransactions(context.Background(), user, payload)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
