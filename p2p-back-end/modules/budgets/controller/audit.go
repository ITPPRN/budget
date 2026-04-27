package controller

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/middlewares"
)

type auditController struct {
	auditSrv models.AuditService
}

func NewAuditController(
	router fiber.Router,
	auditSrv models.AuditService,
) {
	controller := &auditController{auditSrv: auditSrv}

	router.Post("/audit/basket/add", controller.addBasket)
	router.Get("/audit/basket/list", controller.getBasket)
	router.Get("/audit/basket/in-basket-tx-ids", controller.getInBasketTxIDs)
	router.Patch("/audit/basket/:id", controller.updateBasketNote)
	router.Delete("/audit/basket/:id", controller.removeBasket)
	router.Post("/audit/approve", controller.approve)
	router.Post("/audit/report", controller.report)
	router.Post("/audit/reportable", controller.getReportableTransactions)
	router.Get("/audit/logs", controller.listLogs)
	router.Get("/audit/logs/:id/items", controller.getRejectedItems)
	router.Get("/audit/check-complete", controller.checkAuditComplete)
}

func (h *auditController) addBasket(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.UserInfo)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var itemReq []models.BasketAddItem

	if err := c.BodyParser(&itemReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format: expected an array of {transaction_id, note}"})
	}

	if len(itemReq) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no request"})
	}

	err := h.auditSrv.AddToBasket(c.UserContext(), user, itemReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save basket: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Added %d items to basket", len(itemReq)),
	})
}

func (h *auditController) updateBasketNote(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.UserInfo)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	transactionID := c.Params("id")
	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Transaction ID is required"})
	}

	var body struct {
		Note string `json:"note"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.auditSrv.UpdateBasketNote(c.UserContext(), user, transactionID, body.Note); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update note: " + err.Error()})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Note updated"})
}

func (h *auditController) getBasket(c *fiber.Ctx) error {
    user, ok := c.Locals("user").(*models.UserInfo)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }

    // เรียก Service เพื่อดึงข้อมูลตะกร้า
    items, err := h.auditSrv.GetBasketItems(c.UserContext(), user)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get basket: " + err.Error()})
    }

    // ส่งคืนข้อมูลเป็น Array ของ Object ได้เลย
    return c.JSON(items)
}

func (h *auditController) getInBasketTxIDs(c *fiber.Ctx) error {
    user, ok := c.Locals("user").(*models.UserInfo)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }
    ids, err := h.auditSrv.GetInBasketTxIDs(c.UserContext(), user)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch in-basket tx ids: " + err.Error()})
    }
    return c.JSON(ids)
}

func (h *auditController) removeBasket(c *fiber.Ctx) error {
    user, ok := c.Locals("user").(*models.UserInfo)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }

    transactionID := c.Params("id")
    
    if transactionID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Transaction ID is required"})
    }

    // โยนให้ Service จัดการลบทีละ 1 รายการ
    err := h.auditSrv.RemoveFromBasket(c.UserContext(), user, transactionID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove from basket: " + err.Error()})
    }

    return c.JSON(fiber.Map{
        "status": "success", 
        "message": "Item removed from basket",
    })
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

	// 🛠️ The Approval now processes the "Reject Basket" (rejected_item_ids)
	// and auto-completes everything else in the month.
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
	middlewares.EnforceBranchScopeFromCtx(ctx, filter)

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
	if req == nil {
		req = map[string]interface{}{}
	}
	middlewares.EnforceBranchScopeFromCtx(ctx, req)

	items, err := c.auditSrv.GetReportableTransactions(ctx.UserContext(), user, req)
	if err != nil {
		fmt.Println("GetReportableTransactions Error:", err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(items)
}

func (c *auditController) checkAuditComplete(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*models.UserInfo)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	year := ctx.Query("year")
	month := ctx.Query("month")
	if year == "" || month == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year and month are required"})
	}

	result, err := c.auditSrv.CheckAuditComplete(ctx.UserContext(), user, year, month)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(result)
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
