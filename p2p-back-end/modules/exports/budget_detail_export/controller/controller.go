package controller

import (
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/budget_detail_export/service"
	"p2p-back-end/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

type budgetExportController struct {
	srv service.BudgetExportService
}

func NewExportController(router fiber.Router, srv service.BudgetExportService) {
	c := &budgetExportController{srv: srv}
	router.Post("/export-budget-detail", c.exportBudgetDetail)
}

func (c *budgetExportController) exportBudgetDetail(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req == nil {
		req = map[string]interface{}{}
	}
	middlewares.EnforceBranchScopeFromCtx(ctx, req)

	user := ctx.Locals("user").(*models.UserInfo)
	data, filename, err := c.srv.ExportBudgetDetailExcel(ctx.UserContext(), user, req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.Set("Access-Control-Expose-Headers", "Content-Disposition")

	return ctx.Status(fiber.StatusOK).Send(data)
}
