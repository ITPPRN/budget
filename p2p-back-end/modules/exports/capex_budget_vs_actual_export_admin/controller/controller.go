package controller

import (
	"p2p-back-end/modules/exports/capex_budget_vs_actual_export_admin/service"
	"github.com/gofiber/fiber/v2"
)

type capexVsActualController struct {
	srv service.CapexVsActualService
}

func NewExportController(router fiber.Router, srv service.CapexVsActualService) {
	c := &capexVsActualController{srv: srv}
	router.Post("/export-capex-budget-vs-actual-admin", c.exportCapexVsActual)
}

func (c *capexVsActualController) exportCapexVsActual(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	data, filename, err := c.srv.ExportCapexVsActualExcel(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.Set("Access-Control-Expose-Headers", "Content-Disposition")

	return ctx.Status(fiber.StatusOK).Send(data)
}
