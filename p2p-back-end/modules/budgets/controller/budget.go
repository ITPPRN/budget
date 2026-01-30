package controller

import (
	"fmt"
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

	// List APIs
	router.Get("/files-budget", controller.listBudgetFiles)
	router.Get("/files-capex-budget", controller.listCapexBudgetFiles)
	router.Get("/files-capex-actual", controller.listCapexActualFiles)

	// Delete APIs
	router.Delete("/files-budget/:id", controller.deleteBudgetFile)
	router.Delete("/files-capex-budget/:id", controller.deleteCapexBudgetFile)
	router.Delete("/files-capex-actual/:id", controller.deleteCapexActualFile)

	// Rename APIs (Patch)
	router.Patch("/files-budget/:id", controller.renameBudgetFile)
	router.Patch("/files-capex-budget/:id", controller.renameCapexBudgetFile)
	router.Patch("/files-capex-actual/:id", controller.renameCapexActualFile)

	// Sync APIs (Post)
	router.Post("/files-budget/:id/sync", controller.syncBudget)
	router.Post("/files-capex-budget/:id/sync", controller.syncCapexBudget)
	router.Post("/files-capex-actual/:id/sync", controller.syncCapexActual)

	// Dashboard APIs
	router.Get("/filter-options", controller.getFilterOptions)
	router.Post("/details", controller.getBudgetDetails)
}

// ---------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------

func (c *budgetController) getFilterOptions(ctx *fiber.Ctx) error {
	options, err := c.budgetSrv.GetFilterOptions()
	if err != nil {
		fmt.Printf("[Error] getFilterOptions failed: %v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(options)
}

func (c *budgetController) getBudgetDetails(ctx *fiber.Ctx) error {
	type FilterRequest struct {
		Groups      []string `json:"groups"`
		Departments []string `json:"departments"`
		EntityGLs   []string `json:"entity_gls"`
		ConsoGLs    []string `json:"conso_gls"`
	}
	var req FilterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	details, err := c.budgetSrv.GetBudgetDetails(req.Groups, req.Departments, req.EntityGLs, req.ConsoGLs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(details)
}

func (c *budgetController) importBudget(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.budgetSrv.ImportBudget(fileHeader, "system", versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncBudget(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.SyncBudget(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *budgetController) importCapexBudget(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.budgetSrv.ImportCapexBudget(fileHeader, "system", versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncCapexBudget(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.SyncCapexBudget(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *budgetController) importCapexActual(ctx *fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.budgetSrv.ImportCapexActual(fileHeader, "system", versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncCapexActual(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.SyncCapexActual(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

// ---------------------------------------------------------------------
// Delete Handlers
// ---------------------------------------------------------------------

func (c *budgetController) deleteBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.DeleteBudgetFile(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) deleteCapexBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.DeleteCapexBudgetFile(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) deleteCapexActualFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.budgetSrv.DeleteCapexActualFile(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

// ---------------------------------------------------------------------
// Rename Handlers
// ---------------------------------------------------------------------

func (c *budgetController) renameBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	type RenameReq struct {
		NewName string `json:"new_name"`
	}
	var req RenameReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.budgetSrv.RenameBudgetFile(id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) renameCapexBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	type RenameReq struct {
		NewName string `json:"new_name"`
	}
	var req RenameReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.budgetSrv.RenameCapexBudgetFile(id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) renameCapexActualFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	type RenameReq struct {
		NewName string `json:"new_name"`
	}
	var req RenameReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.budgetSrv.RenameCapexActualFile(id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

// ---------------------------------------------------------------------
// List Handlers
// ---------------------------------------------------------------------

func (c *budgetController) listBudgetFiles(ctx *fiber.Ctx) error {
	files, err := c.budgetSrv.ListBudgetFiles()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *budgetController) listCapexBudgetFiles(ctx *fiber.Ctx) error {
	files, err := c.budgetSrv.ListCapexBudgetFiles()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *budgetController) listCapexActualFiles(ctx *fiber.Ctx) error {
	files, err := c.budgetSrv.ListCapexActualFiles()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}
