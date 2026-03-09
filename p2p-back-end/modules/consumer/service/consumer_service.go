package service

import (
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

type consumerService struct {
	authService   models.AuthService
	masterService models.MasterService
}

func NewConsumerService(authService models.AuthService, masterService models.MasterService) models.ConsumerService {
	return &consumerService{
		authService:   authService,
		masterService: masterService,
	}
}

func (s *consumerService) ProcessCompanyChange(body []byte) error {
	logs.Info("📥 Received Event: Company changed. Updating local database...")
	// TODO: Implementation logic
	return nil
}

func (s *consumerService) ProcessDepartmentChange(body []byte) error {
	logs.Info("📥 Received Event: Department changed. Updating local database...")
	// TODO: Implementation logic
	return nil
}

func (s *consumerService) ProcessSectionChange(body []byte) error {
	logs.Info("📥 Received Event: Section changed. Updating local database...")
	// TODO: Implementation logic
	return nil
}

func (s *consumerService) ProcessPositionChange(body []byte) error {
	logs.Info("📥 Received Event: Position changed. Updating local database...")
	// TODO: Implementation logic
	return nil
}

func (s *consumerService) ProcessUserChange(body []byte) error {
	logs.Info("📥 Received Event: User changed. Updating local database...")
	// TODO: Implementation logic
	return nil
}
