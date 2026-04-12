package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

// MockEvenProducer implements models.EvenProducer
type MockEvenProducer struct {
	mock.Mock
}

func (m *MockEvenProducer) Producer(event events.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestNewProducerService(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)
	assert.NotNil(t, svc)
	assert.Implements(t, (*models.ProducerService)(nil), svc)
}

func TestUserChange_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageUserEvent{
		Users: []events.UserEvent{{ID: 1, Username: "testuser"}},
	}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.UserChange(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestUserChange_Error(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageUserEvent{
		Users: []events.UserEvent{{ID: 1, Username: "testuser"}},
	}
	expectedErr := errors.New("producer error")
	mockProducer.On("Producer", event).Return(expectedErr)

	err := svc.UserChange(event)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockProducer.AssertExpectations(t)
}

func TestCompanyChange_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageCompaniesEvent{
		Companies: []events.CompanyEvent{{ID: 1, Name: "TestCorp"}},
	}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.CompanyChange(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestDepartmentChange_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageDepartmentEvent{
		Departments: []events.DepartmentEvent{{ID: 1, Code: "D001", Name: "Engineering"}},
	}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.DepartmentChange(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestSectionChange_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageSectionEvent{
		Sections: []events.SectionEvent{{ID: 1, Name: "Backend", Code: "S001", DepartmentID: 1}},
	}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.SectionChange(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestPositionChange_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessagePositionEvent{
		Positions: []events.PositionEvent{{ID: 1, Name: "Developer", Code: "P001"}},
	}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.PositionChange(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestUserBegin_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageUserBeginEvent{}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.UserBegin(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestCompanyBegin_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageCompaniesBeginEvent{}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.CompanyBegin(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestDepartmentBegin_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageDepartmentBeginEvent{}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.DepartmentBegin(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestSectionBegin_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessageSectionBeginEvent{}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.SectionBegin(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestPositionBegin_Success(t *testing.T) {
	mockProducer := new(MockEvenProducer)
	svc := NewProducerService(mockProducer)

	event := &events.MessagePositionBeginEvent{}
	mockProducer.On("Producer", event).Return(nil)

	err := svc.PositionBegin(event)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}
