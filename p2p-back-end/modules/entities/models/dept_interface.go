package models

type DepartmentService interface {
	ManageDepartments() error
	GetMasterDepartment(navCode, entity string) (*DepartmentEntity, error)
}
