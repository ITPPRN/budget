package controller

import (
	"fmt"
	"p2p-back-end/modules/entities/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type budgetController struct {
	plSrv     models.PLBudgetService
	capexSrv  models.CapexService
	actualSrv models.ActualService
	masterSrv models.MasterDataService
	dashSrv   models.DashboardService
}

func NewBudgetController(
	router fiber.Router,
	plSrv models.PLBudgetService,
	capexSrv models.CapexService,
	actualSrv models.ActualService,
	masterSrv models.MasterDataService,
	dashSrv models.DashboardService,
) {
	controller := &budgetController{plSrv: plSrv, capexSrv: capexSrv, actualSrv: actualSrv, masterSrv: masterSrv, dashSrv: dashSrv}
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

	// Clear APIs (Post)
	router.Post("/clear-budget", controller.clearBudget)
	router.Post("/clear-capex-budget", controller.clearCapexBudget)
	router.Post("/clear-capex-actual", controller.clearCapexActual)

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
	router.Get("/actual-years", controller.getActualYears)            // New: Distinct Years
	router.Get("/debug-date", controller.getDebugDate)                // Debug

	// GL Mapping APIs
	router.Get("/gl-mappings", controller.listGLMappings)
	router.Get("/gl-mappings/:id", controller.getGLMapping)
	router.Post("/gl-mappings", controller.createGLMapping)
	router.Put("/gl-mappings/:id", controller.updateGLMapping)
	router.Delete("/gl-mappings/:id", controller.deleteGLMapping)
	router.Post("/gl-mappings/import", controller.importGLMapping)

	// Budget Structure API
	router.Get("/budget-structure", controller.getBudgetStructureTree)
	router.Get("/budget-structure-list", controller.listBudgetStructure)
	router.Get("/budget-structure-list/:id", controller.getBudgetStructure)
	router.Post("/budget-structure", controller.createBudgetStructure)
	router.Put("/budget-structure/:id", controller.updateBudgetStructure)
	router.Delete("/budget-structure/:id", controller.deleteBudgetStructure)

	// User Config APIs
	router.Get("/configs", controller.getUserConfigs)
	router.Post("/configs/:key", controller.setUserConfig)
}

func (c *budgetController) getDebugDate(ctx *fiber.Ctx) error {
	date, err := c.actualSrv.GetRawDate(ctx.UserContext())
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

	details, err := c.dashSrv.GetActualTransactions(ctx.UserContext(), req)
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

	summary, err := c.dashSrv.GetDashboardSummary(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(summary)
}

func (c *budgetController) getActualYears(ctx *fiber.Ctx) error {
	years, err := c.dashSrv.GetActualYears(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(years)
}

func (c *budgetController) getFilterOptions(ctx *fiber.Ctx) error {
	options, err := c.dashSrv.GetFilterOptions(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(options)
}

func (c *budgetController) getOrganizationStructure(ctx *fiber.Ctx) error {
	structure, err := c.dashSrv.GetOrganizationStructure(ctx.UserContext())
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

	details, err := c.dashSrv.GetBudgetDetails(ctx.UserContext(), req)
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

	details, err := c.dashSrv.GetActualDetails(ctx.UserContext(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(details)
}

func (c *budgetController) importBudget(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID not found in context"})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.plSrv.ImportBudget(ctx.UserContext(), fileHeader, userID, versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncBudget(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.plSrv.SyncBudget(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *budgetController) syncActuals(ctx *fiber.Ctx) error {
	type SyncReq struct {
		Year   string   `json:"year"`
		Months []string `json:"months"` // New: Optional month list
	}
	var req SyncReq
	if err := ctx.BodyParser(&req); err != nil {
		// Allow empty body -> Default to current year
	}
	if req.Year == "" {
		req.Year = fmt.Sprintf("%d", time.Now().Year())
	}

	if err := c.actualSrv.SyncActuals(ctx.UserContext(), req.Year, req.Months); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": fmt.Sprintf("Actuals for %s synced successfully from Database", req.Year)})
}

func (c *budgetController) importCapexBudget(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID not found in context"})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.capexSrv.ImportCapexBudget(ctx.UserContext(), fileHeader, userID, versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncCapexBudget(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.SyncCapexBudget(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

func (c *budgetController) importCapexActual(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID not found in context"})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	versionName := ctx.FormValue("version_name")

	if err := c.capexSrv.ImportCapexActual(ctx.UserContext(), fileHeader, userID, versionName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "uploaded", "message": "File uploaded successfully. Please Sync to apply data."})
}

func (c *budgetController) syncCapexActual(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.SyncCapexActual(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "synced", "message": "Data synced successfully"})
}

// ---------------------------------------------------------------------
// Clear Handlers
// ---------------------------------------------------------------------

func (c *budgetController) clearBudget(ctx *fiber.Ctx) error {
	if err := c.plSrv.ClearBudget(ctx.UserContext()); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "cleared", "message": "Budget data cleared successfully"})
}

func (c *budgetController) clearCapexBudget(ctx *fiber.Ctx) error {
	if err := c.capexSrv.ClearCapexBudget(ctx.UserContext()); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "cleared", "message": "Capex Budget data cleared successfully"})
}

func (c *budgetController) clearCapexActual(ctx *fiber.Ctx) error {
	if err := c.capexSrv.ClearCapexActual(ctx.UserContext()); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "cleared", "message": "Capex Actual data cleared successfully"})
}

// ---------------------------------------------------------------------
// Delete Handlers
// ---------------------------------------------------------------------

func (c *budgetController) deleteBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.plSrv.DeleteBudgetFile(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) deleteCapexBudgetFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.DeleteCapexBudgetFile(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}
func (c *budgetController) deleteCapexActualFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.capexSrv.DeleteCapexActualFile(ctx.UserContext(), id); err != nil {
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
	if err := c.plSrv.RenameBudgetFile(ctx.UserContext(), id, req.NewName); err != nil {
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
	if err := c.capexSrv.RenameCapexBudgetFile(ctx.UserContext(), id, req.NewName); err != nil {
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
	if err := c.capexSrv.RenameCapexActualFile(ctx.UserContext(), id, req.NewName); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

// ---------------------------------------------------------------------
// List Handlers
// ---------------------------------------------------------------------

func (c *budgetController) listBudgetFiles(ctx *fiber.Ctx) error {
	files, err := c.plSrv.ListBudgetFiles(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *budgetController) listCapexBudgetFiles(ctx *fiber.Ctx) error {
	files, err := c.capexSrv.ListCapexBudgetFiles(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(files)
}

func (c *budgetController) listCapexActualFiles(ctx *fiber.Ctx) error {
	files, err := c.capexSrv.ListCapexActualFiles(ctx.UserContext())
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

	if err := c.actualSrv.DeleteActualFacts(ctx.UserContext(), year); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "Actuals deleted successfully", "year": year})
}

// ---------------------------------------------------------------------
// GL Mapping Handlers
// ---------------------------------------------------------------------

func (c *budgetController) listGLMappings(ctx *fiber.Ctx) error {
	mappings, err := c.masterSrv.ListGLMappings(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(mappings)
}

func (c *budgetController) getGLMapping(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	mapping, err := c.masterSrv.GetGLMappingByID(ctx.UserContext(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mapping not found"})
	}
	return ctx.JSON(mapping)
}

func (c *budgetController) createGLMapping(ctx *fiber.Ctx) error {
	var body models.GlMappingEntity
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.masterSrv.CreateGLMapping(ctx.UserContext(), &body); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(body)
}

func (c *budgetController) updateGLMapping(ctx *fiber.Ctx) error {
	var body models.GlMappingEntity
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	// ensure ID is correct from param if needed, but assuming body has it.
	if err := c.masterSrv.UpdateGLMapping(ctx.UserContext(), &body); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(body)
}

func (c *budgetController) deleteGLMapping(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.masterSrv.DeleteGLMapping(ctx.UserContext(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

func (c *budgetController) importGLMapping(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to get file"})
	}

	err = c.masterSrv.ImportGLMapping(ctx.UserContext(), file)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "GL Mapping imported successfully"})
}

func (c *budgetController) getBudgetStructureTree(ctx *fiber.Ctx) error {
	tree, err := c.masterSrv.GetBudgetStructureTree(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(tree)
}

func (c *budgetController) listBudgetStructure(ctx *fiber.Ctx) error {
	list, err := c.masterSrv.ListBudgetStructure(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(list)
}

func (c *budgetController) getBudgetStructure(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}
	entity, err := c.masterSrv.GetBudgetStructureByID(ctx.UserContext(), uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Structure node not found"})
	}
	return ctx.JSON(entity)
}

func (c *budgetController) createBudgetStructure(ctx *fiber.Ctx) error {
	var body models.BudgetStructureEntity
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := c.masterSrv.CreateBudgetStructure(ctx.UserContext(), &body); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(body)
}

func (c *budgetController) updateBudgetStructure(ctx *fiber.Ctx) error {
	var body models.BudgetStructureEntity
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}
	body.ID = uint(id)

	if err := c.masterSrv.UpdateBudgetStructure(ctx.UserContext(), &body); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(body)
}

func (c *budgetController) deleteBudgetStructure(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}
	if err := c.masterSrv.DeleteBudgetStructure(ctx.UserContext(), uint(id)); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete: " + err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}

// User Config Handlers

func (c *budgetController) getUserConfigs(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID not found in context"})
	}

	configs, err := c.masterSrv.GetUserConfigs(ctx.UserContext(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(configs)
}

func (c *budgetController) setUserConfig(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID not found in context"})
	}

	key := ctx.Params("key")
	if key == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Key is required"})
	}

	type ConfigReq struct {
		Value string `json:"value"`
	}
	var req ConfigReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	if err := c.masterSrv.SetUserConfig(ctx.UserContext(), userID, key, req.Value); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(fiber.Map{"status": "success"})
}


