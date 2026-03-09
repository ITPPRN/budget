package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

// --- Mock Objects ---

type MockMasterRepo struct{ mock.Mock }

func (m *MockMasterRepo) SyncCompany(c []models.Companies) ([]models.Companies, error) {
	args := m.Called(c)
	return args.Get(0).([]models.Companies), args.Error(1)
}
func (m *MockMasterRepo) GetCompanies(lastID uint, limit int) ([]models.Companies, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.Companies), args.Error(1)
}
func (m *MockMasterRepo) SyncDepartment(d []models.Departments) ([]models.Departments, error) {
	args := m.Called(d)
	return args.Get(0).([]models.Departments), args.Error(1)
}
func (m *MockMasterRepo) GetDepartments(lastID uint, limit int) ([]models.Departments, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.Departments), args.Error(1)
}
func (m *MockMasterRepo) SyncSection(s []models.Sections) ([]models.Sections, error) {
	args := m.Called(s)
	return args.Get(0).([]models.Sections), args.Error(1)
}
func (m *MockMasterRepo) GetSections(lastID uint, limit int) ([]models.Sections, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.Sections), args.Error(1)
}
func (m *MockMasterRepo) SyncPosition(p []models.Positions) ([]models.Positions, error) {
	args := m.Called(p)
	return args.Get(0).([]models.Positions), args.Error(1)
}
func (m *MockMasterRepo) GetPositions(lastID uint, limit int) ([]models.Positions, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.Positions), args.Error(1)
}

type MockSourceRepo struct{ mock.Mock }

func (m *MockSourceRepo) GetCompanies(lastID uint, limit int) ([]models.CentralCompany, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.CentralCompany), args.Error(1)
}
func (m *MockSourceRepo) GetDepartments(lastID uint, limit int) ([]models.CentralDepartment, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.CentralDepartment), args.Error(1)
}
func (m *MockSourceRepo) GetSections(lastID uint, limit int) ([]models.CentralSection, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.CentralSection), args.Error(1)
}
func (m *MockSourceRepo) GetPositions(lastID uint, limit int) ([]models.CentralPosition, error) {
	args := m.Called(lastID, limit)
	return args.Get(0).([]models.CentralPosition), args.Error(1)
}

type MockProducerSrv struct{ mock.Mock }

func (m *MockProducerSrv) UserChange(e *events.MessageUserEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) CompanyChange(e *events.MessageCompaniesEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) DepartmentChange(e *events.MessageDepartmentEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) SectionChange(e *events.MessageSectionEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) PositionChange(e *events.MessagePositionEvent) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *MockProducerSrv) RequestCompanySync() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockProducerSrv) RequestDepartmentSync() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockProducerSrv) RequestUserSync() error {
	args := m.Called()
	return args.Error(0)
}

// --- Unit Tests ---

func TestSaveCompaniesBatch_ShouldSendEventWhenDataChanged(t *testing.T) {
	logs.Loginit()

	mockRepo := new(MockMasterRepo)
	mockSource := new(MockSourceRepo)
	mockProducer := new(MockProducerSrv)
	srv := NewMasterService(mockRepo, mockSource, mockProducer)

	inputData := []models.CentralCompany{{CompanyID: 1, Name: "Company A"}}
	changedRows := []models.Companies{{ID: 1, Name: "Company A"}}

	mockRepo.On("SyncCompany", mock.Anything).Return(changedRows, nil)
	mockProducer.On("CompanyChange", mock.Anything).Return(nil)

	err := srv.(*masterService).saveCompaniesBatch(inputData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestSaveCompaniesBatch_ShouldNotSendEventWhenNoChanges(t *testing.T) {
	logs.Loginit()

	mockRepo := new(MockMasterRepo)
	mockProducer := new(MockProducerSrv)
	srv := NewMasterService(mockRepo, nil, mockProducer)

	inputData := []models.CentralCompany{{CompanyID: 1, Name: "No Change"}}

	mockRepo.On("SyncCompany", mock.Anything).Return([]models.Companies{}, nil)

	err := srv.(*masterService).saveCompaniesBatch(inputData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertNotCalled(t, "CompanyChange", mock.Anything)
}
