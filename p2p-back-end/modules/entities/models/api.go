package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// request

type UserRegisReq struct {
}

type RegisterKCReq struct {
	Username     string   `json:"username" validate:"required" example:"test1"`
	Password     string   `json:"password" validate:"required,min=6" example:"test1"` // #nosec G117 -- login request DTO
	Email        string   `json:"email" validate:"required,email" example:"test@example.com"`
	FirstName    string   `json:"first_name" validate:"required" example:"test1"`
	LastName     string   `json:"last_name" validate:"required" example:"test1"`
	Roles        []string `json:"roles" example:"[\"employee\", \"manager\"]"`
	DepartmentID string   `json:"department_id" example:"uuid-of-department"`
}

type LoginReq struct {
	Username string `json:"username" validate:"required" example:"test1"`
	Password string `json:"password" validate:"required" example:"test1"` // #nosec G117 -- login request DTO
}

type ChangePasswordReq struct {
	OldPassword     string `json:"old_password" validate:"required" example:"old_secret123"`
	NewPassword     string `json:"new_password" validate:"required,min=6" example:"new_secret123"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword" example:"new_secret123"`
}

type AdminResetPasswordReq struct {
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// res/////////////////////////////////////////////////////////
type UserInfo struct {
	ID               string               `json:"id"`
	Username         string               `json:"username"`
	Name             string               `json:"name"`
	NameTh           string               `json:"name_th"`
	Email            string               `json:"email"`
	Roles            []string             `json:"roles,omitempty"` // System-level roles (Keycloak)
	Company          string               `json:"company,omitempty"`
	CompanyID        *uuid.UUID           `json:"company_id,omitempty"` // Internal: drives BRANCH_DELEGATE scope
	Branch           string               `json:"branch,omitempty"`
	BranchCodes      []string             `json:"branch_codes,omitempty"` // Resolved via company_branch_code_mappings (1 company → many codes)
	Department       string               `json:"department,omitempty"`
	DepartmentCode   string               `json:"department_code,omitempty"`
	MappedDepartment string               `json:"mapped_department,omitempty"`
	Permissions      []UserPermissionInfo `json:"permissions,omitempty"` // Explicit Dept Permissions
}

type UserPermissionInfo struct {
	DepartmentCode string `json:"department_code"`
	Role           string `json:"role"`
	IsActive       bool   `json:"is_active"`
}

type UpdatePermissionsReq struct {
	Permissions []UserPermissionInfo `json:"permissions"`
	Roles       []string             `json:"roles"`
}

type ResponseError struct {
	Message    string `json:"message"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

type ResponseData struct {
	Message    string      `json:"message"`
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

type UserRes struct {
}

type AuditRejectBasketReq struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	UserID        uuid.UUID `json:"user_id"`
}
type AuditRejectBasketRes struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	UserID        uuid.UUID `json:"user_id"`
}


// --- เพิ่ม Struct ตัวแทนไว้ตรงนี้ ---

// rawTransactionRow มีเฉพาะคอลัมน์ใน DB (ไม่มี Source)
type RawTransactionRow struct {
	PostingDate   string          `gorm:"column:posting_date"`
	DocNo         string          `gorm:"column:doc_no"`
	Description   string          `gorm:"column:description"`
	EntityGL      string          `gorm:"column:entity_gl"`
	GLAccountName string          `gorm:"column:gl_account_name"`
	Department    string          `gorm:"column:department"`
	Amount        decimal.Decimal `gorm:"column:amount"`
	Vendor        string          `gorm:"column:vendor"`
	Company       string          `gorm:"column:company"`
	Branch        string          `gorm:"column:branch"`
}

// streamRawRow สำหรับทำ Pagination ของ Stream
type StreamRawRow struct {
	ID int64 `gorm:"column:id"`
	RawTransactionRow
}
// ---------------------------------