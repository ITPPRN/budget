package models

import (
	"time"

	"gorm.io/gorm"
)

// --- Local Master Entities ---

type Companies struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `json:"name"`
	BranchName   string         `json:"branch_name"`
	BranchNameEn string         `json:"branch_name_en"`
	BranchNo     string         `json:"branch_no"`
	Address      string         `json:"address"`
	TaxID        string         `gorm:"column:taxid" json:"taxid"`
	Province     string         `json:"province"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Departments struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Sections struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	DepartmentID uint           `json:"department_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Positions struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// --- Source (Central) Entities ---

type CentralCompany struct {
	CompanyID    uint   `gorm:"column:company_id;primaryKey"`
	Name         string `gorm:"column:name"`
	BranchName   string `gorm:"column:branch_name"`
	BranchNameEn string `gorm:"column:branch_name_en"`
	BranchNo     string `gorm:"column:branch_no"`
	Address      string `gorm:"column:address"`
	TaxID        string `gorm:"column:taxid"`
	Province     string `gorm:"column:province"`
}

type CentralDepartment struct {
	DeptID uint   `gorm:"column:department_id;primaryKey"`
	Name   string `gorm:"column:name"`
	Code   string `gorm:"column:code"`
}

type CentralSection struct {
	SectionID    uint   `gorm:"column:section_id;primaryKey"`
	Name         string `gorm:"column:name"`
	Code         string `gorm:"column:code"`
	DepartmentID uint   `gorm:"column:department_id"`
}

type CentralPosition struct {
	PositionID uint   `gorm:"column:position_id;primaryKey"`
	Name       string `gorm:"column:name"`
	Code       string `gorm:"column:code"`
}
