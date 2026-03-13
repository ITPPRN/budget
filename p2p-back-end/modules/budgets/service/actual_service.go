package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"p2p-back-end/modules/entities/models"
)

type actualService struct {
	repo         models.ActualRepository
	masterSrv    models.MasterDataService
	dashboardSrv models.DashboardService
	depSrv       models.DepartmentService
}

func NewActualService(repo models.ActualRepository, masterSrv models.MasterDataService, dashboardSrv models.DashboardService, depSrv models.DepartmentService) models.ActualService {
	return &actualService{repo: repo, masterSrv: masterSrv, dashboardSrv: dashboardSrv, depSrv: depSrv}
}

type AggKey struct {
	Entity, Branch, Dept, NavCode, EntityGL, ConsoGL, GLName, VendorName, Month string
}

type HeaderKey struct {
	Entity, Branch, Dept, NavCode, EntityGL, ConsoGL, GLName, VendorName string
}

func (s *actualService) SyncActuals(year string, months []string) error {
	fmt.Printf("[DEBUG] SyncActuals: Start DB Sync (Optimized Batch) for Year %s, Months %v...\n", year, months)

	// 1. Fetch Mapping Metadata (GL Whitening)
	mappings, err := s.masterSrv.ListGLMappings()
	if err != nil {
		return err
	}
	mappingMap := make(map[string]models.GlMappingEntity)
	for _, m := range mappings {
		if m.IsActive {
			mappingMap[m.EntityGL] = m
		}
	}

	// 2. Define Mapping Configs (Move maps outside loops)
	entityNameMap := map[string]string{
		"HONDA MALIWAN":    "HMW",
		"AUTOCORP HOLDING": "ACG",
		"CLIK":             "CLIK",
	}
	branchNameMap := map[string]string{
		"BURIRUM": "BUR", "HEAD OFFICE": "HOF", "KRABI": "KBI",
		"MINI_SURIN": "MSR", "MUEANG KRABI": "MKB", "NAKA": "NAK",
		"NANGRONG": "AVN", "PHACHA": "PHC", "PHUKET": "PRA",
		"SURIN": "SUR", "VEERAWAT": "VEE",
		"AUTOCORP HEAD OFFICE": "HQ",
		"":                     "Branch00",
		"BRANCH01": "Branch01", "BRANCH02": "Branch02", "BRANCH03": "Branch03",
		"BRANCH04": "Branch04", "BRANCH05": "Branch05", "BRANCH06": "Branch06",
		"BRANCH07": "Branch07", "BRANCH08": "Branch08", "BRANCH09": "Branch09",
		"BRANCH10": "Branch10", "BRANCH11": "Branch11", "BRANCH12": "Branch12",
		"BRANCH13": "Branch13", "BRANCH14": "Branch14", "BRANCH15": "Branch15",
		"HEADOFFICE": "HOF",
	}
	monthToCodeMap := map[string]string{
		"01": "JAN", "02": "FEB", "03": "MAR", "04": "APR", "05": "MAY", "06": "JUN",
		"07": "JUL", "08": "AUG", "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
	}

	normalize := func(s string) string {
		return strings.TrimSpace(strings.ToUpper(s))
	}

	mapToCode := func(rawVal string, m map[string]string) string {
		norm := normalize(rawVal)
		if code, ok := m[norm]; ok {
			return code
		}
		if rawVal == "" {
			if v, ok := m[""]; ok {
				return v
			}
		}
		return strings.TrimSpace(rawVal)
	}

	// Determine months to process
	targetMonths := months
	if len(targetMonths) == 0 {
		targetMonths = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	}
	fullYearSync := len(months) == 0

	// 3. GL Profile Map for Fact Building (Safe memory size)
	glProfileMap := make(map[string]string)
	structures, _ := s.masterSrv.ListBudgetStructure()
	for _, st := range structures {
		if st.ConsoGL != "" {
			glProfileMap[st.ConsoGL] = st.Group1
		}
	}

	// 4. Persistence with Transaction & Batching
	return s.repo.WithTrx(func(trxRepo models.ActualRepository) error {
		if fullYearSync {
			// Wipe whole year once
			if err := trxRepo.DeleteActualFactsByYear(year); err != nil {
				return err
			}
			if err := trxRepo.DeleteActualTransactionsByYear(year); err != nil {
				return err
			}
		} else {
			// Wipe target months only
			for _, m := range targetMonths {
				if err := trxRepo.DeleteActualFactsByMonth(year, m); err != nil {
					return err
				}
				if err := trxRepo.DeleteActualTransactionsByMonth(year, m); err != nil {
					return err
				}
			}
		}

		// Keep headers aggregation in memory (Aggregated data is safe)
		mergedFactMap := make(map[AggKey]decimal.Decimal)

		for _, mName := range targetMonths {
			fmt.Printf("[Sync] Processing Month: %s\n", mName)

			// Fetch one month
			hmwRows, err := trxRepo.GetRawTransactionsHMW(year, []string{mName})
			if err != nil {
				return err
			}
			clikRows, err := trxRepo.GetRawTransactionsCLIK(year, []string{mName})
			if err != nil {
				return err
			}

			var transactions []models.ActualTransactionEntity

			processRowsBatch := func(rows []models.ActualTransactionDTO) {
				for _, row := range rows {
					mapping, ok := mappingMap[row.EntityGL]
					if !ok {
						continue
					}

					company := mapToCode(row.Company, entityNameMap)
					branch := mapToCode(row.Branch, branchNameMap)
					deptCode := row.Department
					lookupDept := normalize(row.Department)
					if masterDept, err := s.depSrv.GetMasterDepartment(lookupDept, company); err == nil && masterDept != nil {
						deptCode = masterDept.Code
					}

					// 1. Transaction Table (Centralized Detail)
					transactions = append(transactions, models.ActualTransactionEntity{
						ID:          uuid.New(),
						Source:      row.Source,
						PostingDate: row.PostingDate,
						DocNo:       row.DocNo,
						Description: row.Description,
						Amount:      row.Amount,
						VendorName:  row.Vendor,
						Entity:      company,
						Branch:      branch,
						Department:  deptCode,
						EntityGL:    row.EntityGL,
						ConsoGL:     mapping.ConsoGL,
						Year:        year,
					})

					// 2. Aggregate for Fact Table
					if len(row.PostingDate) >= 7 {
						monCode := row.PostingDate[5:7]
						if mon, ok := monthToCodeMap[monCode]; ok {
							k := AggKey{
								Entity: company, Branch: branch, Dept: deptCode, NavCode: row.Department,
								EntityGL: row.EntityGL, ConsoGL: mapping.ConsoGL, GLName: mapping.AccountName,
								VendorName: row.Vendor, Month: mon,
							}
							mergedFactMap[k] = mergedFactMap[k].Add(row.Amount)
						}
					}
				}
			}

			processRowsBatch(hmwRows)
			processRowsBatch(clikRows)

			// 3. Save Transactions IMMEDIATELY to free memory
			if len(transactions) > 0 {
				if err := trxRepo.CreateActualTransactions(transactions); err != nil {
					return err
				}
				transactions = nil // Help GC
			}
			hmwRows = nil
			clikRows = nil
		}

		// 5. Save Aggregated Facts (Re-build Headers)
		headerMap := make(map[HeaderKey][]models.ActualAmountEntity)
		for k, amt := range mergedFactMap {
			hk := HeaderKey{k.Entity, k.Branch, k.Dept, k.NavCode, k.EntityGL, k.ConsoGL, k.GLName, k.VendorName}
			headerMap[hk] = append(headerMap[hk], models.ActualAmountEntity{
				ID:     uuid.New(),
				Month:  k.Month,
				Amount: amt,
			})
		}

		var headers []models.ActualFactEntity
		for k, amounts := range headerMap {
			headerID := uuid.New()
			for i := range amounts {
				amounts[i].ActualFactID = headerID
			}
			total := decimal.Zero
			for _, a := range amounts {
				total = total.Add(a.Amount)
			}
			headers = append(headers, models.ActualFactEntity{
				ID:            headerID,
				Entity:        k.Entity,
				Branch:        k.Branch,
				Department:    k.Dept,
				NavCode:       k.NavCode,
				EntityGL:      k.EntityGL,
				ConsoGL:       k.ConsoGL,
				GLName:        k.GLName,
				VendorName:    k.VendorName,
				Group:         glProfileMap[k.ConsoGL],
				Year:          year,
				YearTotal:     total,
				ActualAmounts: amounts,
			})
		}

		if len(headers) > 0 {
			const chunkSize = 100
			for i := 0; i < len(headers); i += chunkSize {
				end := i + chunkSize
				if end > len(headers) {
					end = len(headers)
				}
				if err := trxRepo.CreateActualFacts(headers[i:end]); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *actualService) DeleteActualFacts(year string) error {
	if year == "" {
		return fmt.Errorf("year is required")
	}
	return s.repo.DeleteActualFactsByYear(year)
}

func (s *actualService) GetRawDate() (string, error) {
	return s.repo.GetRawDate()
}
