package models

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gofiber/fiber/v2"
)

type AuthService interface {
	Register(*RegisterKCReq) (string, error)
	//Login(*LoginReq) (string, error)
	Login(req *LoginReq) (*gocloak.JWT, error)
	RefreshToken(refreshToken string) (*gocloak.JWT, error)
	ChangePassword(oldPassword, newPassword string, userInfo *UserInfo) error
	AdminResetUserPassword(targetUserID string, newPassword string) error
	GetUserProfile(userID string) (*UserInfo, error) // Returns User + All assigned Permissions

	// User Management
	ListUsersForAdmin(optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	ListUsersForManagement(optional map[string]interface{}, page, size int) ([]UserInfo, int, error)
	GetUserPermissions(userID string) ([]UserPermissionInfo, error)
	UpdateUserPermissions(userID string, perms []UserPermissionInfo) error
	ListDepartments(user *UserInfo) ([]DepartmentEntity, error)
}

// TokenHandler is a function signature for handling JWT tokens
type TokenHandler func(c *fiber.Ctx, user *UserInfo) error

type UserRepository interface {
	GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]UserEntity, int, error)
	IsUserExistByID(string) (bool, error)
	CreateUser(user *UserEntity) error
	UpdateUser(user *UserEntity) error
	GetUserContext(userID string) (*UserEntity, error)

	// New: Explicit Permission Management
	GetUserPermissions(userID string) ([]UserPermissionEntity, error)
	SetUserPermissions(userID string, permissions []UserPermissionEntity) error
	ListDepartments() ([]DepartmentEntity, error)
	GetDepartmentByCode(code string) (*DepartmentEntity, error)
	GetDepartmentByNavCode(navCode string) (*DepartmentEntity, error)
}
