package controller

import (
	"p2p-back-end/modules/entities/models"

	"github.com/gofiber/fiber/v2"
)

type ownerController struct {
	ownerService models.OwnerService
}

func NewOwnerController(r fiber.Router, service models.OwnerService) {
	controller := &ownerController{ownerService: service}

	r.Post("/dashboard-summary", controller.GetDashboardSummary)
	r.Get("/filter-options", controller.GetFilterOptions)
	r.Get("/organization-structure", controller.GetOrganizationStructure)
	r.Get("/filter-lists", controller.GetOwnerFilterLists)
	r.Post("/actual-transactions", controller.GetActualTransactions)
	r.Post("/sync-actuals", controller.AutoSyncOwnerActuals)
}

// GetDashboardSummary
// POST /api/v1/owner/dashboard-summary
func (c *ownerController) GetDashboardSummary(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.UserInfo)
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	summary, err := c.ownerService.GetDashboardSummary(user, req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(summary)
}

// GetFilterOptions (Tree)
// GET /api/v1/owner/filter-options
func (c *ownerController) GetFilterOptions(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.UserInfo)
	opts, err := c.ownerService.GetFilterOptions(user)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(opts)
}

func (c *ownerController) GetOrganizationStructure(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.UserInfo)
	structure, err := c.ownerService.GetOrganizationStructure(user)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(structure)
}

// GetOwnerFilterLists (Dropdowns)
// GET /api/v1/owner/filter-lists
func (c *ownerController) GetOwnerFilterLists(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.UserInfo)
	lists, err := c.ownerService.GetOwnerFilterLists(user)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(lists)
}

// GetActualTransactions
// POST /api/v1/owner/actual-transactions
func (c *ownerController) GetActualTransactions(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.UserInfo)
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	txs, err := c.ownerService.GetActualTransactions(user, req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(txs)
}

// AutoSyncOwnerActuals
// POST /api/v1/owner/sync-actuals
func (c *ownerController) AutoSyncOwnerActuals(ctx *fiber.Ctx) error {
	if err := c.ownerService.AutoSyncOwnerActuals(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "Owner Actuals Synced Successfully"})
}
