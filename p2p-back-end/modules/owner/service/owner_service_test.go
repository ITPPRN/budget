package service

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

// ============================================================
// TestMain
// ============================================================

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

// ============================================================
// Mock: OwnerRepository
// ============================================================

type MockOwnerRepository struct {
	mock.Mock
}

func (m *MockOwnerRepository) GetBudgetFilterOptions(ctx context.Context) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockOwnerRepository) GetOrganizationStructure(ctx context.Context) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockOwnerRepository) GetDashboardAggregates(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardSummaryDTO), args.Error(1)
}

func (m *MockOwnerRepository) GetActualTransactions(ctx context.Context, filter map[string]interface{}) (*models.PaginatedActualTransactionDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedActualTransactionDTO), args.Error(1)
}

func (m *MockOwnerRepository) GetActualDetails(ctx context.Context, filter map[string]interface{}) ([]models.ActualFactEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualFactEntity), args.Error(1)
}

func (m *MockOwnerRepository) GetBudgetDetails(ctx context.Context, filter map[string]interface{}) ([]models.BudgetFactEntity, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.BudgetFactEntity), args.Error(1)
}

func (m *MockOwnerRepository) GetActualYears(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ============================================================
// Mock: AuthService
// ============================================================

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginReq) (*gocloak.JWT, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*gocloak.JWT, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, oldPassword, newPassword string, userInfo *models.UserInfo) error {
	args := m.Called(ctx, oldPassword, newPassword, userInfo)
	return args.Error(0)
}

func (m *MockAuthService) AdminResetUserPassword(ctx context.Context, targetUserID string, newPassword string) error {
	args := m.Called(ctx, targetUserID, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) GetUserProfile(ctx context.Context, userID string) (*models.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) ProvisionUser(ctx context.Context, user *models.UserInfo) (*models.UserInfo, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) ListUsersForAdmin(ctx context.Context, optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	args := m.Called(ctx, optional, page, size)
	return args.Get(0).([]models.UserInfo), args.Int(1), args.Error(2)
}

func (m *MockAuthService) ListUsersForManagement(ctx context.Context, optional map[string]interface{}, page, size int) ([]models.UserInfo, int, error) {
	args := m.Called(ctx, optional, page, size)
	return args.Get(0).([]models.UserInfo), args.Int(1), args.Error(2)
}

func (m *MockAuthService) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserPermissionInfo), args.Error(1)
}

func (m *MockAuthService) UpdateUserPermissions(ctx context.Context, userID string, perms []models.UserPermissionInfo, roles []string) error {
	args := m.Called(ctx, userID, perms, roles)
	return args.Error(0)
}

func (m *MockAuthService) ListDepartments(ctx context.Context, mappedOnly bool, user *models.UserInfo) ([]models.Departments, error) {
	args := m.Called(ctx, mappedOnly, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Departments), args.Error(1)
}

// ============================================================
// Mock: CapexService
// ============================================================

type MockCapexService struct {
	mock.Mock
}

func (m *MockCapexService) ImportCapexBudget(ctx context.Context, file *multipart.FileHeader, userID string, versionName string) error {
	args := m.Called(ctx, file, userID, versionName)
	return args.Error(0)
}

func (m *MockCapexService) ImportCapexActual(ctx context.Context, file *multipart.FileHeader, userID string, versionName string) error {
	args := m.Called(ctx, file, userID, versionName)
	return args.Error(0)
}

func (m *MockCapexService) SyncCapexBudget(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexService) SyncCapexActual(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexService) ClearCapexBudget(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapexService) ClearCapexActual(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapexService) ListCapexBudgetFiles(ctx context.Context) ([]models.FileCapexBudgetEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FileCapexBudgetEntity), args.Error(1)
}

func (m *MockCapexService) ListCapexActualFiles(ctx context.Context) ([]models.FileCapexActualEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FileCapexActualEntity), args.Error(1)
}

func (m *MockCapexService) DeleteCapexBudgetFile(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexService) DeleteCapexActualFile(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCapexService) RenameCapexBudgetFile(ctx context.Context, id string, newName string) error {
	args := m.Called(ctx, id, newName)
	return args.Error(0)
}

func (m *MockCapexService) RenameCapexActualFile(ctx context.Context, id string, newName string) error {
	args := m.Called(ctx, id, newName)
	return args.Error(0)
}

func (m *MockCapexService) GetCapexDashboardSummary(ctx context.Context, filter map[string]interface{}) (*models.DashboardSummaryDTO, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardSummaryDTO), args.Error(1)
}

// ============================================================
// Helpers
// ============================================================

func newAdminUser() *models.UserInfo {
	return &models.UserInfo{
		ID:       "admin-001",
		Username: "admin",
		Name:     "Admin User",
		Roles:    []string{"ADMIN"},
		Permissions: []models.UserPermissionInfo{
			{DepartmentCode: "ALL", Role: "admin", IsActive: true},
		},
	}
}

func newOwnerUser(perms []models.UserPermissionInfo) *models.UserInfo {
	return &models.UserInfo{
		ID:          "owner-001",
		Username:    "owneruser",
		Name:        "Owner User",
		Roles:       []string{"OWNER"},
		Permissions: perms,
	}
}

func setupService() (*MockOwnerRepository, *MockAuthService, *MockCapexService, models.OwnerService) {
	repo := new(MockOwnerRepository)
	authSrv := new(MockAuthService)
	capexSrv := new(MockCapexService)
	svc := NewOwnerService(repo, authSrv, capexSrv)
	return repo, authSrv, capexSrv, svc
}

// Suppress unused import warnings -- these are used in mock signatures.
var (
	_ = uuid.UUID{}
	_ = (*gocloak.JWT)(nil)
)

// ============================================================
// Tests
// ============================================================

func TestNewOwnerService(t *testing.T) {
	repo, authSrv, capexSrv, svc := setupService()
	assert.NotNil(t, svc)

	// Verify the concrete type has the correct fields (white-box)
	concrete, ok := svc.(*ownerService)
	assert.True(t, ok)
	assert.Equal(t, repo, concrete.repo)
	assert.Equal(t, authSrv, concrete.authSrv)
	assert.Equal(t, capexSrv, concrete.capexSrv)
}

func TestGetDashboardSummary_AdminUser(t *testing.T) {
	repo, _, capexSrv, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	expectedSummary := &models.DashboardSummaryDTO{
		TotalBudget:    decimal.NewFromInt(1000000),
		TotalActual:    decimal.NewFromInt(750000),
		DepartmentData: []models.DepartmentStatDTO{{Department: "IT", Budget: decimal.NewFromInt(500000), Actual: decimal.NewFromInt(300000)}},
	}

	capexSummary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(200000),
		TotalActual: decimal.NewFromInt(100000),
	}

	// Admin user: no permission injection restricts anything
	repo.On("GetDashboardAggregates", ctx, mock.Anything).Return(expectedSummary, nil)
	capexSrv.On("GetCapexDashboardSummary", ctx, mock.Anything).Return(capexSummary, nil)

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalBudget.Equal(decimal.NewFromInt(1000000)))
	assert.True(t, result.TotalActual.Equal(decimal.NewFromInt(750000)))
	assert.True(t, result.CapexBudget.Equal(decimal.NewFromInt(200000)))
	assert.True(t, result.CapexActual.Equal(decimal.NewFromInt(100000)))
	assert.Len(t, result.DepartmentData, 1)
	repo.AssertExpectations(t)
	capexSrv.AssertExpectations(t)
}

func TestGetDashboardSummary_OwnerUser(t *testing.T) {
	repo, _, capexSrv, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "ACC", Role: "owner", IsActive: true},
		{DepartmentCode: "FIN", Role: "owner", IsActive: true},
	})
	filter := map[string]interface{}{"year": "2025"}

	expectedSummary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(500000),
		TotalActual: decimal.NewFromInt(300000),
	}

	capexSummary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(50000),
		TotalActual: decimal.NewFromInt(25000),
	}

	repo.On("GetDashboardAggregates", ctx, mock.Anything).Return(expectedSummary, nil)
	capexSrv.On("GetCapexDashboardSummary", ctx, mock.Anything).Return(capexSummary, nil)

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalBudget.Equal(decimal.NewFromInt(500000)))
	assert.True(t, result.CapexBudget.Equal(decimal.NewFromInt(50000)))
	repo.AssertExpectations(t)
	capexSrv.AssertExpectations(t)
}

func TestGetDashboardSummary_RepoError(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	repo.On("GetDashboardAggregates", ctx, mock.Anything).Return(nil, errors.New("db connection failed"))

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ownerSrv.GetDashboardSummary")
	repo.AssertExpectations(t)
}

func TestGetActualTransactions_Success(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	expected := &models.PaginatedActualTransactionDTO{
		Data: []models.ActualTransactionDTO{
			{DocNo: "DOC-001", Amount: decimal.NewFromInt(5000), Department: "IT"},
		},
		TotalCount: 1,
		Page:       1,
		Limit:      10,
	}

	repo.On("GetActualTransactions", ctx, mock.Anything).Return(expected, nil)

	result, err := svc.GetActualTransactions(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalCount)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "DOC-001", result.Data[0].DocNo)
	repo.AssertExpectations(t)
}

func TestGetActualDetails_Success(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	expected := []models.ActualFactEntity{
		{Entity: "HMW", Department: "IT", Year: "2025"},
	}

	repo.On("GetActualDetails", ctx, mock.Anything).Return(expected, nil)

	result, err := svc.GetActualDetails(ctx, user, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "HMW", result[0].Entity)
	repo.AssertExpectations(t)
}

func TestGetBudgetDetails_Success(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	expected := []models.BudgetFactEntity{
		{Entity: "HMW", Department: "FIN", Year: "2025", YearTotal: decimal.NewFromInt(100000)},
	}

	repo.On("GetBudgetDetails", ctx, mock.Anything).Return(expected, nil)

	result, err := svc.GetBudgetDetails(ctx, user, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "FIN", result[0].Department)
	assert.True(t, result[0].YearTotal.Equal(decimal.NewFromInt(100000)))
	repo.AssertExpectations(t)
}

func TestGetFilterOptions_Admin(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	allFacts := []models.BudgetFactEntity{
		{Entity: "HMW", Department: "IT", Year: "2025"},
		{Entity: "HMW", Department: "FIN", Year: "2025"},
		{Entity: "ACG", Department: "HR", Year: "2025"},
	}

	repo.On("GetBudgetFilterOptions", ctx).Return(allFacts, nil)

	result, err := svc.GetFilterOptions(ctx, user)

	assert.NoError(t, err)
	facts, ok := result.([]models.BudgetFactEntity)
	assert.True(t, ok, "admin should receive raw []BudgetFactEntity")
	assert.Len(t, facts, 3)
	repo.AssertExpectations(t)
}

func TestGetFilterOptions_Owner(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "IT", Role: "owner", IsActive: true},
	})

	allFacts := []models.BudgetFactEntity{
		{Entity: "HMW", Department: "IT", Year: "2025"},
		{Entity: "HMW", Department: "FIN", Year: "2025"},
		{Entity: "ACG", Department: "HR", Year: "2025"},
	}

	repo.On("GetBudgetFilterOptions", ctx).Return(allFacts, nil)

	result, err := svc.GetFilterOptions(ctx, user)

	assert.NoError(t, err)
	facts, ok := result.([]models.BudgetFactEntity)
	assert.True(t, ok)
	// Only IT department should pass the filter
	assert.Len(t, facts, 1)
	assert.Equal(t, "IT", facts[0].Department)
	repo.AssertExpectations(t)
}

func TestGetOrganizationStructure_Admin(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	facts := []models.BudgetFactEntity{
		{Entity: "HMW", Branch: "BKK", Department: "IT"},
		{Entity: "HMW", Branch: "BKK", Department: "FIN"},
		{Entity: "ACG", Branch: "CNX", Department: "HR"},
	}

	repo.On("GetOrganizationStructure", ctx).Return(facts, nil)

	result, err := svc.GetOrganizationStructure(ctx, user)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Admin sees all entities
	totalDepts := 0
	for _, org := range result {
		for _, br := range org.Branches {
			totalDepts += len(br.Departments)
		}
	}
	assert.Equal(t, 3, totalDepts)
	repo.AssertExpectations(t)
}

func TestGetOrganizationStructure_Owner(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "IT", Role: "owner", IsActive: true},
	})

	facts := []models.BudgetFactEntity{
		{Entity: "HMW", Branch: "BKK", Department: "IT"},
		{Entity: "HMW", Branch: "BKK", Department: "FIN"},
		{Entity: "ACG", Branch: "CNX", Department: "HR"},
	}

	repo.On("GetOrganizationStructure", ctx).Return(facts, nil)

	result, err := svc.GetOrganizationStructure(ctx, user)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Owner with IT permission should only see IT department
	totalDepts := 0
	for _, org := range result {
		for _, br := range org.Branches {
			totalDepts += len(br.Departments)
		}
	}
	assert.Equal(t, 1, totalDepts)
	// The only visible entity/branch should contain IT
	assert.Len(t, result, 1)
	assert.Equal(t, "HMW", result[0].Entity)
	assert.Len(t, result[0].Branches, 1)
	assert.Contains(t, result[0].Branches[0].Departments, "IT")
	repo.AssertExpectations(t)
}

func TestGetOwnerFilterLists_Success(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	repo.On("GetActualYears", ctx).Return([]string{"2024", "2025"}, nil)

	result, err := svc.GetOwnerFilterLists(ctx, user)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"HMW", "ACG", "CLIK"}, result.Companies)
	assert.Equal(t, []string{}, result.Branches)
	assert.Equal(t, []string{"2024", "2025"}, result.Years)
	repo.AssertExpectations(t)
}

func TestGetActualYears_Success(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	repo.On("GetActualYears", ctx).Return([]string{"2023", "2024", "2025"}, nil)

	result, err := svc.GetActualYears(ctx, user)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "2023", result[0])
	repo.AssertExpectations(t)
}

func TestInjectPermissions_AdminUser(t *testing.T) {
	_, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	result := svc.InjectPermissions(ctx, user, filter)

	// Admin: no departments filter should be injected
	_, hasDepts := result["departments"]
	assert.False(t, hasDepts, "admin user should NOT have departments filter")
	_, hasRestricted := result["is_restricted"]
	assert.False(t, hasRestricted, "admin user should NOT be restricted")
}

func TestInjectPermissions_OwnerUser(t *testing.T) {
	_, _, _, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "ACC", Role: "owner", IsActive: true},
		{DepartmentCode: "FIN", Role: "owner", IsActive: true},
		{DepartmentCode: "OLD", Role: "owner", IsActive: false}, // inactive, should be excluded
	})
	filter := map[string]interface{}{"year": "2025"}

	result := svc.InjectPermissions(ctx, user, filter)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"ACC", "FIN"}, depts)
	assert.Equal(t, true, result["is_restricted"])
}

func TestInjectPermissions_OwnerNoPermissions(t *testing.T) {
	_, authSrv, _, svc := setupService()
	ctx := context.Background()
	// Owner with no permissions and empty ID (so no lookup)
	user := &models.UserInfo{
		ID:       "",
		Username: "owneruser",
		Roles:    []string{"OWNER"},
	}
	filter := map[string]interface{}{}

	// authSrv.GetUserPermissions should NOT be called since ID is empty
	result := svc.InjectPermissions(ctx, user, filter)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"__RESTRICTED__"}, depts)
	assert.Equal(t, true, result["is_restricted"])
	authSrv.AssertNotCalled(t, "GetUserPermissions", mock.Anything, mock.Anything)
}

func TestInjectPermissions_OwnerNoPermissions_WithLookup(t *testing.T) {
	_, authSrv, _, svc := setupService()
	ctx := context.Background()
	// Owner with no local permissions but has an ID, so authSrv.GetUserPermissions is called
	user := &models.UserInfo{
		ID:       "owner-002",
		Username: "owneruser2",
		Roles:    []string{"OWNER"},
	}

	// Return empty permissions from auth service
	authSrv.On("GetUserPermissions", ctx, "owner-002").Return([]models.UserPermissionInfo{}, nil)

	result := svc.InjectPermissions(ctx, user, nil)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"__RESTRICTED__"}, depts)
	assert.Equal(t, true, result["is_restricted"])
	authSrv.AssertExpectations(t)
}

func TestInjectPermissions_OwnerWithExistingFilter(t *testing.T) {
	_, _, _, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "ACC", Role: "owner", IsActive: true},
		{DepartmentCode: "FIN", Role: "owner", IsActive: true},
		{DepartmentCode: "IT", Role: "owner", IsActive: true},
	})

	// Frontend already provided a departments filter; intersection should be applied
	filter := map[string]interface{}{
		"year":        "2025",
		"departments": []string{"ACC", "HR"}, // HR is NOT in user's permissions
	}

	result := svc.InjectPermissions(ctx, user, filter)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	// Only ACC should survive the intersection (HR is not permitted)
	assert.Equal(t, []string{"ACC"}, depts)
	assert.Equal(t, true, result["is_restricted"])
}

func TestInjectPermissions_OwnerWithExistingFilter_NoOverlap(t *testing.T) {
	_, _, _, svc := setupService()
	ctx := context.Background()
	user := newOwnerUser([]models.UserPermissionInfo{
		{DepartmentCode: "ACC", Role: "owner", IsActive: true},
	})

	// Frontend requests departments the owner has no access to
	filter := map[string]interface{}{
		"departments": []string{"HR", "LEGAL"},
	}

	result := svc.InjectPermissions(ctx, user, filter)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	// No overlap -> restricted
	assert.Equal(t, []string{"__RESTRICTED__"}, depts)
	assert.Equal(t, true, result["is_restricted"])
}

func TestInjectPermissions_NilFilter(t *testing.T) {
	_, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	result := svc.InjectPermissions(ctx, user, nil)

	assert.NotNil(t, result, "nil filter should be initialized to an empty map")
	_, hasDepts := result["departments"]
	assert.False(t, hasDepts)
}

func TestGetDashboardSummary_CapexError_StillReturns(t *testing.T) {
	repo, _, capexSrv, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025"}

	expectedSummary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(1000000),
		TotalActual: decimal.NewFromInt(750000),
	}

	repo.On("GetDashboardAggregates", ctx, mock.Anything).Return(expectedSummary, nil)
	// Capex fails, but dashboard should still return with zero capex values
	capexSrv.On("GetCapexDashboardSummary", ctx, mock.Anything).Return(nil, errors.New("capex db error"))

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalBudget.Equal(decimal.NewFromInt(1000000)))
	assert.True(t, result.CapexBudget.Equal(decimal.Zero))
	assert.True(t, result.CapexActual.Equal(decimal.Zero))
	repo.AssertExpectations(t)
	capexSrv.AssertExpectations(t)
}

func TestGetDashboardSummary_ConsoGLsMapping(t *testing.T) {
	repo, _, capexSrv, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	// Include conso_gls in the filter to verify it gets mapped to budget_gls
	filter := map[string]interface{}{
		"year":      "2025",
		"conso_gls": []string{"GL001", "GL002"},
	}

	expectedSummary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(100),
		TotalActual: decimal.NewFromInt(50),
	}

	repo.On("GetDashboardAggregates", ctx, mock.MatchedBy(func(f map[string]interface{}) bool {
		// Verify conso_gls got mapped to budget_gls
		budgetGLs, ok := f["budget_gls"]
		if !ok {
			return false
		}
		gls, ok := budgetGLs.([]string)
		return ok && len(gls) == 2
	})).Return(expectedSummary, nil)
	capexSrv.On("GetCapexDashboardSummary", ctx, mock.Anything).Return(nil, errors.New("skip"))

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	repo.AssertExpectations(t)
}

func TestGetFilterOptions_RepoError(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	repo.On("GetBudgetFilterOptions", ctx).Return(nil, errors.New("db error"))

	result, err := svc.GetFilterOptions(ctx, user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ownerSrv.GetFilterOptions")
	repo.AssertExpectations(t)
}

func TestGetOrganizationStructure_RepoError(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	repo.On("GetOrganizationStructure", ctx).Return(nil, errors.New("db error"))

	result, err := svc.GetOrganizationStructure(ctx, user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ownerSrv.GetOrganizationStructure")
	repo.AssertExpectations(t)
}

func TestGetOwnerFilterLists_RepoError(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	repo.On("GetActualYears", ctx).Return(nil, errors.New("years fetch failed"))

	result, err := svc.GetOwnerFilterLists(ctx, user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ownerSrv.GetOwnerFilterLists")
	repo.AssertExpectations(t)
}

func TestInjectPermissions_FetchesPermissionsFromAuthService(t *testing.T) {
	_, authSrv, _, svc := setupService()
	ctx := context.Background()

	// User has ID but no local permissions -- should trigger authSrv lookup
	user := &models.UserInfo{
		ID:       "owner-003",
		Username: "owneruser3",
		Roles:    []string{"OWNER"},
	}

	authSrv.On("GetUserPermissions", ctx, "owner-003").Return([]models.UserPermissionInfo{
		{DepartmentCode: "SALES", Role: "owner", IsActive: true},
	}, nil)

	result := svc.InjectPermissions(ctx, user, nil)

	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"SALES"}, depts)
	assert.Equal(t, true, result["is_restricted"])
	// Verify the user object was mutated with the fetched permissions
	assert.Len(t, user.Permissions, 1)
	assert.Equal(t, "SALES", user.Permissions[0].DepartmentCode)
	authSrv.AssertExpectations(t)
}

func TestInjectPermissions_AuthServiceError_FallsBackToRestricted(t *testing.T) {
	_, authSrv, _, svc := setupService()
	ctx := context.Background()

	user := &models.UserInfo{
		ID:       "owner-004",
		Username: "owneruser4",
		Roles:    []string{"OWNER"},
	}

	authSrv.On("GetUserPermissions", ctx, "owner-004").Return(nil, errors.New("auth service down"))

	result := svc.InjectPermissions(ctx, user, nil)

	// Auth failed, user has no permissions -> restricted
	depts, ok := result["departments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"__RESTRICTED__"}, depts)
	assert.Equal(t, true, result["is_restricted"])
	authSrv.AssertExpectations(t)
}

func TestInjectPermissions_AdminByUsername(t *testing.T) {
	_, authSrv, _, svc := setupService()
	ctx := context.Background()

	// User has no admin role but username contains "admin"
	user := &models.UserInfo{
		ID:       "user-admin-special",
		Username: "superadmin",
		Roles:    []string{"VIEWER"},
	}

	// Permissions are empty so authSrv.GetUserPermissions will be called
	authSrv.On("GetUserPermissions", ctx, "user-admin-special").Return([]models.UserPermissionInfo{}, nil)

	result := svc.InjectPermissions(ctx, user, nil)

	_, hasDepts := result["departments"]
	assert.False(t, hasDepts, "user with 'admin' in username should be treated as admin")
	_, hasRestricted := result["is_restricted"]
	assert.False(t, hasRestricted)
	authSrv.AssertExpectations(t)
}

func TestGetDashboardSummary_CapexFilterSetsYearToAll(t *testing.T) {
	repo, _, capexSrv, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()
	filter := map[string]interface{}{"year": "2025", "entities": []string{"HMW"}}

	summary := &models.DashboardSummaryDTO{
		TotalBudget: decimal.NewFromInt(100),
		TotalActual: decimal.NewFromInt(50),
	}

	repo.On("GetDashboardAggregates", ctx, mock.Anything).Return(summary, nil)
	capexSrv.On("GetCapexDashboardSummary", ctx, mock.MatchedBy(func(f map[string]interface{}) bool {
		// Capex filter should have year="All" regardless of what was passed
		return f["year"] == "All"
	})).Return(nil, errors.New("skip"))

	result, err := svc.GetDashboardSummary(ctx, user, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	capexSrv.AssertExpectations(t)
}

func TestGetOrganizationStructure_SkipsEmptyEntity(t *testing.T) {
	repo, _, _, svc := setupService()
	ctx := context.Background()
	user := newAdminUser()

	facts := []models.BudgetFactEntity{
		{Entity: "", Branch: "BKK", Department: "IT"},    // empty entity, should be skipped
		{Entity: "HMW", Branch: "BKK", Department: "IT"}, // valid
	}

	repo.On("GetOrganizationStructure", ctx).Return(facts, nil)

	result, err := svc.GetOrganizationStructure(ctx, user)

	assert.NoError(t, err)
	// Only HMW should appear (empty entity is skipped)
	assert.Len(t, result, 1)
	assert.Equal(t, "HMW", result[0].Entity)
	repo.AssertExpectations(t)
}
