package utils

import (
	"encoding/json"

	"github.com/google/uuid"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

// --- Company Mappings ---

func SourceCompaniesToCompanies(src *models.CentralCompany) *models.Companies {
	return &models.Companies{
		ID:           uuid.New(),
		CentralID:    src.CompanyID,
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
		ID:           src.CentralID,
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
		ID:        uuid.New(),
		CentralID: src.DeptID,
		Name:      src.Name,
		Code:      src.Code,
	}
}

func DepartmentToDepartmentChangeEvent(src *models.Departments) *events.DepartmentEvent {
	return &events.DepartmentEvent{
		ID:   src.CentralID,
		Name: src.Name,
		Code: src.Code,
	}
}

// --- Section Mappings ---

func SourceSectionsToSections(src *models.CentralSection) *models.Sections {
	return &models.Sections{
		ID:        uuid.New(),
		CentralID: src.SectionID,
		Name:      src.Name,
		Code:      src.Code,
	}
}

func SectionToSectionsChangeEvent(src *models.Sections) *events.SectionEvent {
	return &events.SectionEvent{
		ID:   src.CentralID,
		Name: src.Name,
		Code: src.Code,
	}
}

// --- Position Mappings ---

func SourcePositionsToPositions(src *models.CentralPosition) *models.Positions {
	return &models.Positions{
		ID:        uuid.New(),
		CentralID: src.PositionID,
		Name:      src.Name,
		Code:      src.Code,
	}
}

func SourcePositionsToPositinos(src *models.CentralPosition) *models.Positions {
	return SourcePositionsToPositions(src)
}

func PositionToPositionChangeEvent(src *models.Positions) *events.PositionEvent {
	return &events.PositionEvent{
		ID:   src.CentralID,
		Name: src.Name,
		Code: src.Code,
	}
}

// --- User Mappings ---

func SourceUserToUser(src *models.CentralUser) *models.UserEntity {
	return &models.UserEntity{
		ID:        uuid.New().String(),
		CentralID: src.UserID,
		Username:  src.Username,
		NameTh:    src.NameTh,
		NameEn:    src.NameEn,
	}
}

func UsersToUsersResponse(src *models.UserEntity) *models.UserResponse {
	var roles []string
	if src.Roles != nil {
		_ = json.Unmarshal(src.Roles, &roles)
	}

	var companyID uint
	if src.Company != nil {
		companyID = src.Company.CentralID
	}
	var departmentID uint
	if src.Department != nil {
		departmentID = src.Department.CentralID
	}
	var sectionID uint
	if src.Section != nil {
		sectionID = src.Section.CentralID
	}
	var positionID uint
	if src.Position != nil {
		positionID = src.Position.CentralID
	}

	return &models.UserResponse{
		ID:           src.ID,
		Username:     src.Username,
		NameTh:       src.NameTh,
		NameEn:       src.NameEn,
		CompanyID:    companyID,
		DepartmentID: departmentID,
		SectionID:    sectionID,
		PositionID:   positionID,
		Roles:        roles,
	}
}

func UserToUserChangeEvent(src *models.UserEntity) *events.UserEvent {
	var companyID, deptID, secID, posID uint
	if src.Company != nil {
		companyID = src.Company.CentralID
	}
	if src.Department != nil {
		deptID = src.Department.CentralID
	}
	if src.Section != nil {
		secID = src.Section.CentralID
	}
	if src.Position != nil {
		posID = src.Position.CentralID
	}

	return &events.UserEvent{
		ID:           src.CentralID,
		Username:     src.Username,
		NameTh:       src.NameTh,
		NameEn:       src.NameEn,
		CompanyID:    companyID,
		DepartmentID: deptID,
		SectionID:    secID,
		PositionID:   posID,
	}
}

// --- Event to Entity Mappings (Reverse) ---

func EventUserToUsers(src *events.UserEvent) *models.UserEntity {
	return &models.UserEntity{
		ID:        uuid.New().String(),
		CentralID: src.ID,
		Username:  src.Username,
		NameTh:    src.NameTh,
		NameEn:    src.NameEn,
	}
}

func EventCompanyToCompanies(src *events.CompanyEvent) *models.Companies {
	return &models.Companies{
		ID:           uuid.New(),
		CentralID:    src.ID,
		Name:         src.Name,
		BranchName:   src.BranchName,
		BranchNameEn: src.BranchNameEn,
		BranchNo:     src.BranchNo,
		Address:      src.Address,
		TaxID:        src.TaxID,
		Province:     src.Province,
	}
}

func EventDepartmentToDepartments(src *events.DepartmentEvent) *models.Departments {
	return &models.Departments{
		ID:        uuid.New(),
		CentralID: src.ID,
		Name:      src.Name,
		Code:      src.Code,
	}
}

func EventSectionToSections(src *events.SectionEvent) *models.Sections {
	return &models.Sections{
		ID:        uuid.New(),
		CentralID: src.ID,
		Name:      src.Name,
		Code:      src.Code,
	}
}

func EventPositionToPositions(src *events.PositionEvent) *models.Positions {
	return &models.Positions{
		ID:        uuid.New(),
		CentralID: src.ID,
		Name:      src.Name,
		Code:      src.Code,
	}
}
