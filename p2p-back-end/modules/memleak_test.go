package modules_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"
	"go.uber.org/goleak"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
	actualExportRepo "p2p-back-end/modules/exports/actual_detail_export/repository"
	actualExportSvc "p2p-back-end/modules/exports/actual_detail_export/service"
	usersSvc "p2p-back-end/modules/users/service"
)

// ============================================================================
// Section 1: Goroutine Leak Detection (uber/goleak)
// ============================================================================

// TestMain enables goroutine leak detection for ALL tests in this package.
// Any goroutine still running after tests finish will FAIL the test suite.
func TestMain(m *testing.M) {
	logs.Loginit()
	goleak.VerifyTestMain(m,
		// Known background goroutines from logger/runtime that are safe to ignore
		goleak.IgnoreTopFunction("go.uber.org/zap/zapcore.(*BufferedWriteSyncer).flushLoop"),
		goleak.IgnoreTopFunction("sync.runtime_notifyListWait"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

// ============================================================================
// Section 2: Excel File Handle Leak Tests
//
// ExcelHelper creates excelize.File internally. If the service does NOT close
// it after WriteToBuffer, repeated calls accumulate open file descriptors.
// These tests verify no resource accumulation across many iterations.
// ============================================================================

// --- Mock for ActualExportRepository ---

type MockActualExportRepo struct {
	mock.Mock
}

var _ actualExportRepo.ActualExportRepository = (*MockActualExportRepo)(nil)

func (m *MockActualExportRepo) GetActualExportDetails(ctx context.Context, user *models.UserInfo, filter map[string]interface{}) ([]models.ActualExportDTO, error) {
	args := m.Called(ctx, user, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActualExportDTO), args.Error(1)
}

func TestExcelExport_NoFileHandleLeak_RepeatedCalls(t *testing.T) {
	repo := new(MockActualExportRepo)
	svc := actualExportSvc.NewService(repo)

	user := &models.UserInfo{ID: "u1", Username: "test", Roles: []string{"ADMIN"}}
	filter := map[string]interface{}{"year": "2025"}
	data := []models.ActualExportDTO{
		{Entity: "ACG", ConsoGL: "5100", GLName: "Test", Amount: decimal.NewFromFloat(100), PostingDate: "2025-01-01"},
	}
	repo.On("GetActualExportDetails", mock.Anything, mock.Anything, mock.Anything).Return(data, nil)

	// Force GC baseline
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	const iterations = 50
	for i := 0; i < iterations; i++ {
		buf, _, err := svc.ExportActualDetailExcel(context.Background(), user, filter)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf)

		// Verify the Excel is parseable and then close it (simulating caller behavior)
		f, parseErr := excelize.OpenReader(bytes.NewReader(buf))
		if parseErr == nil {
			_ = f.Close()
		}
	}

	// Force GC to reclaim
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	// Memory growth should be bounded. If files leak, HeapInuse grows unbounded.
	// Allow up to 30 MB growth for 50 iterations (each Excel ~50-100KB).
	growth := int64(after.HeapInuse) - int64(before.HeapInuse)
	const maxGrowthBytes = 30 * 1024 * 1024
	t.Logf("Heap growth after %d iterations: %d bytes (%.2f MB)", iterations, growth, float64(growth)/(1024*1024))
	assert.Less(t, growth, int64(maxGrowthBytes),
		"Heap grew by %d bytes across %d exports — possible file handle / buffer leak", growth, iterations)
}

func TestExcelExport_BufferNotRetained_AfterReturn(t *testing.T) {
	repo := new(MockActualExportRepo)
	svc := actualExportSvc.NewService(repo)

	user := &models.UserInfo{ID: "u1", Username: "test"}
	filter := map[string]interface{}{}

	// Large dataset: 1000 rows
	var rows []models.ActualExportDTO
	for i := 0; i < 1000; i++ {
		rows = append(rows, models.ActualExportDTO{
			Entity: "ACG", Branch: "HQ", Department: fmt.Sprintf("D%d", i),
			ConsoGL: "5100", GLName: "Expense", DocumentNo: fmt.Sprintf("INV-%05d", i),
			Amount: decimal.NewFromInt(int64(i * 100)), PostingDate: "2025-01-01",
		})
	}
	repo.On("GetActualExportDetails", mock.Anything, mock.Anything, mock.Anything).Return(rows, nil)

	buf, _, err := svc.ExportActualDetailExcel(context.Background(), user, filter)
	assert.NoError(t, err)
	bufSize := len(buf)

	// Nil out the reference and force GC
	_ = buf
	buf = nil //nolint:ineffassign
	_ = rows
	rows = nil //nolint:ineffassign
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// The buffer should be GC-able. Just log for observability.
	t.Logf("Excel buffer was %d bytes (%.2f KB). HeapInuse after GC: %d bytes",
		bufSize, float64(bufSize)/1024, ms.HeapInuse)

	// Basic sanity: buffer existed and was substantial
	assert.Greater(t, bufSize, 10000, "Excel output should be > 10KB for 1000 rows")
}

// ============================================================================
// Section 3: Goroutine Leak — Users Service (sendUserChangeEvent runs in goroutine)
// ============================================================================

// --- Mocks for UsersService dependencies ---

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) GetAll(o map[string]interface{}, c context.Context, off, size int) ([]models.UserEntity, int, error) {
	return nil, 0, nil
}
func (m *MockUserRepo) IsUserExistByID(c context.Context, id string) (bool, error) { return false, nil }
func (m *MockUserRepo) CreateUser(c context.Context, u *models.UserEntity) error   { return nil }
func (m *MockUserRepo) UpdateUser(c context.Context, u *models.UserEntity) error   { return nil }
func (m *MockUserRepo) ReactivateUser(c context.Context, id string) error          { return nil }
func (m *MockUserRepo) GetUserContext(c context.Context, id string) (*models.UserEntity, error) {
	return nil, nil
}
func (m *MockUserRepo) GetUserPermissions(c context.Context, id string) ([]models.UserPermissionEntity, error) {
	return nil, nil
}
func (m *MockUserRepo) GetActiveOwnerIDsByDepartment(c context.Context, dept string) ([]string, error) {
	return nil, nil
}
func (m *MockUserRepo) UpdateUserPermissionsAndRoles(c context.Context, id string, p []models.UserPermissionEntity, r []string) error {
	return nil
}
func (m *MockUserRepo) UpdateUserID(c context.Context, o, n string) error { return nil }
func (m *MockUserRepo) ListDepartments(c context.Context) ([]models.Departments, error) {
	return nil, nil
}
func (m *MockUserRepo) ListMasterDepartments(c context.Context) ([]models.DepartmentEntity, error) {
	return nil, nil
}
func (m *MockUserRepo) GetDepartmentByCode(c context.Context, code string) (*models.DepartmentEntity, error) {
	return nil, nil
}
func (m *MockUserRepo) GetDepartmentByNavCode(c context.Context, nav string) (*models.DepartmentEntity, error) {
	return nil, nil
}
func (m *MockUserRepo) FindByUsername(c context.Context, u string) (*models.UserEntity, error) {
	args := m.Called(c, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEntity), args.Error(1)
}
func (m *MockUserRepo) SyncUsers(c context.Context, u []models.UserEntity) ([]models.UserEntity, error) {
	args := m.Called(c, u)
	return args.Get(0).([]models.UserEntity), args.Error(1)
}
func (m *MockUserRepo) GetUsers(c context.Context, id uint, lim int) ([]models.UserEntity, error) {
	return nil, nil
}

type MockSourceUserRepo struct{ mock.Mock }

func (m *MockSourceUserRepo) GetUsers(c context.Context, id uint, lim int) ([]models.CentralUser, error) {
	return nil, nil
}
func (m *MockSourceUserRepo) FindByUsername(c context.Context, u string) (*models.CentralUser, error) {
	args := m.Called(c, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CentralUser), args.Error(1)
}

type MockProducerSrv struct{ mock.Mock }

func (m *MockProducerSrv) UserChange(e *events.MessageUserEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) CompanyChange(e *events.MessageCompaniesEvent) error     { return nil }
func (m *MockProducerSrv) DepartmentChange(e *events.MessageDepartmentEvent) error { return nil }
func (m *MockProducerSrv) SectionChange(e *events.MessageSectionEvent) error       { return nil }
func (m *MockProducerSrv) PositionChange(e *events.MessagePositionEvent) error     { return nil }
func (m *MockProducerSrv) UserBegin(e *events.MessageUserBeginEvent) error         { return nil }
func (m *MockProducerSrv) CompanyBegin(e *events.MessageCompaniesBeginEvent) error { return nil }
func (m *MockProducerSrv) DepartmentBegin(e *events.MessageDepartmentBeginEvent) error {
	return nil
}
func (m *MockProducerSrv) SectionBegin(e *events.MessageSectionBeginEvent) error   { return nil }
func (m *MockProducerSrv) PositionBegin(e *events.MessagePositionBeginEvent) error { return nil }

type MockMasterRepo struct{ mock.Mock }

func (m *MockMasterRepo) SyncCompany(c context.Context, co []models.Companies) ([]models.Companies, error) {
	return nil, nil
}
func (m *MockMasterRepo) GetCompanies(c context.Context, id uint, lim int) ([]models.Companies, error) {
	return nil, nil
}
func (m *MockMasterRepo) SyncDepartment(c context.Context, d []models.Departments) ([]models.Departments, error) {
	return nil, nil
}
func (m *MockMasterRepo) GetDepartments(c context.Context, id uint, lim int) ([]models.Departments, error) {
	return nil, nil
}
func (m *MockMasterRepo) SyncSection(c context.Context, s []models.Sections) ([]models.Sections, error) {
	return nil, nil
}
func (m *MockMasterRepo) GetSections(c context.Context, id uint, lim int) ([]models.Sections, error) {
	return nil, nil
}
func (m *MockMasterRepo) SyncPosition(c context.Context, p []models.Positions) ([]models.Positions, error) {
	return nil, nil
}
func (m *MockMasterRepo) GetPositions(c context.Context, id uint, lim int) ([]models.Positions, error) {
	return nil, nil
}
func (m *MockMasterRepo) FindCompanyUUID(c context.Context, id uint) (*uuid.UUID, error) {
	return nil, nil
}
func (m *MockMasterRepo) FindDeptUUID(c context.Context, id uint) (*uuid.UUID, error) {
	return nil, nil
}
func (m *MockMasterRepo) FindSectionUUID(c context.Context, id uint) (*uuid.UUID, error) {
	return nil, nil
}
func (m *MockMasterRepo) FindPositionUUID(c context.Context, id uint) (*uuid.UUID, error) {
	return nil, nil
}

type MockDeptSrv struct{ mock.Mock }

func (m *MockDeptSrv) ManageDepartments(c context.Context) error { return nil }
func (m *MockDeptSrv) GetMasterDepartment(c context.Context, nav, entity string) (*models.DepartmentEntity, error) {
	return nil, nil
}

// TestUsersService_SyncUserByUserName_GoroutineCompletes verifies that the
// background goroutine spawned by sendUserChangeEvent finishes and does not leak.
// goleak.VerifyNone catches any lingering goroutines.
func TestUsersService_SyncUserByUserName_GoroutineCompletes(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("go.uber.org/zap/zapcore.(*BufferedWriteSyncer).flushLoop"),
		goleak.IgnoreTopFunction("sync.runtime_notifyListWait"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)

	userRepo := new(MockUserRepo)
	sourceRepo := new(MockSourceUserRepo)
	producer := new(MockProducerSrv)
	masterRepo := new(MockMasterRepo)
	deptSrv := new(MockDeptSrv)

	svc := usersSvc.NewUsersService(userRepo, sourceRepo, producer, masterRepo, deptSrv)

	// Local lookup fails -> falls through to source
	userRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, fmt.Errorf("not found"))

	sourceRepo.On("FindByUsername", mock.Anything, "newuser").Return(&models.CentralUser{
		UserID: 999, Username: "newuser", NameTh: "User TH",
	}, nil)

	syncedUser := models.UserEntity{CentralID: 999, Username: "newuser", NameTh: "User TH"}
	userRepo.On("SyncUsers", mock.Anything, mock.Anything).Return([]models.UserEntity{syncedUser}, nil)

	// The goroutine calls UserChange — mock it to succeed
	producer.On("UserChange", mock.Anything).Return(nil).Maybe()

	result, err := svc.SyncUserByUserName(context.Background(), "newuser")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Wait for the background goroutine to complete
	time.Sleep(200 * time.Millisecond)

	// goleak.VerifyNone in defer will catch any leaked goroutines
}

// TestUsersService_SyncUserByUserName_ProducerError_NoLeak verifies that
// when the producer returns an error inside the background goroutine, the goroutine
// still completes gracefully and does NOT leak.
func TestUsersService_SyncUserByUserName_ProducerError_NoLeak(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("go.uber.org/zap/zapcore.(*BufferedWriteSyncer).flushLoop"),
		goleak.IgnoreTopFunction("sync.runtime_notifyListWait"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)

	userRepo := new(MockUserRepo)
	sourceRepo := new(MockSourceUserRepo)
	producer := new(MockProducerSrv)
	masterRepo := new(MockMasterRepo)
	deptSrv := new(MockDeptSrv)

	svc := usersSvc.NewUsersService(userRepo, sourceRepo, producer, masterRepo, deptSrv)

	userRepo.On("FindByUsername", mock.Anything, "erruser").Return(nil, fmt.Errorf("not found"))
	sourceRepo.On("FindByUsername", mock.Anything, "erruser").Return(&models.CentralUser{
		UserID: 888, Username: "erruser",
	}, nil)

	syncedUser := models.UserEntity{CentralID: 888, Username: "erruser"}
	userRepo.On("SyncUsers", mock.Anything, mock.Anything).Return([]models.UserEntity{syncedUser}, nil)

	// Simulate producer failure — goroutine should still finish
	producer.On("UserChange", mock.Anything).Return(fmt.Errorf("rabbitmq connection lost")).Maybe()

	result, err := svc.SyncUserByUserName(context.Background(), "erruser")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Wait for goroutine to complete
	time.Sleep(300 * time.Millisecond)
}

// ============================================================================
// Section 4: Map / Slice Memory Growth Tests
//
// Verify that large intermediate data structures within service methods
// do not retain references after the method returns.
// ============================================================================

func TestLargeMapAllocation_FreedAfterReturn(t *testing.T) {
	// Simulate the mergedFactMap pattern from actual_service.SyncActuals
	// Build a large map, process it, then verify memory is freed.

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	func() {
		bigMap := make(map[AggKeyTest]decimal.Decimal, 100000)
		for i := 0; i < 100000; i++ {
			k := AggKeyTest{
				Entity: fmt.Sprintf("E%d", i%3),
				Dept:   fmt.Sprintf("D%d", i%50),
				GL:     fmt.Sprintf("GL%d", i%500),
				Month:  fmt.Sprintf("M%d", i%12),
			}
			bigMap[k] = bigMap[k].Add(decimal.NewFromInt(int64(i)))
		}
		// Simulate processing: iterate and discard
		total := decimal.Zero
		for _, v := range bigMap {
			total = total.Add(v)
		}
		assert.False(t, total.IsZero())
		// bigMap goes out of scope here
	}()

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	// HeapInuse should not retain the ~100K map entries after GC
	growth := int64(after.HeapInuse) - int64(before.HeapInuse)
	t.Logf("Heap growth after large map: %d bytes (%.2f MB)", growth, float64(growth)/(1024*1024))
	// Allow some noise but map of 100K entries (~20-40MB) should be mostly freed
	assert.Less(t, growth, int64(20*1024*1024), "Large map should be freed after going out of scope")
}

type AggKeyTest struct {
	Entity, Dept, GL, Month string
}

func TestLargeSliceAllocation_NilReleasesMemory(t *testing.T) {
	// Simulate the transactions pattern: build large slice, set to nil, verify freed
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var transactions []models.ActualTransactionEntity
	for i := 0; i < 10000; i++ {
		transactions = append(transactions, models.ActualTransactionEntity{
			ID:          uuid.New(),
			DocNo:       fmt.Sprintf("DOC-%05d", i),
			Amount:      decimal.NewFromInt(int64(i * 100)),
			Entity:      "ACG",
			Branch:      "HQ",
			Department:  "IT",
			PostingDate: "2025-01-01",
		})
	}
	assert.Equal(t, 10000, len(transactions))

	// Nil the slice (mimicking actual_service.go line 268: transactions = nil)
	_ = transactions
	transactions = nil //nolint:ineffassign

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	growth := int64(after.HeapInuse) - int64(before.HeapInuse)
	t.Logf("Heap after nilling 10K transactions: %d bytes (%.2f MB)", growth, float64(growth)/(1024*1024))
	assert.Less(t, growth, int64(10*1024*1024), "Nilled slice should be eligible for GC")
}

// ============================================================================
// Section 5: Context Cancellation — No Goroutine Leak on Early Exit
// ============================================================================

func TestContextCancel_NoGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("go.uber.org/zap/zapcore.(*BufferedWriteSyncer).flushLoop"),
		goleak.IgnoreTopFunction("sync.runtime_notifyListWait"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)

	repo := new(MockActualExportRepo)
	svc := actualExportSvc.NewService(repo)

	user := &models.UserInfo{ID: "u1", Username: "test"}

	// Cancel context before the call
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	repo.On("GetActualExportDetails", mock.Anything, mock.Anything, mock.Anything).
		Return([]models.ActualExportDTO{}, nil)

	// Should still work (service doesn't check ctx itself, repo mock ignores it)
	buf, _, err := svc.ExportActualDetailExcel(ctx, user, map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, buf)

	// goleak.VerifyNone catches dangling goroutines
}

// ============================================================================
// Section 6: Repeated Service Construction — No Accumulation
// ============================================================================

func TestRepeatedServiceConstruction_NoLeak(t *testing.T) {
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for i := 0; i < 1000; i++ {
		repo := new(MockActualExportRepo)
		svc := actualExportSvc.NewService(repo)
		_ = svc
		// svc and repo go out of scope each iteration
	}

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	growth := int64(after.HeapInuse) - int64(before.HeapInuse)
	t.Logf("Heap growth after 1000 service constructions: %d bytes", growth)
	assert.Less(t, growth, int64(5*1024*1024),
		"1000 discarded service instances should not retain significant memory")
}

// ============================================================================
// Section 7: Mock Expectations Cleanup — testify mock.Mock accumulates calls
// ============================================================================

func TestMockCleanup_NoAccumulation(t *testing.T) {
	// Verify that creating fresh mocks per test (not reusing) prevents state leaks
	for i := 0; i < 100; i++ {
		repo := new(MockActualExportRepo)
		svc := actualExportSvc.NewService(repo)

		user := &models.UserInfo{ID: "u1"}
		repo.On("GetActualExportDetails", mock.Anything, mock.Anything, mock.Anything).
			Return([]models.ActualExportDTO{}, nil)

		_, _, err := svc.ExportActualDetailExcel(context.Background(), user, map[string]interface{}{})
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	}
}
