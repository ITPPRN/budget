package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

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

// --- Mocks ---

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
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockUserRepository) ListMasterDepartments(ctx context.Context) ([]models.DepartmentEntity, error) {
	args := m.Called(ctx)
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
	return args.Get(0).([]models.UserEntity), args.Error(1)
}

type MockSourceUserRepository struct {
	mock.Mock
}

func (m *MockSourceUserRepository) GetUsers(ctx context.Context, lastID uint, limit int) ([]models.CentralUser, error) {
	args := m.Called(ctx, lastID, limit)
	return args.Get(0).([]models.CentralUser), args.Error(1)
}

func (m *MockSourceUserRepository) FindByUsername(ctx context.Context, username string) (*models.CentralUser, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CentralUser), args.Error(1)
}

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

type MockMasterRepository struct {
	mock.Mock
}

func (m *MockMasterRepository) SyncCompany(ctx context.Context, companies []models.Companies) ([]models.Companies, error) {
	args := m.Called(ctx, companies)
	return args.Get(0).([]models.Companies), args.Error(1)
}

func (m *MockMasterRepository) GetCompanies(ctx context.Context, lastID uint, limit int) ([]models.Companies, error) {
	args := m.Called(ctx, lastID, limit)
	return args.Get(0).([]models.Companies), args.Error(1)
}

func (m *MockMasterRepository) SyncDepartment(ctx context.Context, departments []models.Departments) ([]models.Departments, error) {
	args := m.Called(ctx, departments)
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockMasterRepository) GetDepartments(ctx context.Context, lastID uint, limit int) ([]models.Departments, error) {
	args := m.Called(ctx, lastID, limit)
	return args.Get(0).([]models.Departments), args.Error(1)
}

func (m *MockMasterRepository) SyncSection(ctx context.Context, sections []models.Sections) ([]models.Sections, error) {
	args := m.Called(ctx, sections)
	return args.Get(0).([]models.Sections), args.Error(1)
}

func (m *MockMasterRepository) GetSections(ctx context.Context, lastID uint, limit int) ([]models.Sections, error) {
	args := m.Called(ctx, lastID, limit)
	return args.Get(0).([]models.Sections), args.Error(1)
}

func (m *MockMasterRepository) SyncPosition(ctx context.Context, positions []models.Positions) ([]models.Positions, error) {
	args := m.Called(ctx, positions)
	return args.Get(0).([]models.Positions), args.Error(1)
}

func (m *MockMasterRepository) GetPositions(ctx context.Context, lastID uint, limit int) ([]models.Positions, error) {
	args := m.Called(ctx, lastID, limit)
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

// --- Tests ---

func TestNewUsersService(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	assert.NotNil(t, svc)
}

func TestSyncUserByUserName_ExistsLocally(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	localUser := &models.UserEntity{
		ID:       "user-123",
		Username: "john.doe",
		NameTh:   "จอห์น",
		NameEn:   "John Doe",
		IsActive: true,
	}

	userRepo.On("FindByUsername", mock.Anything, "john.doe").Return(localUser, nil)

	result, err := svc.SyncUserByUserName(context.Background(), "john.doe")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "john.doe", result.Username)
	userRepo.AssertExpectations(t)
	sourceUserRepo.AssertNotCalled(t, "FindByUsername")
}

func TestSyncUserByUserName_FromSource(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	// Local lookup fails
	userRepo.On("FindByUsername", mock.Anything, "jane.doe").Return(nil, errors.New("not found"))

	// Source lookup succeeds
	sourceUser := &models.CentralUser{
		UserID:       10,
		Username:     "jane.doe",
		NameTh:       "เจน",
		NameEn:       "Jane Doe",
		CompanyID:    1,
		DepartmentID: 2,
		SectionID:    3,
		PositionID:   4,
	}
	sourceUserRepo.On("FindByUsername", mock.Anything, "jane.doe").Return(sourceUser, nil)

	// UUID resolution
	companyUUID := uuid.New()
	deptUUID := uuid.New()
	sectionUUID := uuid.New()
	positionUUID := uuid.New()
	masterRepo.On("FindCompanyUUID", mock.Anything, uint(1)).Return(&companyUUID, nil)
	masterRepo.On("FindDeptUUID", mock.Anything, uint(2)).Return(&deptUUID, nil)
	masterRepo.On("FindSectionUUID", mock.Anything, uint(3)).Return(&sectionUUID, nil)
	masterRepo.On("FindPositionUUID", mock.Anything, uint(4)).Return(&positionUUID, nil)

	// SyncUsers returns the synced user
	syncedUser := models.UserEntity{
		ID:           "user-456",
		CentralID:    10,
		Username:     "jane.doe",
		NameTh:       "เจน",
		NameEn:       "Jane Doe",
		CompanyID:    &companyUUID,
		DepartmentID: &deptUUID,
		SectionID:    &sectionUUID,
		PositionID:   &positionUUID,
	}
	userRepo.On("SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity")).Return([]models.UserEntity{syncedUser}, nil)

	// Producer may be called asynchronously via goroutine
	producerSrv.On("UserChange", mock.Anything).Return(nil).Maybe()

	result, err := svc.SyncUserByUserName(context.Background(), "jane.doe")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "jane.doe", result.Username)
	userRepo.AssertCalled(t, "SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity"))

	// Wait for the goroutine to complete
	time.Sleep(100 * time.Millisecond)
}

func TestSyncUserByUserName_NotFoundAnywhere(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	userRepo.On("FindByUsername", mock.Anything, "unknown").Return(nil, errors.New("not found"))
	sourceUserRepo.On("FindByUsername", mock.Anything, "unknown").Return(nil, errors.New("not found in source"))

	result, err := svc.SyncUserByUserName(context.Background(), "unknown")

	assert.Error(t, err)
	assert.Nil(t, result)
	userRepo.AssertExpectations(t)
	sourceUserRepo.AssertExpectations(t)
}

func TestSyncUsersFromEvent_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	userEvents := []events.UserEvent{
		{
			ID:       1,
			Username: "user1",
			NameTh:   "ผู้ใช้1",
			NameEn:   "User One",
		},
		{
			ID:       2,
			Username: "user2",
			NameTh:   "ผู้ใช้2",
			NameEn:   "User Two",
		},
	}

	userRepo.On("SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity")).Return([]models.UserEntity{}, nil)

	err := svc.SyncUsersFromEvent(context.Background(), userEvents)

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity"))
}

func TestSyncUsersFromEvent_Error(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	userEvents := []events.UserEvent{
		{
			ID:       1,
			Username: "user1",
			NameTh:   "ผู้ใช้1",
			NameEn:   "User One",
		},
	}

	userRepo.On("SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity")).Return(nil, errors.New("db error"))

	err := svc.SyncUsersFromEvent(context.Background(), userEvents)

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func TestSyncUsersFromEvent_WithUUIDs(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	companyUUID := uuid.New()
	deptUUID := uuid.New()
	sectionUUID := uuid.New()
	positionUUID := uuid.New()

	userEvents := []events.UserEvent{
		{
			ID:           100,
			Username:     "uuid_user",
			NameTh:       "ยูยูไอดี",
			NameEn:       "UUID User",
			CompanyID:    10,
			DepartmentID: 20,
			SectionID:    30,
			PositionID:   40,
		},
	}

	masterRepo.On("FindCompanyUUID", mock.Anything, uint(10)).Return(&companyUUID, nil)
	masterRepo.On("FindDeptUUID", mock.Anything, uint(20)).Return(&deptUUID, nil)
	masterRepo.On("FindSectionUUID", mock.Anything, uint(30)).Return(&sectionUUID, nil)
	masterRepo.On("FindPositionUUID", mock.Anything, uint(40)).Return(&positionUUID, nil)

	userRepo.On("SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity")).
		Run(func(args mock.Arguments) {
			users := args.Get(1).([]models.UserEntity)
			assert.Len(t, users, 1)
			assert.Equal(t, &companyUUID, users[0].CompanyID)
			assert.Equal(t, &deptUUID, users[0].DepartmentID)
			assert.Equal(t, &sectionUUID, users[0].SectionID)
			assert.Equal(t, &positionUUID, users[0].PositionID)
		}).
		Return([]models.UserEntity{}, nil)

	err := svc.SyncUsersFromEvent(context.Background(), userEvents)

	assert.NoError(t, err)
	masterRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestBroadcastAllLocalUsers_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	batch1 := []models.UserEntity{
		{
			ID:        "u1",
			CentralID: 1,
			Username:  "user1",
			NameTh:    "ผู้ใช้1",
			NameEn:    "User One",
		},
		{
			ID:        "u2",
			CentralID: 2,
			Username:  "user2",
			NameTh:    "ผู้ใช้2",
			NameEn:    "User Two",
		},
	}
	batch2 := []models.UserEntity{}

	// First call returns data, second call returns empty (signals end of batching)
	userRepo.On("GetUsers", mock.Anything, uint(0), 1000).Return(batch1, nil)
	userRepo.On("GetUsers", mock.Anything, uint(2), 1000).Return(batch2, nil)

	producerSrv.On("UserChange", mock.Anything).Return(nil)

	err := svc.BroadcastAllLocalUsers(context.Background())

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "GetUsers", mock.Anything, uint(0), 1000)
	producerSrv.AssertCalled(t, "UserChange", mock.Anything)
}

func TestSyncAllUsersData_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sourceUserRepo := new(MockSourceUserRepository)
	producerSrv := new(MockProducerService)
	masterRepo := new(MockMasterRepository)
	deptSrv := new(MockDepartmentService)

	svc := NewUsersService(userRepo, sourceUserRepo, producerSrv, masterRepo, deptSrv)

	centralUsers := []models.CentralUser{
		{
			UserID:       1,
			Username:     "central_user1",
			NameTh:       "เซ็นทรัล1",
			NameEn:       "Central One",
			CompanyID:    5,
			DepartmentID: 6,
		},
	}
	emptyBatch := []models.CentralUser{}

	companyUUID := uuid.New()
	deptUUID := uuid.New()

	// First call returns data, second returns empty
	sourceUserRepo.On("GetUsers", mock.Anything, uint(0), 1000).Return(centralUsers, nil)
	sourceUserRepo.On("GetUsers", mock.Anything, uint(1), 1000).Return(emptyBatch, nil)

	masterRepo.On("FindCompanyUUID", mock.Anything, uint(5)).Return(&companyUUID, nil)
	masterRepo.On("FindDeptUUID", mock.Anything, uint(6)).Return(&deptUUID, nil)

	syncedUsers := []models.UserEntity{
		{
			ID:           "synced-1",
			CentralID:    1,
			Username:     "central_user1",
			CompanyID:    &companyUUID,
			DepartmentID: &deptUUID,
		},
	}
	userRepo.On("SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity")).Return(syncedUsers, nil)

	// Producer called because SyncUsers returned changed rows
	producerSrv.On("UserChange", mock.Anything).Return(nil)

	err := svc.SyncAllUsersData(context.Background())

	assert.NoError(t, err)
	sourceUserRepo.AssertCalled(t, "GetUsers", mock.Anything, uint(0), 1000)
	userRepo.AssertCalled(t, "SyncUsers", mock.Anything, mock.AnythingOfType("[]models.UserEntity"))
}
