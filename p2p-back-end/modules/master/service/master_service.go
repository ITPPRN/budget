package service

import (
	"fmt"
	"sync"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type masterService struct {
	masterRepo       models.MasterRepository
	sourceMasterRepo models.SourceMasterRepository
	producerSrv      models.ProducerService
}

func NewMasterService(
	masterRepo models.MasterRepository,
	sourceMasterRepo models.SourceMasterRepository,
	producerSrv models.ProducerService,
) models.MasterService {
	return &masterService{
		masterRepo:       masterRepo,
		sourceMasterRepo: sourceMasterRepo,
		producerSrv:      producerSrv,
	}
}

func (s *masterService) SyncAllData() {
	maxConcurrent := 4
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	jobs := []struct {
		Name string
		Func func() error
	}{
		{"Companies Table", s.SyncAllCompaniesData},
		{"Departments Table", s.SyncAllDepartmentData},
		{"Sections Table", s.SyncAllSectionData},
		{"Positions Table", s.SyncAllPositionData},
	}

	logs.Info("🚀 Starting Sync Data...")

	for _, job := range jobs {
		wg.Add(1)
		go func(jName string, jFunc func() error) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					errChan <- fmt.Errorf("💥 Panic in %s: %v", jName, r)
				}
			}()

			logs.Info(fmt.Sprintf("⏳ Syncing %s...", jName))
			if err := jFunc(); err != nil {
				errChan <- fmt.Errorf("❌ Failed %s: %v", jName, err)
			} else {
				logs.Info(fmt.Sprintf("✅ Success %s", jName))
			}
		}(job.Name, job.Func)
	}

	wg.Wait()
	close(errChan)

	errorCount := 0
	for err := range errChan {
		logs.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		logs.Warn(fmt.Sprintf("⚠️ Sync completed with %d errors", errorCount))
	} else {
		logs.Info("✨ All tables synced successfully!")
	}
}

func (s *masterService) BroadcastAllData() {
	logs.Info("📢 Broadcasting all local master data...")

	if err := s.BroadcastAllLocalCompanies(); err != nil {
		logs.Warnf("Failed to broadcast companies: %v", err)
	}
	if err := s.BroadcastAllLocalDepartments(); err != nil {
		logs.Warnf("Failed to broadcast departments: %v", err)
	}
	if err := s.BroadcastAllLocalSections(); err != nil {
		logs.Warnf("Failed to broadcast sections: %v", err)
	}
	if err := s.BroadcastAllLocalPositions(); err != nil {
		logs.Warnf("Failed to broadcast positions: %v", err)
	}
}

// --- Companies ---

func (s *masterService) SyncAllCompaniesData() error {
	return utils.BatchSync[models.CentralCompany, uint](
		0, 1000,
		s.sourceMasterRepo.GetCompanies,
		s.saveCompaniesBatch,
		func(item models.CentralCompany) uint { return item.CompanyID },
	)
}

func (s *masterService) saveCompaniesBatch(data []models.CentralCompany) error {
	var companies []models.Companies
	for _, company := range data {
		com := company
		companies = append(companies, *utils.SourceCompaniesToCompanies(&com))
	}

	changedRows, err := s.masterRepo.SyncCompany(companies)
	if err != nil {
		logs.Warn(err.Error())
		return err
	}

	if len(changedRows) > 0 {
		var companiesEvent []events.CompanyEvent
		for _, row := range changedRows {
			companiesEvent = append(companiesEvent, *utils.CompanyToCompanyChangeEvent(&row))
		}
		syncEvent := &events.MessageCompaniesEvent{Companies: companiesEvent}
		if s.producerSrv != nil {
			if err := s.producerSrv.CompanyChange(syncEvent); err != nil {
				logs.Warnf("⚠️ Producer Error: %v", err)
			}
		}
	}
	return nil
}

func (s *masterService) BroadcastAllLocalCompanies() error {
	return utils.BatchSync(
		0, 1000,
		s.masterRepo.GetCompanies,
		func(data []models.Companies) error {
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
		func(item models.Companies) uint { return item.ID },
	)
}

// --- Departments ---

func (s *masterService) SyncAllDepartmentData() error {
	return utils.BatchSync(
		0, 1000,
		s.sourceMasterRepo.GetDepartments,
		s.saveDepartmentBatch,
		func(item models.CentralDepartment) uint { return item.DeptID },
	)
}

func (s *masterService) saveDepartmentBatch(data []models.CentralDepartment) error {
	var departments []models.Departments
	for _, department := range data {
		dep := department
		departments = append(departments, *utils.SourceDepartmentsToDepartments(&dep))
	}

	changedRows, err := s.masterRepo.SyncDepartment(departments)
	if err != nil {
		logs.Warn(err.Error())
		return err
	}

	if len(changedRows) > 0 {
		var departmentsEvent []events.DepartmentEvent
		for _, row := range changedRows {
			departmentsEvent = append(departmentsEvent, *utils.DepartmentToDepartmentChangeEvent(&row))
		}
		syncEvent := &events.MessageDepartmentEvent{Departments: departmentsEvent}
		if s.producerSrv != nil {
			if err := s.producerSrv.DepartmentChange(syncEvent); err != nil {
				logs.Warnf("⚠️ Producer Error: %v", err)
			}
		}
	}
	return nil
}

func (s *masterService) BroadcastAllLocalDepartments() error {
	return utils.BatchSync(
		0, 1000,
		s.masterRepo.GetDepartments,
		func(data []models.Departments) error {
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
		func(item models.Departments) uint { return item.ID },
	)
}

// --- Sections ---

func (s *masterService) SyncAllSectionData() error {
	return utils.BatchSync(
		0, 1000,
		s.sourceMasterRepo.GetSections,
		s.saveSectionsBatch,
		func(item models.CentralSection) uint { return item.SectionID },
	)
}

func (s *masterService) saveSectionsBatch(data []models.CentralSection) error {
	var sections []models.Sections
	for _, section := range data {
		sec := section
		sections = append(sections, *utils.SourceSectionsToSections(&sec))
	}

	changedRows, err := s.masterRepo.SyncSection(sections)
	if err != nil {
		logs.Warn(err.Error())
		return err
	}

	if len(changedRows) > 0 {
		var sectionsEvent []events.SectionEvent
		for _, row := range changedRows {
			sectionsEvent = append(sectionsEvent, *utils.SectionToSectionsChangeEvent(&row))
		}
		syncEvent := &events.MessageSectionEvent{Sections: sectionsEvent}
		if s.producerSrv != nil {
			if err := s.producerSrv.SectionChange(syncEvent); err != nil {
				logs.Warnf("⚠️ Producer Error: %v", err)
			}
		}
	}
	return nil
}

func (s *masterService) BroadcastAllLocalSections() error {
	return utils.BatchSync(
		0, 1000,
		s.masterRepo.GetSections,
		func(data []models.Sections) error {
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
		func(item models.Sections) uint { return item.ID },
	)
}

// --- Positions ---

func (s *masterService) SyncAllPositionData() error {
	return utils.BatchSync(
		0, 1000,
		s.sourceMasterRepo.GetPositions,
		s.savePositionsBatch,
		func(item models.CentralPosition) uint { return item.PositionID },
	)
}

func (s *masterService) savePositionsBatch(data []models.CentralPosition) error {
	var positions []models.Positions
	for _, position := range data {
		pos := position
		positions = append(positions, *utils.SourcePositionsToPositinos(&pos))
	}

	changedRows, err := s.masterRepo.SyncPosition(positions)
	if err != nil {
		logs.Warn(err.Error())
		return err
	}

	if len(changedRows) > 0 {
		var positionsEvent []events.PositionEvent
		for _, row := range changedRows {
			positionsEvent = append(positionsEvent, *utils.PositionToPositionChangeEvent(&row))
		}
		syncEvent := &events.MessagePositionEvent{Positions: positionsEvent}
		if s.producerSrv != nil {
			if err := s.producerSrv.PositionChange(syncEvent); err != nil {
				logs.Warnf("⚠️ Producer Error: %v", err)
			}
		}
	}
	return nil
}

func (s *masterService) BroadcastAllLocalPositions() error {
	return utils.BatchSync(
		0, 1000,
		s.masterRepo.GetPositions,
		func(data []models.Positions) error {
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
		func(item models.Positions) uint { return item.ID },
	)
}
