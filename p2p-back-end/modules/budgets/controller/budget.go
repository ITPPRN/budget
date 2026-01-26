package controller

import (
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"

	"github.com/gofiber/fiber/v2"
)

type budgetController struct {
	budgetSrv models.BudgetService
}

func NewBudgetController(router fiber.Router, budgetSrv models.BudgetService) {
	controller := &budgetController{budgetSrv: budgetSrv}
	router.Post("/import-budget", controller.importBudget)
	router.Post("/import-capex-budget", controller.importCapexBudget)
	router.Post("/import-capex-actual", controller.importCapexActual)
}

func (c *budgetController) importBudget(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}
	user := ctx.Locals("user").(*models.UserInfo)

	if err := c.budgetSrv.ImportBudget(fileHeader, user.UserId); err != nil {
		logs.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(models.ResponseData{Status: "success", StatusCode: 200, Message: "Budget imported successfully"})
}

func (c *budgetController) importCapexBudget(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}
	user := ctx.Locals("user").(*models.UserInfo)

	if err := c.budgetSrv.ImportCapexBudget(fileHeader, user.UserId); err != nil {
		logs.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(models.ResponseData{Status: "success", StatusCode: 200, Message: "CAPEX Budget imported successfully"})
}

func (c *budgetController) importCapexActual(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}
	user := ctx.Locals("user").(*models.UserInfo)

	if err := c.budgetSrv.ImportCapexActual(fileHeader, user.UserId); err != nil {
		logs.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(models.ResponseData{Status: "success", StatusCode: 200, Message: "CAPEX Actual imported successfully"})
}
