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
}

// TokenHandler is a function signature for handling JWT tokens
type TokenHandler func(c *fiber.Ctx, user *UserInfo) error

type UserRepository interface {
	GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]UserEntity, int, error)
	IsUserExistByID(string) (bool, error)
}
