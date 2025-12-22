package models

import "github.com/gofiber/fiber/v2"

type AuthService interface {
	Register(*RegisterKCReq) (string, error)
	Login(*LoginReq) (string, error)
	// IsUserExistByID(string) (bool, error)
	// ChangePassword(*ChangePasswordReq, *UserInfo) (string, error)

}

// TokenHandler is a function signature for handling JWT tokens
type TokenHandler func(c *fiber.Ctx, user *UserInfo) error

type UserRepository interface {
	IsUserExistByID(string) (bool, error)
}
