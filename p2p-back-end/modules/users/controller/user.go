package usercontroller

import (
	"github.com/gofiber/fiber/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/middlewares"
)

type userController struct {
	userSrv models.UsersService
}

func NewUserController(router fiber.Router, userSrv models.UsersService) {
	controller := &userController{userSrv: userSrv}

	router.Get("user/:username", middlewares.InternalAuth(controller.FindByUsername))
}

func (h *userController) FindByUsername(c *fiber.Ctx) error {
	username := c.Params("username")

	if username == "" {
		return badReqErrResponse(c, "username is required")
	}

	res, err := h.userSrv.SyncUserByUserName(c.UserContext(), username)
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, res)
}
