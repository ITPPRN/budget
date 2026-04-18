package controller

import (
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/exports/actual_detail_export_owner/service"
	"p2p-back-end/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

type ownerActualExportController struct {
	srv service.OwnerActualExportService
}

func NewExportController(router fiber.Router, srv service.OwnerActualExportService) {
	c := &ownerActualExportController{srv: srv}
	router.Post("/export-actual-detail", c.exportOwnerActualDetail)
}

func (c *ownerActualExportController) exportOwnerActualDetail(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req == nil {
		req = map[string]interface{}{}
	}
	middlewares.EnforceBranchScopeFromCtx(ctx, req)

	user := ctx.Locals("user").(*models.UserInfo)
	data, filename, err := c.srv.ExportOwnerActualDetailExcel(ctx.UserContext(), user, req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.Set("Access-Control-Expose-Headers", "Content-Disposition")

	return ctx.Status(fiber.StatusOK).Send(data)
}
