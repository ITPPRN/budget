package utils

import (
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

// --- Company Mappings ---

func SourceCompaniesToCompanies(src *models.CentralCompany) *models.Companies {
	return &models.Companies{
		ID:           src.CompanyID,
		Name:         src.Name,
		BranchName:   src.BranchName,
		BranchNameEn: src.BranchNameEn,
		BranchNo:     src.BranchNo,
		Address:      src.Address,
		TaxID:        src.TaxID,
		Province:     src.Province,
	}
}

func CompanyToCompanyChangeEvent(src *models.Companies) *events.CompanyEvent {
	return &events.CompanyEvent{
		ID:           src.ID,
		Name:         src.Name,
		BranchName:   src.BranchName,
		BranchNameEn: src.BranchNameEn,
		BranchNo:     src.BranchNo,
		Address:      src.Address,
		TaxID:        src.TaxID,
		Province:     src.Province,
	}
}

// --- Department Mappings ---

func SourceDepartmentsToDepartments(src *models.CentralDepartment) *models.Departments {
	return &models.Departments{
		ID:   src.DeptID,
		Name: src.Name,
		Code: src.Code,
	}
}

func DepartmentToDepartmentChangeEvent(src *models.Departments) *events.DepartmentEvent {
	return &events.DepartmentEvent{
		ID:   src.ID,
		Name: src.Name,
		Code: src.Code,
	}
}

// --- Section Mappings ---

func SourceSectionsToSections(src *models.CentralSection) *models.Sections {
	return &models.Sections{
		ID:           src.SectionID,
		Name:         src.Name,
		Code:         src.Code,
		DepartmentID: src.DepartmentID,
	}
}

func SectionToSectionsChangeEvent(src *models.Sections) *events.SectionEvent {
	return &events.SectionEvent{
		ID:           src.ID,
		Name:         src.Name,
		Code:         src.Code,
		DepartmentID: src.DepartmentID,
	}
}

// --- Position Mappings ---

func SourcePositionsToPositions(src *models.CentralPosition) *models.Positions {
	return &models.Positions{
		ID:   src.PositionID,
		Name: src.Name,
		Code: src.Code,
	}
}

// Support the typo in senior's code if necessary, or I'll just fix the call site
func SourcePositionsToPositinos(src *models.CentralPosition) *models.Positions {
	return SourcePositionsToPositions(src)
}

func PositionToPositionChangeEvent(src *models.Positions) *events.PositionEvent {
	return &events.PositionEvent{
		ID:   src.ID,
		Name: src.Name,
		Code: src.Code,
	}
}
