package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

func TestMain(m *testing.M) {
	logs.Loginit()
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Mock: MasterRepository
// ---------------------------------------------------------------------------

type MockMasterRepository struct {
	mock.Mock
}

func (m *MockMasterRepository) SyncCompany(ctx context.Context, companies []models.Companies) ([]models.Companies, error) {
	args := m.Called(ctx, companies)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Companies), args.Error(1)
}

func (m *MockMasterRepository) GetCompanies(ctx context.Context, lastID uint, limit int) ([]models.Companies, error) {
	args := m.Called(ctx, lastID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Companies), args.Error(1)
}

func (m *MockMasterRepository) SyncDepartment(ctx context.Context, departments []models.Departments) ([]models.Departments, error) {
	args := m.Called(ctx, departments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockMasterRepository) GetDepartments(ctx context.Context, lastID uint, limit int) ([]models.Departments, error) {
	args := m.Called(ctx, lastID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockMasterRepository) SyncSection(ctx context.Context, sections []models.Sections) ([]models.Sections, error) {
	args := m.Called(ctx, sections)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Sections), args.Error(1)
}

func (m *MockMasterRepository) GetSections(ctx context.Context, lastID uint, limit int) ([]models.Sections, error) {
	args := m.Called(ctx, lastID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Sections), args.Error(1)
}

func (m *MockMasterRepository) SyncPosition(ctx context.Context, positions []models.Positions) ([]models.Positions, error) {
	args := m.Called(ctx, positions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Positions), args.Error(1)
}

func (m *MockMasterRepository) GetPositions(ctx context.Context, lastID uint, limit int) ([]models.Positions, error) {
	args := m.Called(ctx, lastID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Positions), args.Error(1)
}

func (m *MockMasterRepository) FindCompanyUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	args := m.Called(ctx, centralID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockMasterRepository) FindDeptUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	args := m.Called(ctx, centralID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockMasterRepository) FindSectionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	args := m.Called(ctx, centralID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockMasterRepository) FindPositionUUID(ctx context.Context, centralID uint) (*uuid.UUID, error) {
	args := m.Called(ctx, centralID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

// ---------------------------------------------------------------------------
// Mock: ProducerService
// ---------------------------------------------------------------------------

type MockProducerService struct {
	mock.Mock
}

func (m *MockProducerService) UserChange(event *events.MessageUserEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) CompanyChange(event *events.MessageCompaniesEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) DepartmentChange(event *events.MessageDepartmentEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) SectionChange(event *events.MessageSectionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) PositionChange(event *events.MessagePositionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) UserBegin(event *events.MessageUserBeginEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) CompanyBegin(event *events.MessageCompaniesBeginEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) DepartmentBegin(event *events.MessageDepartmentBeginEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) SectionBegin(event *events.MessageSectionBeginEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockProducerService) PositionBegin(event *events.MessagePositionBeginEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

// ===========================================================================
// Tests
// ===========================================================================

// --- 1. TestNewMasterService ---

func TestNewMasterService(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)

	svc := NewMasterService(mockRepo, mockProducer)
	assert.NotNil(t, svc)

	// Verify the concrete type contains the expected fields.
	concrete, ok := svc.(*masterService)
	assert.True(t, ok)
	assert.Equal(t, mockRepo, concrete.masterRepo)
	assert.Equal(t, mockProducer, concrete.producerSrv)
}

// --- 2. TestSyncCompaniesFromEvent_Success ---

func TestSyncCompaniesFromEvent_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	companyEvents := []events.CompanyEvent{
		{ID: 1, Name: "Company A", BranchName: "HQ", BranchNameEn: "HQ EN", BranchNo: "001", Address: "123 St", TaxID: "TAX1", Province: "BKK"},
		{ID: 2, Name: "Company B", BranchName: "Branch", BranchNameEn: "Branch EN", BranchNo: "002", Address: "456 St", TaxID: "TAX2", Province: "CNX"},
	}

	mockRepo.On("SyncCompany", mock.Anything, mock.AnythingOfType("[]models.Companies")).
		Return([]models.Companies{}, nil)

	err := svc.SyncCompaniesFromEvent(context.Background(), companyEvents)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- 3. TestSyncCompaniesFromEvent_Error ---

func TestSyncCompaniesFromEvent_Error(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	companyEvents := []events.CompanyEvent{
		{ID: 1, Name: "Company A"},
	}

	expectedErr := errors.New("sync company failed")
	mockRepo.On("SyncCompany", mock.Anything, mock.AnythingOfType("[]models.Companies")).
		Return(nil, expectedErr)

	err := svc.SyncCompaniesFromEvent(context.Background(), companyEvents)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

// --- 4. TestSyncDepartmentsFromEvent_Success ---

func TestSyncDepartmentsFromEvent_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	deptEvents := []events.DepartmentEvent{
		{ID: 10, Code: "D001", Name: "Engineering"},
		{ID: 20, Code: "D002", Name: "Finance"},
	}

	mockRepo.On("SyncDepartment", mock.Anything, mock.AnythingOfType("[]models.Departments")).
		Return([]models.Departments{}, nil)

	err := svc.SyncDepartmentsFromEvent(context.Background(), deptEvents)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- 5. TestSyncDepartmentsFromEvent_Error ---

func TestSyncDepartmentsFromEvent_Error(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	deptEvents := []events.DepartmentEvent{
		{ID: 10, Code: "D001", Name: "Engineering"},
	}

	expectedErr := errors.New("sync department failed")
	mockRepo.On("SyncDepartment", mock.Anything, mock.AnythingOfType("[]models.Departments")).
		Return(nil, expectedErr)

	err := svc.SyncDepartmentsFromEvent(context.Background(), deptEvents)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

// --- 6. TestSyncSectionsFromEvent_Success ---

func TestSyncSectionsFromEvent_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	sectionEvents := []events.SectionEvent{
		{ID: 100, Name: "Section A", Code: "S001", DepartmentID: 10},
		{ID: 200, Name: "Section B", Code: "S002", DepartmentID: 20},
	}

	mockRepo.On("SyncSection", mock.Anything, mock.AnythingOfType("[]models.Sections")).
		Return([]models.Sections{}, nil)

	err := svc.SyncSectionsFromEvent(context.Background(), sectionEvents)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- 7. TestSyncPositionsFromEvent_Success ---

func TestSyncPositionsFromEvent_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	positionEvents := []events.PositionEvent{
		{ID: 1, Name: "Manager", Code: "P001"},
		{ID: 2, Name: "Director", Code: "P002"},
	}

	mockRepo.On("SyncPosition", mock.Anything, mock.AnythingOfType("[]models.Positions")).
		Return([]models.Positions{}, nil)

	err := svc.SyncPositionsFromEvent(context.Background(), positionEvents)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- 8. TestBroadcastAllLocalCompanies_Success ---

func TestBroadcastAllLocalCompanies_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	companyID := uuid.New()
	companies := []models.Companies{
		{ID: companyID, CentralID: 5, Name: "Test Co", BranchName: "HQ", BranchNameEn: "HQ EN", BranchNo: "001", Address: "Addr", TaxID: "T1", Province: "BKK"},
	}

	// Returns fewer items than limit (1000), so BatchSync stops after one call.
	mockRepo.On("GetCompanies", mock.Anything, uint(0), 1000).
		Return(companies, nil).Once()

	mockProducer.On("CompanyChange", mock.AnythingOfType("*events.MessageCompaniesEvent")).
		Return(nil).Once()

	err := svc.BroadcastAllLocalCompanies(context.Background())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// --- 9. TestBroadcastAllLocalCompanies_NilProducer ---

func TestBroadcastAllLocalCompanies_NilProducer(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	// producerSrv is nil -- should not panic.
	svc := NewMasterService(mockRepo, nil)

	companyID := uuid.New()
	companies := []models.Companies{
		{ID: companyID, CentralID: 3, Name: "NilProducer Co"},
	}

	mockRepo.On("GetCompanies", mock.Anything, uint(0), 1000).
		Return(companies, nil).Once()

	assert.NotPanics(t, func() {
		err := svc.BroadcastAllLocalCompanies(context.Background())
		assert.NoError(t, err)
	})
	mockRepo.AssertExpectations(t)
}

// --- 10. TestBroadcastAllLocalDepartments_Success ---

func TestBroadcastAllLocalDepartments_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	deptID := uuid.New()
	departments := []models.Departments{
		{ID: deptID, CentralID: 10, Name: "Engineering", Code: "ENG"},
	}

	mockRepo.On("GetDepartments", mock.Anything, uint(0), 1000).
		Return(departments, nil).Once()

	mockProducer.On("DepartmentChange", mock.AnythingOfType("*events.MessageDepartmentEvent")).
		Return(nil).Once()

	err := svc.BroadcastAllLocalDepartments(context.Background())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// --- 11. TestBroadcastAllLocalSections_Success ---

func TestBroadcastAllLocalSections_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	sectionID := uuid.New()
	deptUUID := uuid.New()
	sections := []models.Sections{
		{ID: sectionID, CentralID: 50, Name: "Section X", Code: "SX", DepartmentID: &deptUUID},
	}

	mockRepo.On("GetSections", mock.Anything, uint(0), 1000).
		Return(sections, nil).Once()

	mockProducer.On("SectionChange", mock.AnythingOfType("*events.MessageSectionEvent")).
		Return(nil).Once()

	err := svc.BroadcastAllLocalSections(context.Background())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// --- 12. TestBroadcastAllLocalPositions_Success ---

func TestBroadcastAllLocalPositions_Success(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	posID := uuid.New()
	positions := []models.Positions{
		{ID: posID, CentralID: 77, Name: "Analyst", Code: "ANL"},
	}

	mockRepo.On("GetPositions", mock.Anything, uint(0), 1000).
		Return(positions, nil).Once()

	mockProducer.On("PositionChange", mock.AnythingOfType("*events.MessagePositionEvent")).
		Return(nil).Once()

	err := svc.BroadcastAllLocalPositions(context.Background())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// --- 13. TestBroadcastAllData ---

func TestBroadcastAllData(t *testing.T) {
	mockRepo := new(MockMasterRepository)
	mockProducer := new(MockProducerService)
	svc := NewMasterService(mockRepo, mockProducer)

	// Each entity: first call returns empty slice so BatchSync finishes immediately.
	mockRepo.On("GetCompanies", mock.Anything, uint(0), 1000).
		Return([]models.Companies{}, nil).Once()
	mockRepo.On("GetDepartments", mock.Anything, uint(0), 1000).
		Return([]models.Departments{}, nil).Once()
	mockRepo.On("GetSections", mock.Anything, uint(0), 1000).
		Return([]models.Sections{}, nil).Once()
	mockRepo.On("GetPositions", mock.Anything, uint(0), 1000).
		Return([]models.Positions{}, nil).Once()

	assert.NotPanics(t, func() {
		svc.BroadcastAllData(context.Background())
	})
	mockRepo.AssertExpectations(t)
}
