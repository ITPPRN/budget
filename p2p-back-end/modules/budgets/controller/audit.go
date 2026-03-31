package controller

import (
	"fmt"
	"p2p-back-end/modules/entities/models"

	"github.com/gofiber/fiber/v2"
)

type auditController struct {
	auditSrv models.AuditService
}

func NewAuditController(
	router fiber.Router,
	auditSrv models.AuditService,
) {
	controller := &auditController{auditSrv: auditSrv}
	
	router.Post("/audit/approve", controller.approve)
	router.Post("/audit/report", controller.report)
	router.Post("/audit/reportable", controller.getReportableTransactions)
	router.Get("/audit/logs", controller.listLogs)
	router.Get("/audit/logs/:id/items", controller.getRejectedItems)
}

func (c *auditController) approve(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*models.UserInfo)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	// 🛡️ Prevent empty selection approval (especially for Owners)
	ids, _ := req["selected_item_ids"].([]interface{})
	if len(ids) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Please select at least one item to approve/report"})
	}

	if err := c.auditSrv.Approve(ctx.UserContext(), user, req); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"status": "success", "message": "Selection approved and processed"})
}

func (c *auditController) report(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*models.UserInfo)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	if err := c.auditSrv.Report(ctx.UserContext(), user, req); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"status": "success", "message": "Audit log reported"})
}

func (c *auditController) listLogs(ctx *fiber.Ctx) error {
	filter := make(map[string]interface{})
	if dept := ctx.Query("department"); dept != "" {
		filter["department"] = dept
	}
	if year := ctx.Query("year"); year != "" {
		filter["year"] = year
	}
	if month := ctx.Query("month"); month != "" {
		filter["month"] = month
	}
	if entity := ctx.Query("entity"); entity != "" {
		filter["entity"] = entity
	}

	logs, err := c.auditSrv.ListLogs(ctx.UserContext(), filter)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(logs)
}

func (c *auditController) getReportableTransactions(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*models.UserInfo)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	items, err := c.auditSrv.GetReportableTransactions(ctx.UserContext(), user, req)
	if err != nil {
		fmt.Println("GetReportableTransactions Error:", err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(items)
}

func (c *auditController) getRejectedItems(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Log ID is required"})
	}

	items, err := c.auditSrv.GetRejectedItemDetails(ctx.UserContext(), id)
	if err != nil {
		fmt.Println("GetRejectedItems Error:", err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(items)
}
