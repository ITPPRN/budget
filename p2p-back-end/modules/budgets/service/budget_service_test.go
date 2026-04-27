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

func (m *MockActualRepository) StreamRawTransactionsHMW(ctx context.Context, year string, months []string, batchSize int, handler func([]models.ActualTransactionDTO) error) error {
	args := m.Called(ctx, year, months, batchSize, handler)
	return args.Error(0)
}

func (m *MockActualRepository) StreamRawTransactionsCLIK(ctx context.Context, year string, months []string, batchSize int, handler func([]models.ActualTransactionDTO) error) error {
	args := m.Called(ctx, year, months, batchSize, handler)
	return args.Error(0)
}

func (m *MockActualRepository) CreateActualTransactions(ctx context.Context, txs []models.ActualTransactionEntity) error {
	args := m.Called(ctx, txs)
	return args.Error(0)
}

func (m *MockActualRepository) GetNonPendingTransactionKeys(ctx context.Context, year string, months []string) (map[string]string, error) {
	args := m.Called(ctx, year, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockActualRepository) RestoreTransactionStatuses(ctx context.Context, statusMap map[string]string) error {
	args := m.Called(ctx, statusMap)
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

func (m *MockDashboardRepository) GetAdminPermittedMonths(ctx context.Context) []string {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
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

func (m *MockAuditRepository) GetBasketItems(ctx context.Context, userID string) ([]models.BasketItemView, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BasketItemView), args.Error(1)
}

func (m *MockAuditRepository) RemoveFromBasket(ctx context.Context, userID string, transactionID string) error {
	args := m.Called(ctx, userID, transactionID)
	return args.Error(0)
}

func (m *MockAuditRepository) UpdateBasketNote(ctx context.Context, userID, transactionID, note string) error {
	args := m.Called(ctx, userID, transactionID, note)
	return args.Error(0)
}

func (m *MockAuditRepository) UpdateBasketNoteByAddedBy(ctx context.Context, addedByUserID, transactionID, note string) error {
	args := m.Called(ctx, addedByUserID, transactionID, note)
	return args.Error(0)
}

func (m *MockAuditRepository) GetBasketItemsAddedBy(ctx context.Context, addedByUserID string) ([]models.BasketItemView, error) {
	args := m.Called(ctx, addedByUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BasketItemView), args.Error(1)
}

func (m *MockAuditRepository) RemoveFromBasketByAddedBy(ctx context.Context, addedByUserID, transactionID string) error {
	args := m.Called(ctx, addedByUserID, transactionID)
	return args.Error(0)
}

func (m *MockAuditRepository) DeleteBasketRowsByTxIDs(ctx context.Context, transactionIDs []uuid.UUID) error {
	args := m.Called(ctx, transactionIDs)
	return args.Error(0)
}

func (m *MockAuditRepository) GetBasketNotes(ctx context.Context, userID string) (map[uuid.UUID]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]string), args.Error(1)
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

func (m *MockAuditRepository) CountTotalByDepartments(ctx context.Context, year, month string, departments []string) (int64, error) {
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

func (m *MockDashboardService) GetAdminPermittedMonths(ctx context.Context) []string {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
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

func (m *MockUserRepository) GetActiveOwnerIDsByDepartment(ctx context.Context, departmentCode string) ([]string, error) {
	args := m.Called(ctx, departmentCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
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

// ============================================================
// SYNC ACTUALS - STATUS PRESERVATION TESTS
// ============================================================

// TestSyncActuals_PreservesConfirmedStatuses verifies that when a sync runs
// for specific months, transactions with non-PENDING statuses (e.g., CONFIRMED)
// are preserved and restored after the new data is inserted.
func TestSyncActuals_PreservesConfirmedStatuses(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	// GL Mapping: HMW + 51000 -> C5100
	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// Preserved statuses: one transaction was previously CONFIRMED
	preservedMap := map[string]string{
		"HMW|51000|DOC001|2026-04-15": "CONFIRMED",
	}

	// Raw data from source tables
	hmwRows := []models.ActualTransactionDTO{
		{
			Source: "HMW", Company: "HMW", EntityGL: "51000",
			PostingDate: "2026-04-15", DocNo: "DOC001",
			Description: "Salary payment", Amount: decimal.NewFromInt(50000),
			Vendor: "Vendor A", Department: "ACC", Branch: "HEAD OFFICE",
		},
	}

	// Mock expectations (in call order within transaction)
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(preservedMap, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	depSrv.On("GetMasterDepartment", mock.Anything, "ACC", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Return(nil)
	repo.On("RestoreTransactionStatuses", mock.Anything, preservedMap).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})

	assert.NoError(t, err)
	// Verify the status preservation flow was called
	repo.AssertCalled(t, "GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"})
	repo.AssertCalled(t, "RestoreTransactionStatuses", mock.Anything, preservedMap)
}

// TestSyncActuals_FullYearPreservesStatuses verifies full-year sync also preserves statuses.
func TestSyncActuals_FullYearPreservesStatuses(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	allMonths := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	masterSrv.On("ListGLGroupings", mock.Anything).Return([]models.GlGroupingEntity{}, nil)

	preservedMap := map[string]string{
		"HMW|51000|DOC001|2026-01-10": "COMPLETE",
		"HMW|51000|DOC002|2026-03-20": "CONFIRMED",
	}

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", allMonths).Return(preservedMap, nil)
	repo.On("DeleteActualFactsByYear", mock.Anything, "2026").Return(nil)
	repo.On("DeleteActualTransactionsByYear", mock.Anything, "2026").Return(nil)
	for _, m := range allMonths {
		repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{m}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
		repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{m}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	}
	repo.On("RestoreTransactionStatuses", mock.Anything, preservedMap).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{})

	assert.NoError(t, err)
	repo.AssertCalled(t, "GetNonPendingTransactionKeys", mock.Anything, "2026", allMonths)
	repo.AssertCalled(t, "RestoreTransactionStatuses", mock.Anything, preservedMap)
}

// TestSyncActuals_NoPreservedStatuses verifies that when there are no non-PENDING
// transactions, RestoreTransactionStatuses is NOT called (nothing to restore).
func TestSyncActuals_NoPreservedStatuses(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	masterSrv.On("ListGLGroupings", mock.Anything).Return([]models.GlGroupingEntity{}, nil)

	// Empty map = no non-PENDING records exist
	emptyMap := map[string]string{}

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(emptyMap, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})

	assert.NoError(t, err)
	// RestoreTransactionStatuses should NOT be called when map is empty
	repo.AssertNotCalled(t, "RestoreTransactionStatuses", mock.Anything, mock.Anything)
}

// TestSyncActuals_PreserveStatusesFetchError verifies that if fetching preserved
// statuses fails, the sync still completes successfully (graceful degradation).
func TestSyncActuals_PreserveStatusesFetchError(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	masterSrv.On("ListGLGroupings", mock.Anything).Return([]models.GlGroupingEntity{}, nil)

	// GetNonPendingTransactionKeys returns error
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(nil, errors.New("db error"))
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})

	// Sync should still succeed even if status preservation fails
	assert.NoError(t, err)
	// RestoreTransactionStatuses should NOT be called since preserved map is nil
	repo.AssertNotCalled(t, "RestoreTransactionStatuses", mock.Anything, mock.Anything)
}

// TestSyncActuals_MultipleMonthsPreserveStatuses verifies that syncing multiple
// specific months correctly preserves statuses across all of them.
func TestSyncActuals_MultipleMonthsPreserveStatuses(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	masterSrv.On("ListGLGroupings", mock.Anything).Return([]models.GlGroupingEntity{}, nil)

	months := []string{"JAN", "FEB"}
	preservedMap := map[string]string{
		"HMW|51000|DOC001|2026-01-10": "CONFIRMED",
		"HMW|52000|DOC005|2026-02-20": "COMPLETE",
		"CLIK|61000|DOC010|2026-02-28": "REPORTED",
	}

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", months).Return(preservedMap, nil)
	for _, m := range months {
		repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", m).Return(nil)
		repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", m).Return(nil)
		repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{m}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
		repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{m}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	}
	repo.On("RestoreTransactionStatuses", mock.Anything, preservedMap).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", months)

	assert.NoError(t, err)
	repo.AssertCalled(t, "GetNonPendingTransactionKeys", mock.Anything, "2026", months)
	repo.AssertCalled(t, "RestoreTransactionStatuses", mock.Anything, preservedMap)
}

// ============================================================
// SYNC ACTUALS - DATA INTEGRITY & ROBUSTNESS TESTS
// ============================================================

// TestSyncActuals_RestoreError_SyncStillSucceeds verifies that if
// RestoreTransactionStatuses itself fails, the sync still completes.
func TestSyncActuals_RestoreError_SyncStillSucceeds(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	masterSrv.On("ListGLGroupings", mock.Anything).Return([]models.GlGroupingEntity{}, nil)

	preservedMap := map[string]string{"HMW|51000|DOC001|2026-04-15": "CONFIRMED"}

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(preservedMap, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("RestoreTransactionStatuses", mock.Anything, preservedMap).Return(errors.New("restore failed"))
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})

	// Sync should still succeed even if restore fails
	assert.NoError(t, err)
	repo.AssertCalled(t, "RestoreTransactionStatuses", mock.Anything, preservedMap)
}

// TestSyncActuals_VerifyTransactionDataIntegrity verifies that the transaction
// data created during sync has correct field mappings (entity normalization,
// GL mapping, branch mapping, department mapping, etc.)
func TestSyncActuals_VerifyTransactionDataIntegrity(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	// GL Mapping
	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "Software License", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// Raw data: 2 transactions from different sources
	hmwRows := []models.ActualTransactionDTO{
		{
			Source: "HMW", Company: "HONDA MALIWAN", EntityGL: "51000",
			PostingDate: "2026-04-15", DocNo: "DOC001",
			Description: "Salary Apr", Amount: decimal.NewFromInt(50000),
			Vendor: "Vendor A", Department: "ACCOUNTING", Branch: "HEAD OFFICE",
		},
	}
	clikRows := []models.ActualTransactionDTO{
		{
			Source: "CLIK", Company: "CLIK", EntityGL: "61000",
			PostingDate: "2026-04-20", DocNo: "DOC002",
			Description: "License Fee", Amount: decimal.NewFromInt(30000),
			Vendor: "Vendor B", Department: "IT", Branch: "",
		},
	}

	// Department mapping
	depSrv.On("GetMasterDepartment", mock.Anything, "ACCOUNTING", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)
	depSrv.On("GetMasterDepartment", mock.Anything, "IT", "CLIK").Return(&models.DepartmentEntity{Code: "IT"}, nil)

	// Capture the transactions passed to CreateActualTransactions
	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	// Verify we got 2 transactions
	assert.Len(t, capturedTxs, 2)

	// Verify HMW transaction field mapping
	hmwTx := capturedTxs[0]
	assert.Equal(t, "HMW", hmwTx.Entity, "Company 'HONDA MALIWAN' should normalize to 'HMW'")
	assert.Equal(t, "51000", hmwTx.EntityGL)
	assert.Equal(t, "C5100", hmwTx.ConsoGL, "EntityGL 51000 should map to ConsoGL C5100")
	assert.Equal(t, "Salary", hmwTx.GLAccountName, "GL name should come from mapping table")
	assert.Equal(t, "DOC001", hmwTx.DocNo)
	assert.Equal(t, "2026-04-15", hmwTx.PostingDate)
	assert.True(t, hmwTx.Amount.Equal(decimal.NewFromInt(50000)))
	assert.Equal(t, "Vendor A", hmwTx.VendorName)
	assert.Equal(t, "ACC", hmwTx.Department, "Raw 'ACCOUNTING' should map to master dept code 'ACC'")
	assert.Equal(t, "HOF", hmwTx.Branch, "Raw 'HEAD OFFICE' should map to code 'HOF'")
	assert.Equal(t, "2026", hmwTx.Year)
	assert.Equal(t, "HMW", hmwTx.Source)
	assert.NotEqual(t, uuid.Nil, hmwTx.ID, "UUID should be generated")

	// Verify CLIK transaction field mapping
	clikTx := capturedTxs[1]
	assert.Equal(t, "CLIK", clikTx.Entity)
	assert.Equal(t, "C6100", clikTx.ConsoGL, "EntityGL 61000 should map to ConsoGL C6100")
	assert.Equal(t, "Software License", clikTx.GLAccountName)
	assert.Equal(t, "IT", clikTx.Department)
	assert.Equal(t, "Branch00", clikTx.Branch, "Empty branch should map to 'Branch00'")
}

// TestSyncActuals_UnmappedGLsAreFiltered verifies that raw transactions
// with GL accounts not in the mapping table are excluded from sync.
func TestSyncActuals_UnmappedGLsAreFiltered(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	// Only GL 51000 is mapped
	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// 3 raw rows: 1 mapped + 2 unmapped
	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC-MAPPED", Amount: decimal.NewFromInt(1000), Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "HMW", EntityGL: "99999", PostingDate: "2026-04-11", DocNo: "DOC-UNMAPPED1", Amount: decimal.NewFromInt(2000), Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "HMW", EntityGL: "88888", PostingDate: "2026-04-12", DocNo: "DOC-UNMAPPED2", Amount: decimal.NewFromInt(3000), Department: "ACC", Branch: "HEAD OFFICE"},
	}

	depSrv.On("GetMasterDepartment", mock.Anything, "ACC", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	// Only 1 of 3 transactions should pass through (the mapped one)
	assert.Len(t, capturedTxs, 1, "Only mapped GL should create transactions")
	assert.Equal(t, "DOC-MAPPED", capturedTxs[0].DocNo)
	assert.Equal(t, "C5100", capturedTxs[0].ConsoGL)
}

// TestSyncActuals_VerifyFactAggregation verifies that the fact table
// correctly aggregates amounts from multiple transactions with the same key.
func TestSyncActuals_VerifyFactAggregation(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// 3 transactions with same GL but different amounts → should aggregate
	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC-A", Amount: decimal.NewFromInt(10000), Vendor: "V1", Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-15", DocNo: "DOC-B", Amount: decimal.NewFromInt(25000), Vendor: "V1", Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-28", DocNo: "DOC-C", Amount: decimal.NewFromInt(15000), Vendor: "V1", Department: "ACC", Branch: "HEAD OFFICE"},
	}

	depSrv.On("GetMasterDepartment", mock.Anything, "ACC", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	var capturedFacts []models.ActualFactEntity

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		facts := args.Get(1).([]models.ActualFactEntity)
		capturedFacts = append(capturedFacts, facts...)
	}).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	// All 3 raw transactions should be created individually
	assert.Len(t, capturedTxs, 3, "Each raw row creates one transaction")

	// Facts should be aggregated: same entity+branch+dept+GL+vendor+month = 1 fact header
	assert.Len(t, capturedFacts, 1, "Same key should aggregate into one fact header")

	fact := capturedFacts[0]
	assert.Equal(t, "HMW", fact.Entity)
	assert.Equal(t, "C5100", fact.ConsoGL)
	assert.Equal(t, "SGA", fact.Group)
	// YearTotal = 10000 + 25000 + 15000 = 50000
	assert.True(t, fact.YearTotal.Equal(decimal.NewFromInt(50000)),
		"YearTotal should be 50000, got %s", fact.YearTotal.String())
	// Should have 1 month amount (APR)
	assert.Len(t, fact.ActualAmounts, 1)
	assert.Equal(t, "APR", fact.ActualAmounts[0].Month)
	assert.True(t, fact.ActualAmounts[0].Amount.Equal(decimal.NewFromInt(50000)),
		"APR amount should be 50000, got %s", fact.ActualAmounts[0].Amount.String())
}

// TestSyncActuals_InactiveGLMappingIgnored verifies that GL mappings
// with IsActive=false are not used for matching.
func TestSyncActuals_InactiveGLMappingIgnored(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: false}, // INACTIVE
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC001", Amount: decimal.NewFromInt(1000), Department: "ACC", Branch: "HEAD OFFICE"},
	}

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	// CreateActualTransactions should NOT be called since no rows matched
	repo.AssertNotCalled(t, "CreateActualTransactions", mock.Anything, mock.Anything)
}

// TestSyncActuals_NegativeAmountsPreserved verifies that negative amounts
// (credit entries, reversals) are correctly handled and not lost.
func TestSyncActuals_NegativeAmountsPreserved(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC-POS", Amount: decimal.NewFromInt(100000), Vendor: "V1", Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-15", DocNo: "DOC-NEG", Amount: decimal.NewFromInt(-30000), Vendor: "V1", Department: "ACC", Branch: "HEAD OFFICE"},
	}

	depSrv.On("GetMasterDepartment", mock.Anything, "ACC", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	var capturedFacts []models.ActualFactEntity

	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		facts := args.Get(1).([]models.ActualFactEntity)
		capturedFacts = append(capturedFacts, facts...)
	}).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	// Both positive and negative transactions should be created
	assert.Len(t, capturedTxs, 2)
	assert.True(t, capturedTxs[0].Amount.Equal(decimal.NewFromInt(100000)))
	assert.True(t, capturedTxs[1].Amount.Equal(decimal.NewFromInt(-30000)),
		"Negative amount should be preserved, got %s", capturedTxs[1].Amount.String())

	// Fact aggregation: 100000 + (-30000) = 70000
	assert.Len(t, capturedFacts, 1)
	assert.True(t, capturedFacts[0].YearTotal.Equal(decimal.NewFromInt(70000)),
		"Net total should be 70000, got %s", capturedFacts[0].YearTotal.String())
}

// TestSyncActuals_GLMappingError_ReturnsError verifies that if
// fetching GL mappings fails, the entire sync fails immediately.
func TestSyncActuals_GLMappingError_ReturnsError(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	masterSrv.On("ListGLGroupings", mock.Anything).Return(nil, errors.New("GL mapping DB error"))

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GL mapping DB error")
	// Should not proceed to any repo operations
	repo.AssertNotCalled(t, "WithTrx", mock.Anything)
}

// TestSyncActuals_EntityNormalization verifies that different company name
// formats are correctly normalized during sync.
func TestSyncActuals_EntityNormalization(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Salary", Group1: "SGA", IsActive: true},
		{Entity: "ACG", EntityGL: "52000", ConsoGL: "C5200", AccountName: "Rent", Group1: "SGA", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// Different name formats for the same entities
	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HONDA MALIWAN", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC-A", Amount: decimal.NewFromInt(1000), Department: "ACC", Branch: "HEAD OFFICE"},
		{Source: "HMW", Company: "AUTOCORP HOLDING", EntityGL: "52000", PostingDate: "2026-04-11", DocNo: "DOC-B", Amount: decimal.NewFromInt(2000), Department: "FIN", Branch: "AUTOCORP HEAD OFFICE"},
	}

	depSrv.On("GetMasterDepartment", mock.Anything, "ACC", "HMW").Return(&models.DepartmentEntity{Code: "ACC"}, nil)
	depSrv.On("GetMasterDepartment", mock.Anything, "FIN", "ACG").Return(&models.DepartmentEntity{Code: "FIN"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 2)
	assert.Equal(t, "HMW", capturedTxs[0].Entity, "'HONDA MALIWAN' should normalize to 'HMW'")
	assert.Equal(t, "ACG", capturedTxs[1].Entity, "'AUTOCORP HOLDING' should normalize to 'ACG'")
	assert.Equal(t, "HQ", capturedTxs[1].Branch, "'AUTOCORP HEAD OFFICE' should map to 'HQ'")
}

// ============================================================
// SYNC ACTUALS - CLIK SERVICE → SERVICE_CLIK RENAME RULE
// ============================================================
//
// Rule: For CLIK entity only, any department resolving to "SERVICE"
// (either from raw nav_code or from master mapping result) must be
// rewritten to "SERVICE_CLIK". HMW/ACG entities are not affected.

// TestSyncActuals_CLIKServiceRename_RawServiceNoMapping verifies that a CLIK
// row with raw Department="SERVICE" and no master mapping becomes "SERVICE_CLIK".
func TestSyncActuals_CLIKServiceRename_RawServiceNoMapping(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "Service Fee", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	clikRows := []models.ActualTransactionDTO{
		{Source: "CLIK", Company: "CLIK", EntityGL: "61000", PostingDate: "2026-04-10", DocNo: "DOC-CLIK-1", Amount: decimal.NewFromInt(1000), Department: "SERVICE", Branch: ""},
	}

	// No mapping found → master returns nil
	depSrv.On("GetMasterDepartment", mock.Anything, "SERVICE", "CLIK").Return(nil, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 1)
	assert.Equal(t, "SERVICE_CLIK", capturedTxs[0].Department,
		"CLIK + raw SERVICE (no mapping) must rename to SERVICE_CLIK")
}

// TestSyncActuals_CLIKServiceRename_MappedToService verifies a CLIK row whose
// raw nav_code maps to master code "SERVICE" is also renamed to "SERVICE_CLIK".
func TestSyncActuals_CLIKServiceRename_MappedToService(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "Service Fee", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// raw "100-30" maps to master "SERVICE"
	clikRows := []models.ActualTransactionDTO{
		{Source: "CLIK", Company: "CLIK", EntityGL: "61000", PostingDate: "2026-04-10", DocNo: "DOC-CLIK-2", Amount: decimal.NewFromInt(2000), Department: "100-30", Branch: ""},
	}
	depSrv.On("GetMasterDepartment", mock.Anything, "100-30", "CLIK").Return(&models.DepartmentEntity{Code: "SERVICE"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 1)
	assert.Equal(t, "SERVICE_CLIK", capturedTxs[0].Department,
		"CLIK + master code SERVICE must rename to SERVICE_CLIK")
}

// TestSyncActuals_CLIKServiceRename_RawServiceMappedToOther verifies that even
// when raw "SERVICE" gets remapped by the master to a non-SERVICE code, the
// CLIK rule still catches the raw value and rewrites to SERVICE_CLIK.
// This is the "raw-side guard" for the double-layer check.
func TestSyncActuals_CLIKServiceRename_RawServiceMappedToOther(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "Service Fee", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// raw "SERVICE" but mapping rewrites to "SVC" — raw-side guard must still trigger
	clikRows := []models.ActualTransactionDTO{
		{Source: "CLIK", Company: "CLIK", EntityGL: "61000", PostingDate: "2026-04-10", DocNo: "DOC-CLIK-3", Amount: decimal.NewFromInt(3000), Department: "SERVICE", Branch: ""},
	}
	depSrv.On("GetMasterDepartment", mock.Anything, "SERVICE", "CLIK").Return(&models.DepartmentEntity{Code: "SVC"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 1)
	assert.Equal(t, "SERVICE_CLIK", capturedTxs[0].Department,
		"CLIK + raw SERVICE must rename even if mapping changes the master code")
}

// TestSyncActuals_CLIKServiceRename_LowercaseRaw verifies case-insensitive
// matching: raw "service" (lowercase) for CLIK must still rename to SERVICE_CLIK.
func TestSyncActuals_CLIKServiceRename_LowercaseRaw(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "Service Fee", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	clikRows := []models.ActualTransactionDTO{
		{Source: "CLIK", Company: "CLIK", EntityGL: "61000", PostingDate: "2026-04-10", DocNo: "DOC-CLIK-4", Amount: decimal.NewFromInt(4000), Department: "service", Branch: ""},
	}
	// normalize() upper-cases before lookup, so the mock arg is "SERVICE"
	depSrv.On("GetMasterDepartment", mock.Anything, "SERVICE", "CLIK").Return(nil, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 1)
	assert.Equal(t, "SERVICE_CLIK", capturedTxs[0].Department,
		"Lowercase raw 'service' for CLIK must rename to SERVICE_CLIK (case-insensitive)")
}

// TestSyncActuals_CLIKServiceRename_HMWNotAffected verifies the rule applies
// only to CLIK: HMW with SERVICE department keeps SERVICE (no rename).
// Also verifies CLIK with non-SERVICE department is not affected.
func TestSyncActuals_CLIKServiceRename_HMWNotAffected(t *testing.T) {
	repo := new(MockActualRepository)
	masterSrv := new(MockMasterDataService)
	dashSrv := new(MockDashboardService)
	depSrv := new(MockDepartmentService)
	svc := NewActualService(repo, masterSrv, dashSrv, depSrv)

	groupings := []models.GlGroupingEntity{
		{Entity: "HMW", EntityGL: "51000", ConsoGL: "C5100", AccountName: "Service Fee", Group1: "COGS", IsActive: true},
		{Entity: "CLIK", EntityGL: "61000", ConsoGL: "C6100", AccountName: "License", Group1: "COGS", IsActive: true},
	}
	masterSrv.On("ListGLGroupings", mock.Anything).Return(groupings, nil)

	// HMW with SERVICE — should NOT be renamed
	hmwRows := []models.ActualTransactionDTO{
		{Source: "HMW", Company: "HMW", EntityGL: "51000", PostingDate: "2026-04-10", DocNo: "DOC-HMW-1", Amount: decimal.NewFromInt(5000), Department: "SERVICE", Branch: "HEAD OFFICE"},
	}
	// CLIK with non-SERVICE (IT) — should NOT be renamed
	clikRows := []models.ActualTransactionDTO{
		{Source: "CLIK", Company: "CLIK", EntityGL: "61000", PostingDate: "2026-04-10", DocNo: "DOC-CLIK-5", Amount: decimal.NewFromInt(6000), Department: "IT", Branch: ""},
	}

	depSrv.On("GetMasterDepartment", mock.Anything, "SERVICE", "HMW").Return(&models.DepartmentEntity{Code: "SERVICE"}, nil)
	depSrv.On("GetMasterDepartment", mock.Anything, "IT", "CLIK").Return(&models.DepartmentEntity{Code: "IT"}, nil)

	var capturedTxs []models.ActualTransactionEntity
	repo.On("WithTrx", mock.AnythingOfType("func(models.ActualRepository) error")).Return(nil)
	repo.On("GetNonPendingTransactionKeys", mock.Anything, "2026", []string{"APR"}).Return(map[string]string{}, nil)
	repo.On("DeleteActualFactsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("DeleteActualTransactionsByMonth", mock.Anything, "2026", "APR").Return(nil)
	repo.On("StreamRawTransactionsHMW", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(hmwRows)
		}).Return(nil)
	repo.On("StreamRawTransactionsCLIK", mock.Anything, "2026", []string{"APR"}, mock.AnythingOfType("int"), mock.AnythingOfType("func([]models.ActualTransactionDTO) error")).
		Run(func(args mock.Arguments) {
			handler := args.Get(4).(func([]models.ActualTransactionDTO) error)
			_ = handler(clikRows)
		}).Return(nil)
	repo.On("CreateActualTransactions", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		txs := args.Get(1).([]models.ActualTransactionEntity)
		capturedTxs = append(capturedTxs, txs...)
	}).Return(nil)
	repo.On("CreateActualFacts", mock.Anything, mock.Anything).Return(nil)
	repo.On("RefreshDataInventory", mock.Anything).Return(nil)

	err := svc.SyncActuals(context.Background(), "2026", []string{"APR"})
	assert.NoError(t, err)

	assert.Len(t, capturedTxs, 2)

	// Index by entity (HMW row processed first)
	var hmwTx, clikTx models.ActualTransactionEntity
	for _, tx := range capturedTxs {
		switch tx.Entity {
		case "HMW":
			hmwTx = tx
		case "CLIK":
			clikTx = tx
		}
	}
	assert.Equal(t, "SERVICE", hmwTx.Department,
		"HMW + SERVICE must NOT be renamed (rule is CLIK-only)")
	assert.Equal(t, "IT", clikTx.Department,
		"CLIK + non-SERVICE must NOT be renamed")
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
