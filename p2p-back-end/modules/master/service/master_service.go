package service

import (
	"context"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type masterService struct {
	masterRepo  models.MasterRepository
	producerSrv models.ProducerService
}

func NewMasterService(
	masterRepo models.MasterRepository,
	producerSrv models.ProducerService,
) models.MasterService {
	return &masterService{
		masterRepo:  masterRepo,
		producerSrv: producerSrv,
	}
}


func (s *masterService) BroadcastAllData(ctx context.Context) {
	logs.Info("📢 Broadcasting all local master data...")

	if err := s.BroadcastAllLocalCompanies(ctx); err != nil {
		logs.Warnf("Failed to broadcast companies: %v", err)
	}
	if err := s.BroadcastAllLocalDepartments(ctx); err != nil {
		logs.Warnf("Failed to broadcast departments: %v", err)
	}
	if err := s.BroadcastAllLocalSections(ctx); err != nil {
		logs.Warnf("Failed to broadcast sections: %v", err)
	}
	if err := s.BroadcastAllLocalPositions(ctx); err != nil {
		logs.Warnf("Failed to broadcast positions: %v", err)
	}
}

// --- Companies ---


func (s *masterService) BroadcastAllLocalCompanies(ctx context.Context) error {
	return utils.BatchSync(
		ctx,
		0, 1000,
		s.masterRepo.GetCompanies,
		func(ctx context.Context, data []models.Companies) error {
			var companiesEvent []events.CompanyEvent
			for _, row := range data {
				data := row
				companiesEvent = append(companiesEvent, *utils.CompanyToCompanyChangeEvent(&data))
			}
			syncEvent := &events.MessageCompaniesEvent{Companies: companiesEvent}
			if s.producerSrv != nil {
				return s.producerSrv.CompanyChange(syncEvent)
			}
			return nil
		},
		func(item models.Companies) uint { return item.CentralID },
	)
}

// --- Departments ---


func (s *masterService) BroadcastAllLocalDepartments(ctx context.Context) error {
	return utils.BatchSync(
		ctx,
		0, 1000,
		s.masterRepo.GetDepartments,
		func(ctx context.Context, data []models.Departments) error {
			var departmentsEvent []events.DepartmentEvent
			for _, row := range data {
				data := row
				departmentsEvent = append(departmentsEvent, *utils.DepartmentToDepartmentChangeEvent(&data))
			}
			syncEvent := &events.MessageDepartmentEvent{Departments: departmentsEvent}
			if s.producerSrv != nil {
				return s.producerSrv.DepartmentChange(syncEvent)
			}
			return nil
		},
		func(item models.Departments) uint { return item.CentralID },
	)
}

// --- Sections ---


func (s *masterService) BroadcastAllLocalSections(ctx context.Context) error {
	return utils.BatchSync(
		ctx,
		0, 1000,
		s.masterRepo.GetSections,
		func(ctx context.Context, data []models.Sections) error {
			var sectionsEvent []events.SectionEvent
			for _, row := range data {
				data := row
				sectionsEvent = append(sectionsEvent, *utils.SectionToSectionsChangeEvent(&data))
			}
			syncEvent := &events.MessageSectionEvent{Sections: sectionsEvent}
			if s.producerSrv != nil {
				return s.producerSrv.SectionChange(syncEvent)
			}
			return nil
		},
		func(item models.Sections) uint { return item.CentralID },
	)
}

// --- Positions ---


func (s *masterService) BroadcastAllLocalPositions(ctx context.Context) error {
	return utils.BatchSync(
		ctx,
		0, 1000,
		s.masterRepo.GetPositions,
		func(ctx context.Context, data []models.Positions) error {
			var positionsEvent []events.PositionEvent
			for _, row := range data {
				data := row
				positionsEvent = append(positionsEvent, *utils.PositionToPositionChangeEvent(&data))
			}
			syncEvent := &events.MessagePositionEvent{Positions: positionsEvent}
			if s.producerSrv != nil {
				return s.producerSrv.PositionChange(syncEvent)
			}
			return nil
		},
		func(item models.Positions) uint { return item.CentralID },
	)
}

func (s *masterService) SyncCompaniesFromEvent(ctx context.Context, companiesEvent []events.CompanyEvent) error {
	var companies []models.Companies
	for _, c := range companiesEvent {
		cd := c
		companies = append(companies, *utils.EventCompanyToCompanies(&cd))
	}
	_, err := s.masterRepo.SyncCompany(ctx, companies)
	return err
}

func (s *masterService) SyncDepartmentsFromEvent(ctx context.Context, departmentsEvent []events.DepartmentEvent) error {
	var departments []models.Departments
	for _, d := range departmentsEvent {
		dd := d
		departments = append(departments, *utils.EventDepartmentToDepartments(&dd))
	}
	_, err := s.masterRepo.SyncDepartment(ctx, departments)
	return err
}

func (s *masterService) SyncSectionsFromEvent(ctx context.Context, sectionsEvent []events.SectionEvent) error {
	var sections []models.Sections
	for _, sec := range sectionsEvent {
		sd := sec
		sections = append(sections, *utils.EventSectionToSections(&sd))
	}
	_, err := s.masterRepo.SyncSection(ctx, sections)
	return err
}

func (s *masterService) SyncPositionsFromEvent(ctx context.Context, positionsEvent []events.PositionEvent) error {
	var positions []models.Positions
	for _, pos := range positionsEvent {
		pd := pos
		positions = append(positions, *utils.EventPositionToPositions(&pd))
	}
	_, err := s.masterRepo.SyncPosition(ctx, positions)
	return err
}
