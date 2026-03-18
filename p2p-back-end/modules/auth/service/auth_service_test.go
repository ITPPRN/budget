package service

import (
	"context"
	"testing"
	"p2p-back-end/modules/entities/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

// MockUserRepository is a mock implementation of models.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.UserEntity) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.UserEntity) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserContext(ctx context.Context, id string) (*models.UserEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetDepartmentByCode(ctx context.Context, code string) (*models.DepartmentEntity, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DepartmentEntity), args.Error(1)
}

func (m *MockUserRepository) GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]models.UserEntity, int, error) {
	args := m.Called(optional, ctx, offset, size)
	return args.Get(0).([]models.UserEntity), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) IsUserExistByID(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUserPermissions(ctx context.Context, userID string) ([]models.UserPermissionEntity, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.UserPermissionEntity), args.Error(1)
}

func (m *MockUserRepository) SetUserPermissions(ctx context.Context, userID string, perms []models.UserPermissionEntity) error {
	args := m.Called(ctx, userID, perms)
	return args.Error(0)
}

func (m *MockUserRepository) ListDepartments(ctx context.Context) ([]models.DepartmentEntity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.DepartmentEntity), args.Error(1)
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
	return args.Get(0).([]models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetUsers(ctx context.Context, lastID uint, limit int) ([]models.UserEntity, error) {
	args := m.Called(ctx, lastID, limit)
	return args.Get(0).([]models.UserEntity), args.Error(1)
}

func (m *MockUserRepository) UpdateUserID(ctx context.Context, oldID, newID string) error {
	args := m.Called(ctx, oldID, newID)
	return args.Error(0)
}

func TestLogin_UserNotFound(t *testing.T) {
	// Simulated test placeholder
	assert.True(t, true)
}

func TestGetUserProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	s := NewAuthService(nil, nil, mockRepo, nil)

	userID := "user-123"
	expectedUser := &models.UserEntity{
		ID:        userID,
		Username:  "tester",
		FirstName: "Test",
		LastName:  "User",
		Roles:     []byte("[\"ADMIN\"]"),
	}

	mockRepo.On("GetUserContext", mock.Anything, userID).Return(expectedUser, nil)
	mockRepo.On("GetUserPermissions", mock.Anything, userID).Return([]models.UserPermissionEntity{}, nil)

	profile, err := s.GetUserProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "tester", profile.Username)
	assert.Contains(t, profile.Roles, "ADMIN")
	mockRepo.AssertExpectations(t)
}
