package controller

import (
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

type capexController struct {
	capexSrv models.CapexService
}

func NewCapexController(router fiber.Router, capexSrv models.CapexService) {
	controller := &capexController{capexSrv: capexSrv}

	// Import
	router.Post("/import-budget", controller.importCapexBudget)
	router.Post("/import-actual", controller.importCapexActual)

	// List
	router.Get("/files-budget", controller.listCapexBudgetFiles)
	router.Get("/files-actual", controller.listCapexActualFiles)

	// Sync
	router.Post("/files-budget/:id/sync", controller.syncCapexBudget)
	router.Post("/files-actual/:id/sync", controller.syncCapexActual)

	// Delete
	router.Delete("/files-budget/:id", controller.deleteCapexBudgetFile)
	router.Delete("/files-actual/:id", controller.deleteCapexActualFile)

	// Rename
	router.Patch("/files-budget/:id", controller.renameCapexBudgetFile)
	router.Patch("/files-actual/:id", controller.renameCapexActualFile)

	// Dashboard
	router.Post("/dashboard-summary", controller.getCapexDashboardSummary)
}

// ---------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------

func (c *capexController) importCapexBudget(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	// Get UserID from Context (Assuming Auth Middleware sets it)
	// For now, hardcode "system" or extract from token if middleware exists
	userID := "system"

	if err := c.capexSrv.ImportCapexBudget(ctx.UserContext(), fileHeader, userID, versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *capexController) importCapexActual(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")
	userID := "system"

	if err := c.capexSrv.ImportCapexActual(ctx.UserContext(), fileHeader, userID, versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *capexController) listCapexBudgetFiles(ctx *fiber.Ctx) error {
	files, err := c.capexSrv.ListCapexBudgetFiles(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *capexController) listCapexActualFiles(ctx *fiber.Ctx) error {
	files, err := c.capexSrv.ListCapexActualFiles(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *capexController) syncCapexBudget(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.SyncCapexBudget(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *capexController) syncCapexActual(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.SyncCapexActual(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *capexController) deleteCapexBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.DeleteCapexBudgetFile(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

func (c *capexController) deleteCapexActualFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.DeleteCapexActualFile(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

func (c *capexController) renameCapexBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	type RenameReq struct {
		NewName string `json:"new_name"`
	}
	var req RenameReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.capexSrv.RenameCapexBudgetFile(ctx.UserContext(), id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

func (c *capexController) renameCapexActualFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	type RenameReq struct {
		NewName string `json:"new_name"`
	}
	var req RenameReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.capexSrv.RenameCapexActualFile(ctx.UserContext(), id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

func (c *capexController) getCapexDashboardSummary(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req == nil {
		req = map[string]interface{}{}
	}
	middlewares.EnforceBranchScopeFromCtx(ctx, req)

	summary, err := c.capexSrv.GetCapexDashboardSummary(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(summary)
}
