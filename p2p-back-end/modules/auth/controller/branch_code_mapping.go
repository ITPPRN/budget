package authcontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"p2p-back-end/modules/entities/models"
)

type upsertBranchCodeMappingReq struct {
	CompanyID  string `json:"company_id"`
	BranchCode string `json:"branch_code"`
}

func (h *authController) listBranchCodeMappings(c *fiber.Ctx, _ *models.UserInfo) error {
	rows, err := h.branchCodeMapSrv.List(c.UserContext())
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, rows)
}

func (h *authController) upsertBranchCodeMapping(c *fiber.Ctx, _ *models.UserInfo) error {
	var req upsertBranchCodeMappingReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid company_id"})
	}

	row, err := h.branchCodeMapSrv.Upsert(c.UserContext(), companyID, req.BranchCode)
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, row)
}

func (h *authController) deleteBranchCodeMapping(c *fiber.Ctx, _ *models.UserInfo) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}
	if err := h.branchCodeMapSrv.Delete(c.UserContext(), id); err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, fiber.Map{"deleted": true})
}

func (h *authController) listCompaniesForMapping(c *fiber.Ctx, _ *models.UserInfo) error {
	rows, err := h.branchCodeMapSrv.ListCompanies(c.UserContext())
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, rows)
}

func (h *authController) listAvailableBranchCodes(c *fiber.Ctx, _ *models.UserInfo) error {
	codes, err := h.branchCodeMapSrv.ListAvailableBranchCodes(c.UserContext())
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, codes)
}

func (h *authController) importBranchCodeMappings(c *fiber.Ctx, _ *models.UserInfo) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing file"})
	}
	res, err := h.branchCodeMapSrv.ImportFromExcel(c.UserContext(), fileHeader)
	if err != nil {
		return responseWithError(c, err)
	}
	return responseSuccess(c, res)
}

func (h *authController) downloadBranchCodeMappingTemplate(c *fiber.Ctx, _ *models.UserInfo) error {
	data, err := h.branchCodeMapSrv.GenerateImportTemplate(c.UserContext())
	if err != nil {
		return responseWithError(c, err)
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="branch_code_mapping_template.xlsx"`)
	c.Set("Access-Control-Expose-Headers", "Content-Disposition")
	return c.Status(fiber.StatusOK).Send(data)
}
