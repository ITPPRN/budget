package controller

import (
	"p2p-back-end/modules/exports/capex_department_status_export_admin/service"
	"p2p-back-end/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

type capexDeptStatusController struct {
	srv service.CapexDeptStatusService
}

func NewExportController(router fiber.Router, srv service.CapexDeptStatusService) {
	c := &capexDeptStatusController{srv: srv}
	router.Post("/export-capex-department-status-admin", c.exportCapexDeptStatus)
}

func (c *capexDeptStatusController) exportCapexDeptStatus(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req == nil {
		req = map[string]interface{}{}
	}
	middlewares.EnforceBranchScopeFromCtx(ctx, req)

	data, filename, err := c.srv.ExportCapexDeptStatusExcel(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	ctx.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.Set("Access-Control-Expose-Headers", "Content-Disposition")

	return ctx.Status(fiber.StatusOK).Send(data)
}
