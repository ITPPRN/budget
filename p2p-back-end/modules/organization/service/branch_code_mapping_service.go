package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

type companyBranchCodeMappingService struct {
	repo models.CompanyBranchCodeMappingRepository
}

func NewCompanyBranchCodeMappingService(repo models.CompanyBranchCodeMappingRepository) models.CompanyBranchCodeMappingService {
	return &companyBranchCodeMappingService{repo: repo}
}

func (s *companyBranchCodeMappingService) List(ctx context.Context) ([]models.CompanyBranchCodeMappingEntity, error) {
	return s.repo.List(ctx)
}

// ResolveBranchCodes returns ALL branch codes mapped to the given company.
// Returns an empty slice (not nil) when no mapping exists.
func (s *companyBranchCodeMappingService) ResolveBranchCodes(ctx context.Context, companyID uuid.UUID) ([]string, error) {
	if companyID == uuid.Nil {
		return []string{}, nil
	}
	rows, err := s.repo.ListByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, r := range rows {
		codes = append(codes, r.BranchCode)
	}
	return codes, nil
}

// Upsert is additive: it inserts (company_id, branch_code) if not present,
// and is a no-op if the exact pair already exists. It does NOT replace
// other codes for the same company.
func (s *companyBranchCodeMappingService) Upsert(ctx context.Context, companyID uuid.UUID, branchCode string) (*models.CompanyBranchCodeMappingEntity, error) {
	if companyID == uuid.Nil {
		return nil, errors.New("company_id is required")
	}
	code := strings.TrimSpace(branchCode)
	if code == "" {
		return nil, errors.New("branch_code is required")
	}

	m := &models.CompanyBranchCodeMappingEntity{
		CompanyID:  companyID,
		BranchCode: code,
	}
	if err := s.repo.Upsert(ctx, m); err != nil {
		return nil, fmt.Errorf("branchCodeMappingService.Upsert: %w", err)
	}
	// Return all current rows for this company so the caller has full state
	rows, err := s.repo.ListByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if r.BranchCode == code {
			return &r, nil
		}
	}
	return nil, nil
}

func (s *companyBranchCodeMappingService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("id is required")
	}
	return s.repo.Delete(ctx, id)
}

func (s *companyBranchCodeMappingService) ListCompanies(ctx context.Context) ([]models.Companies, error) {
	return s.repo.ListCompanies(ctx)
}

func (s *companyBranchCodeMappingService) ListAvailableBranchCodes(ctx context.Context) ([]string, error) {
	return s.repo.ListAvailableBranchCodes(ctx)
}

// GenerateImportTemplate produces an .xlsx pre-filled with every Company's
// (Name, Branch No) pair. The Branch Code column is filled with any existing
// mapping so admins can review/edit in place. Sorted by name then branch_no
// for predictable diffs.
func (s *companyBranchCodeMappingService) GenerateImportTemplate(ctx context.Context) ([]byte, error) {
	companies, err := s.repo.ListCompanies(ctx)
	if err != nil {
		return nil, fmt.Errorf("template: list companies: %w", err)
	}
	mappings, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("template: list mappings: %w", err)
	}
	// Group existing codes by company so we can emit one row per code.
	codesByCompany := map[uuid.UUID][]string{}
	for _, m := range mappings {
		codesByCompany[m.CompanyID] = append(codesByCompany[m.CompanyID], m.BranchCode)
	}

	h := utils.NewExcelHelper("BranchCodeMapping")
	if err := h.SetHeaders([]string{"Company Name", "Branch Name", "Branch No", "Branch Code"}); err != nil {
		return nil, fmt.Errorf("template: set headers: %w", err)
	}
	_ = h.File.SetColWidth(h.Sheet, "A", "A", 45)
	_ = h.File.SetColWidth(h.Sheet, "B", "B", 30)
	_ = h.File.SetColWidth(h.Sheet, "C", "C", 12)
	_ = h.File.SetColWidth(h.Sheet, "D", "D", 18)

	row := 2 // header is row 1
	for _, c := range companies {
		branchName := c.BranchName
		if branchName == "" {
			branchName = c.BranchNameEn
		}
		codes := codesByCompany[c.ID]
		if len(codes) == 0 {
			// Unmapped company: one row with empty Branch Code for admin to fill
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("A%d", row), c.Name)
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("B%d", row), branchName)
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("C%d", row), c.BranchNo)
			row++
			continue
		}
		// One row per existing code (preserves multi-code mappings on round-trip)
		for _, code := range codes {
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("A%d", row), c.Name)
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("B%d", row), branchName)
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("C%d", row), c.BranchNo)
			_ = h.File.SetCellValue(h.Sheet, fmt.Sprintf("D%d", row), code)
			row++
		}
	}

	return h.WriteToBuffer()
}

// ImportFromExcel parses an .xlsx with header row + 4 columns:
//   A: Company Name        (must match companies.name exactly after trim)
//   B: Branch Name         (informational only — not used for matching)
//   C: Branch No           (must match companies.branch_no exactly after trim)
//   D: Branch Code         (the value to upsert)
//
// Each row resolves to a Company by (name, branch_no) pair, then upserts the
// mapping. Rows with empty fields are skipped; rows whose company isn't found
// are reported in Errors but do not abort the import.
//
// Backward-compatible: if the file has only 3 columns (legacy format with no
// Branch Name), columns shift left — A=Company, B=Branch No, C=Branch Code.
func (s *companyBranchCodeMappingService) ImportFromExcel(ctx context.Context, fileHeader *multipart.FileHeader) (*models.ImportBranchCodeMappingResult, error) {
	if fileHeader == nil {
		return nil, errors.New("no file uploaded")
	}
	ext := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, ".")+1:])
	if ext != "xlsx" {
		return nil, errors.New("only .xlsx files are supported")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("read excel: %w", err)
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("excel file has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	if len(rows) < 2 {
		return nil, errors.New("excel file is too short (need header + at least 1 data row)")
	}

	result := &models.ImportBranchCodeMappingResult{}

	// Track existing (company_id, branch_code) combos for insert-vs-update reporting.
	// One company can have many codes — we report per-pair, not per-company.
	type pairKey struct {
		CompanyID uuid.UUID
		Code      string
	}
	existingPair := map[pairKey]bool{}
	if existing, err := s.repo.List(ctx); err == nil {
		for _, m := range existing {
			existingPair[pairKey{m.CompanyID, m.BranchCode}] = true
		}
	}

	// Detect schema by header width: 4 columns (new) or 3 columns (legacy)
	headerWidth := len(rows[0])
	hasBranchName := headerWidth >= 4

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		var companyName, branchNo, branchCode string
		if hasBranchName {
			padded := make([]string, 4)
			copy(padded, row)
			companyName = strings.TrimSpace(padded[0])
			// padded[1] = Branch Name — informational, not used for matching
			branchNo = strings.TrimSpace(padded[2])
			branchCode = strings.TrimSpace(padded[3])
		} else {
			padded := make([]string, 3)
			copy(padded, row)
			companyName = strings.TrimSpace(padded[0])
			branchNo = strings.TrimSpace(padded[1])
			branchCode = strings.TrimSpace(padded[2])
		}

		if companyName == "" && branchNo == "" && branchCode == "" {
			continue
		}
		if companyName == "" || branchCode == "" {
			result.Skipped++
			result.Errors = append(result.Errors,
				fmt.Sprintf("แถว %d: company_name หรือ branch_code ว่าง", i+1))
			continue
		}

		company, err := s.repo.FindCompanyByNameAndBranchNo(ctx, companyName, branchNo)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors,
				fmt.Sprintf("แถว %d: ค้นหา company ล้มเหลว: %v", i+1, err))
			continue
		}
		if company == nil {
			result.Skipped++
			result.Errors = append(result.Errors,
				fmt.Sprintf("แถว %d: ไม่พบ company '%s' / branch_no '%s'", i+1, companyName, branchNo))
			continue
		}

		m := &models.CompanyBranchCodeMappingEntity{
			CompanyID:  company.ID,
			BranchCode: branchCode,
		}
		if err := s.repo.Upsert(ctx, m); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors,
				fmt.Sprintf("แถว %d: upsert ล้มเหลว: %v", i+1, err))
			continue
		}

		key := pairKey{company.ID, branchCode}
		if existingPair[key] {
			result.Updated++
		} else {
			result.Imported++
			existingPair[key] = true
		}
	}

	return result, nil
}
