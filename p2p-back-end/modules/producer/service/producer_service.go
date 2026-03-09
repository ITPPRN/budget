package service

import (
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

type producerService struct {
	eventProducer models.EvenProducer
}

func NewProducerService(eventProducer models.EvenProducer) models.ProducerService {
	return &producerService{eventProducer}
}

func (obj *producerService) UserChange(event *events.MessageUserEvent) error {
	return obj.eventProducer.Producer(event)
}

func (obj *producerService) CompanyChange(event *events.MessageCompaniesEvent) error {
	return obj.eventProducer.Producer(event)
}

func (obj *producerService) DepartmentChange(event *events.MessageDepartmentEvent) error {
	return obj.eventProducer.Producer(event)
}

func (obj *producerService) SectionChange(event *events.MessageSectionEvent) error {
	return obj.eventProducer.Producer(event)
}

func (obj *producerService) PositionChange(event *events.MessagePositionEvent) error {
	return obj.eventProducer.Producer(event)
}

func (obj *producerService) RequestCompanySync() error {
	return obj.eventProducer.Producer(&events.RequestCompanySyncEvent{})
}

func (obj *producerService) RequestDepartmentSync() error {
	return obj.eventProducer.Producer(&events.RequestDepartmentSyncEvent{})
}

func (obj *producerService) RequestUserSync() error {
	return obj.eventProducer.Producer(&events.RequestUserSyncEvent{})
}
