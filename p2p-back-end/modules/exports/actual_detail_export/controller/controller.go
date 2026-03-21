package controller

import (
	"p2p-back-end/modules/exports/actual_detail_export/service"
	"github.com/gofiber/fiber/v2"
)

type actualExportController struct {
	srv service.ActualExportService
}

func NewExportController(router fiber.Router, srv service.ActualExportService) {
	c := &actualExportController{srv: srv}
	router.Post("/export-actual-detail", c.exportActualDetail)
}

func (c *actualExportController) exportActualDetail(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	data, filename, err := c.srv.ExportActualDetailExcel(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.Set("Access-Control-Expose-Headers", "Content-Disposition")

	return ctx.Status(fiber.StatusOK).Send(data)
}
