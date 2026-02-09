package controller

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"time"

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

	// Actuals APIs
	router.Post("/sync-actuals", controller.syncActuals)                   // No file ID needed
	router.Delete("/actuals-facts/:year", controller.deleteActuals)        // New: Delete by Year
	router.Post("/actuals-details", controller.getActualDetails)           // Aggregated View
	router.Post("/actuals-transactions", controller.getActualTransactions) // Detail View

	// Dashboard APIs
	router.Get("/filter-options", controller.getFilterOptions)
	router.Get("/organization-structure", controller.getOrganizationStructure)
	router.Post("/details", controller.getBudgetDetails)
	router.Post("/dashboard-summary", controller.getDashboardSummary) // New
	router.Get("/debug-date", controller.getDebugDate)                // Debug
}

func (c *budgetController) getDebugDate(ctx *fiber.Ctx) error {
	date, err := c.budgetSrv.GetRawDate()
	if err != nil {
		fmt.Println("DEBUG DATE ERROR:", err.Error())
		return ctx.Status(500).SendString(err.Error())
	}
	fmt.Printf("\n[DEBUG] RAW HMW DATE FORMAT: '%s'\n\n", date)
	return ctx.SendString(date)
}

// ---------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------

func (c *budgetController) getActualTransactions(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	details, err := c.budgetSrv.GetActualTransactions(req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(details)
}

func (c *budgetController) getDashboardSummary(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	summary, err := c.budgetSrv.GetDashboardSummary(req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(summary)
}

func (c *budgetController) getFilterOptions(ctx *fiber.Ctx) error {
	options, err := c.budgetSrv.GetFilterOptions()
	if err != nil {
		fmt.Printf("[Error] getFilterOptions failed: %v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(options)
}

func (c *budgetController) getOrganizationStructure(ctx *fiber.Ctx) error {
	structure, err := c.budgetSrv.GetOrganizationStructure()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(structure)
}

func (c *budgetController) getBudgetDetails(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	details, err := c.budgetSrv.GetBudgetDetails(req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(details)
}

func (c *budgetController) getActualDetails(ctx *fiber.Ctx) error {
	var req map[string]interface{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	details, err := c.budgetSrv.GetActualDetails(req)
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

func (c *budgetController) syncActuals(ctx *fiber.Ctx) error {
	type SyncReq struct {
		Year string `json:"year"`
	}
	var req SyncReq
	if err := ctx.BodyParser(&req); err != nil {
		// Allow empty body -> Default to current year
	}
	if req.Year == "" {
		req.Year = fmt.Sprintf("%d", time.Now().Year())
	}

	if err := c.budgetSrv.SyncActuals(req.Year); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": fmt.Sprintf("Actuals for %s synced successfully from Database", req.Year)})
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

func (c *budgetController) deleteActuals(ctx *fiber.Ctx) error {
	year := ctx.Params("year")
	if year == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Year is required"})
	}

	if err := c.budgetSrv.DeleteActualFacts(year); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "Actuals deleted successfully", "year": year})
}
