package service

import (
	"context"
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

func (s *actualService) SyncActuals(ctx context.Context, year string, months []string) error {
	fmt.Printf("[DEBUG] SyncActuals: Start DB Sync (Optimized Batch) for Year %s, Months %v...\n", year, months)

	// 1. Fetch Unified GL Grouping (Mapping & Hierarchy)
	groupings, err := s.masterSrv.ListGLGroupings(ctx)
	if err != nil {
		return fmt.Errorf("actualService.SyncActuals: %w", err)
	}

	mappingMap := make(map[string]models.GlGroupingEntity)
	glProfileMap := make(map[string]string) // ConsoGL -> Group1 (Profile)

	for _, g := range groupings {
		if g.IsActive {
			// Specific mapping: Normalized Company + GL
			normEntity := NormalizeEntityCode(g.Entity)
			key := fmt.Sprintf("%s_%s", normEntity, g.EntityGL)
			mappingMap[key] = g

			// Store ConsoGL -> Profile (Group 1)
			if g.ConsoGL != "" && g.Group1 != "" {
				glProfileMap[g.ConsoGL] = g.Group1
			}
		}
	}

	// 2. Define Mapping Configs (Consolidated in common_service)
	branchNameMap := map[string]string{
		"BURIRUM": "BUR", "HEAD OFFICE": "HOF", "KRABI": "KBI",
		"MINI_SURIN": "MSR", "MUEANG KRABI": "MKB", "NAKA": "NAK",
		"NANGRONG": "AVN", "PHACHA": "PHC", "PHUKET": "PRA",
		"SURIN": "SUR", "VEERAWAT": "VEE",
		"AUTOCORP HEAD OFFICE": "HQ",
		"":                     "Branch00",
		"BRANCH01":             "Branch01", "BRANCH02": "Branch02", "BRANCH03": "Branch03",
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

	// 3. (Legacy) GL Profile Map is now built during Grouping fetch above.

	// 4. Persistence with Transaction & Batching
	return s.repo.WithTrx(func(trxRepo models.ActualRepository) error {
		// Preserve non-PENDING statuses (CONFIRMED, COMPLETE, REPORTED, etc.) before deleting
		preservedStatuses, err := trxRepo.GetNonPendingTransactionKeys(ctx, year, targetMonths)
		if err != nil {
			fmt.Printf("[WARN] Failed to get preserved statuses: %v\n", err)
			preservedStatuses = nil
		} else if len(preservedStatuses) > 0 {
			fmt.Printf("[Sync] Preserving %d non-PENDING transaction statuses\n", len(preservedStatuses))
		}

		if fullYearSync {
			// Wipe whole year once
			if err := trxRepo.DeleteActualFactsByYear(ctx, year); err != nil {
				return fmt.Errorf("transaction.DeleteFactsByYear: %w", err)
			}
			if err := trxRepo.DeleteActualTransactionsByYear(ctx, year); err != nil {
				return fmt.Errorf("transaction.DeleteTxsByYear: %w", err)
			}
		} else {
			// Wipe target months only
			for _, m := range targetMonths {
				if err := trxRepo.DeleteActualFactsByMonth(ctx, year, m); err != nil {
					return fmt.Errorf("transaction.DeleteFactsByMonth(%s): %w", m, err)
				}
				if err := trxRepo.DeleteActualTransactionsByMonth(ctx, year, m); err != nil {
					return fmt.Errorf("transaction.DeleteTxsByMonth(%s): %w", m, err)
				}
			}
		}

		// Keep headers aggregation in memory (Aggregated data is safe)
		mergedFactMap := make(map[AggKey]decimal.Decimal)

		for _, mName := range targetMonths {
			fmt.Printf("[Sync] Processing Month: %s\n", mName)

			// Fetch one month
			hmwRows, err := trxRepo.GetRawTransactionsHMW(ctx, year, []string{mName})
			if err != nil {
				return fmt.Errorf("transaction.GetHMWRows(%s): %w", mName, err)
			}
			clikRows, err := trxRepo.GetRawTransactionsCLIK(ctx, year, []string{mName})
			if err != nil {
				return fmt.Errorf("transaction.GetCLIKRows(%s): %w", mName, err)
			}

			totalRows := 0
			mappedRows := 0
			filteredRows := 0

			var transactions []models.ActualTransactionEntity

			processRowsBatch := func(rows []models.ActualTransactionDTO) {
				for _, row := range rows {
					company := NormalizeEntityCode(row.Company)
					// 1. Look up STRICT mapping (Company + GL) - NO GLOBAL FALLBACK
					key := fmt.Sprintf("%s_%s", company, row.EntityGL)
					mapping, ok := mappingMap[key]

					totalRows++

					// FILTER: Only sync if the GL is specifically mapped for this company
					if !ok || mapping.ConsoGL == "" {
						filteredRows++

						continue
					}

					mappedRows++

					// We now have the best possible mapping (Specific > Global)

					branch := mapToCode(row.Branch, branchNameMap)
					deptCode := row.Department
					lookupDept := normalize(row.Department)
					if masterDept, err := s.depSrv.GetMasterDepartment(ctx, lookupDept, company); err == nil && masterDept != nil {
						deptCode = masterDept.Code
					} 

					// 1. Transaction Table (Centralized Detail)
					transactions = append(transactions, models.ActualTransactionEntity{
						ID:            uuid.New(),
						Source:        row.Source,
						PostingDate:   row.PostingDate,
						DocNo:         row.DocNo,
						Description:   row.Description,
						Amount:        row.Amount,
						VendorName:    row.Vendor,
						Entity:        company,
						Branch:        branch,
						Department:    deptCode,
						EntityGL:      row.EntityGL,
						ConsoGL:       mapping.ConsoGL,
						Year:          year,
						GLAccountName: mapping.AccountName, // Use the name from mapping table as requested
					})

					// 2. Aggregate for Fact Table
					if len(row.PostingDate) >= 7 {
						// Robust Month Extraction: Find any 2-digit number between separators
						// E.g., 2026-03-25 or 25/03/2026 or 2026/03/25
						monCode := ""
						parts := strings.FieldsFunc(row.PostingDate, func(r rune) bool {
							return r == '-' || r == '/' || r == '.'
						})
						for _, p := range parts {
							if len(p) == 2 && p != "20" && p != "26" { // Heuristic: Month is a 2-digit part excluding common years
								monCode = p

								break
							}
						}
						// Fallback to substring if heuristic fails (but is standard ISO)
						if monCode == "" && strings.Contains(row.PostingDate, "-") && len(row.PostingDate) >= 7 {
							monCode = row.PostingDate[5:7]
						}

						if mon, ok := monthToCodeMap[monCode]; ok {
							// Determine GL Name: Use Mapped name if exists, else Original name
							glName := mapping.AccountName
							if glName == "" {
								glName = row.GLAccountName
							}

							k := AggKey{
								Entity: company, Branch: branch, Dept: deptCode, NavCode: row.Department,
								EntityGL: row.EntityGL, ConsoGL: mapping.ConsoGL, GLName: glName,
								VendorName: row.Vendor, Month: mon,
							}
							mergedFactMap[k] = mergedFactMap[k].Add(row.Amount)
						}
					}
				}
			}

			processRowsBatch(hmwRows)
			processRowsBatch(clikRows)

			fmt.Printf("[Sync] Month %s Result: Total %d, Mapped %d, Filtered %d (Optimized)\n", mName, totalRows, mappedRows, filteredRows)

			// 3. Save Transactions IMMEDIATELY to free memory
			if len(transactions) > 0 {
				const chunkSize = 500 // 🍰 แบ่งส่งทีละ 500 รายการ
				for i := 0; i < len(transactions); i += chunkSize {
					end := i + chunkSize
					if end > len(transactions) {
						end = len(transactions)
					}
					// ส่งทีละก้อนเล็กๆ แทนก้อนใหญ่ก้อนเดียว
					if err := trxRepo.CreateActualTransactions(ctx, transactions[i:end]); err != nil {
						return fmt.Errorf("transaction.CreateTransactionsChunk: %w", err)
					}
				}
				transactions = nil
			}
			// if len(transactions) > 0 {
			// 	if err := trxRepo.CreateActualTransactions(ctx, transactions); err != nil {
			// 		return fmt.Errorf("transaction.CreateTransactions(%s): %w", mName, err)
			// 	}
			// 	transactions = nil // Help GC
			// }
			// hmwRows = nil
			// clikRows = nil
		}

		// 4.5 Restore preserved statuses for transactions that were previously confirmed/completed
		if len(preservedStatuses) > 0 {
			if err := trxRepo.RestoreTransactionStatuses(ctx, preservedStatuses); err != nil {
				fmt.Printf("[WARN] Failed to restore transaction statuses: %v\n", err)
			} else {
				fmt.Printf("[Sync] Successfully restored preserved statuses\n")
			}
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
				if err := trxRepo.CreateActualFacts(ctx, headers[i:end]); err != nil {
					return fmt.Errorf("transaction.CreateFactsChunk: %w", err)
				}
			}
		}

		// Update Data Inventory Metadata
		if err := trxRepo.RefreshDataInventory(ctx); err != nil {
			fmt.Printf("[WARN] Failed to refresh data inventory: %v\n", err)
		}

		return nil
	})
}

func (s *actualService) DeleteActualFacts(ctx context.Context, year string) error {
	if year == "" {
		return fmt.Errorf("actualService.DeleteActualFacts: year is required")
	}
	if err := s.repo.DeleteActualFactsByYear(ctx, year); err != nil {
		return fmt.Errorf("actualService.DeleteActualFacts: %w", err)
	}
	// Refresh inventory after deletion
	_ = s.repo.RefreshDataInventory(ctx)
	return nil
}

func (s *actualService) GetRawDate(ctx context.Context) (string, error) {
	return s.repo.GetRawDate(ctx)
}

func (s *actualService) RefreshDataInventory(ctx context.Context) error {
	return s.repo.RefreshDataInventory(ctx)
}

// func (s *actualService) SyncActualsDebug(ctx context.Context, targetDocNo string) error {
//     fmt.Printf("\n--- [🕵️ TARGETED DEBUG] Doc: %s ---\n", targetDocNo)

//     groupings, _ := s.masterSrv.ListGLGroupings(ctx)
//     mappingMap := make(map[string]models.GlGroupingEntity)

//     fmt.Printf("📦 [STEP 1: DB SCANNING]\n")
//     for _, g := range groupings {
//         if g.IsActive && strings.Contains(g.EntityGL, "51310010") {
//             // เราจะ Normalize ทั้งสองฝั่งเพื่อดูความเพี้ยน
//             normEntity := NormalizeEntityCode(g.Entity)
//             key := fmt.Sprintf("%s_%s", normEntity, g.EntityGL)
//             mappingMap[key] = g

//             fmt.Printf("   📌 DB Mapping -> Entity: [%s] | GL: [%s] | Norm: [%s] | Key: [%s]\n",
//                 g.Entity, g.EntityGL, normEntity, key)
//         }
//     }

//     // 🎯 แก้จุดนี้: บิลใบใหม่คือ 2502-0010 (ปี 2025 เดือน FEB)
//     year := "2026"
//     months := []string{"FEB"}

//     return s.repo.WithTrx(func(trxRepo models.ActualRepository) error {
//         for _, mName := range months {
//             hmwRows, _ := trxRepo.GetRawTransactionsHMW(ctx, year, []string{mName})
//             allRows := hmwRows

//             for _, row := range allRows {
//                 if row.DocNo != targetDocNo { continue }

//                 fmt.Printf("\n⚡ [STEP 2: RUNTIME CHECK]\n")
//                 fmt.Printf("   📑 Data ดิบ -> Company: [%s] | GL: [%s]\n", row.Company, row.EntityGL)

//                 company := NormalizeEntityCode(row.Company)
//                 runtimeKey := fmt.Sprintf("%s_%s", company, row.EntityGL)

//                 fmt.Printf("   🎯 ระบบพยายามหา Key: [%s]\n", runtimeKey)

//                 mapping, ok := mappingMap[runtimeKey]
//                 if !ok {
//                     fmt.Printf("   ❌ RESULT -> ไม่เจอ! เพราะ [%s] ไม่ตรงกับ Key ใน DB ด้านบน\n", runtimeKey)

//                     // สแกนหาตัวใกล้เคียงเพื่อฟันธง
//                     for k := range mappingMap {
//                         if strings.Contains(k, row.EntityGL) {
//                             fmt.Printf("      💡 ใน Memory มี Key นี้: [%s] แต่คุณหา [%s] มันเลยไม่เจอกัน!\n", k, runtimeKey)
//                         }
//                     }
//                 } else {
//                     fmt.Printf("   ✅ RESULT -> เจอแล้ว! ConsoGL: %s\n", mapping.ConsoGL)
//                 }
//             }
//         }
//         return nil
//     })
// }

// func (s *actualService) SyncActualsDebug(ctx context.Context, targetDocNo string) error {
//     fmt.Printf("\n--- [🕵️ TARGETED DEBUG] Doc: %s ---\n", targetDocNo)

//     // ดึง Mapping ทั้งหมดมาก่อน
//     groupings, _ := s.masterSrv.ListGLGroupings(ctx)
//     mappingMap := make(map[string]models.GlGroupingEntity)

//     fmt.Printf("📦 [STEP 1: DB SCANNING]\n")
//     for _, g := range groupings {
//         // กรองเฉพาะ GL ที่เราสงสัยเพื่อลด noise ใน log (51310010)
//         if g.IsActive && strings.Contains(g.EntityGL, "51310001") {
//             normEntity := NormalizeEntityCode(g.Entity)
//             key := fmt.Sprintf("%s_%s", normEntity, g.EntityGL)
//             mappingMap[key] = g

//             fmt.Printf("   📌 DB Mapping -> Entity: [%s] | GL: [%s] | Norm: [%s] | Key: [%s]\n",
//                 g.Entity, g.EntityGL, normEntity, key)
//         }
//     }

//     // 🎯 วิเคราะห์จากเลขบิล C00-PVL2602-0082: 26 คือปี 2026, 02 คือ FEB
//     year := "2026"
//     months := []string{"FEB"}

//     return s.repo.WithTrx(func(trxRepo models.ActualRepository) error {
//         for _, mName := range months {
//             // 1. ดึงจากทั้งสอง Table (HMW และ CLIK)
//             hmwRows, _ := trxRepo.GetRawTransactionsHMW(ctx, year, []string{mName})
//             clikRows, _ := trxRepo.GetRawTransactionsCLIK(ctx, year, []string{mName})

//             // รวมข้อมูลเข้าด้วยกัน
//             allRows := append(hmwRows, clikRows...)

//             foundAny := false
//             for _, row := range allRows {
//                 // เช็ค DocNo (ใช้ strings.Contains เผื่อมี Space หรือ Prefix ต่างกันเล็กน้อย)
//                 if !strings.Contains(row.DocNo, targetDocNo) {
//                     continue
//                 }
//                 foundAny = true

//                 fmt.Printf("\n⚡ [STEP 2: RUNTIME CHECK] (Source Table: %s)\n", row.Source)
//                 fmt.Printf("   📑 Data ดิบ -> Company: [%s] | GL: [%s] | Doc: [%s]\n",
//                     row.Company, row.EntityGL, row.DocNo)

//                 // กระบวนการเดียวกับใน Actual Service จริงๆ
//                 company := NormalizeEntityCode(row.Company)
//                 runtimeKey := fmt.Sprintf("%s_%s", company, row.EntityGL)

//                 fmt.Printf("   🎯 ระบบพยายามหา Key: [%s]\n", runtimeKey)

//                 mapping, ok := mappingMap[runtimeKey]
//                 if !ok {
//                     fmt.Printf("   ❌ RESULT -> ไม่เจอใน Mapping! เพราะ Key [%s] ไม่ตรงกับใน DB\n", runtimeKey)

//                     // ช่วยหาว่าในระบบมี GL นี้ภายใต้ชื่อบริษัทอื่นไหม
//                     for k := range mappingMap {
//                         if strings.Contains(k, row.EntityGL) {
//                             fmt.Printf("      💡 คำแนะนำ: ใน DB มี Key [%s] ลองเช็คว่าบริษัทใน Data ดิบส่งมาผิดหรือเปล่า?\n", k)
//                         }
//                     }
//                 } else {
//                     fmt.Printf("   ✅ RESULT -> เจอคู่ Matching! รายการนี้จะถูกส่งไปที่ ConsoGL: [%s]\n", mapping.ConsoGL)
//                 }
//             }

//             if !foundAny {
//                 fmt.Printf("\n⚠️ [RESULT] ไม่พบเลขบิล [%s] ใน Table ของเดือน %s ปี %s เลย (ลองเช็ค Year/Month อีกครั้ง)\n",
//                     targetDocNo, mName, year)
//             }
//         }
//         return nil
//     })
// }

// func (s *actualService) SyncActualsDebug(ctx context.Context, targetDocNo string) error {
//     fmt.Printf("\n--- [🕵️ TARGETED DEBUG] Doc: %s ---\n", targetDocNo)

//     // 1. ดึง Mapping ทั้งหมดมาเก็บไว้ (เอาเงื่อนไข IF ออกเพื่อให้ Match ได้ทุก GL)
//     groupings, _ := s.masterSrv.ListGLGroupings(ctx)
//     mappingMap := make(map[string]models.GlGroupingEntity)

//     for _, g := range groupings {
//         if g.IsActive {
//             normEntity := NormalizeEntityCode(g.Entity)
//             key := fmt.Sprintf("%s_%s", normEntity, g.EntityGL)
//             mappingMap[key] = g
//         }
//     }

//     year := "2026"
//     months := []string{"JAN", "FEB", "MAR", "APR"}

//     return s.repo.WithTrx(func(trxRepo models.ActualRepository) error {
//         for _, mName := range months {
//             hmwRows, _ := trxRepo.GetRawTransactionsHMW(ctx, year, []string{mName})
//             clikRows, _ := trxRepo.GetRawTransactionsCLIK(ctx, year, []string{mName})
//             allRows := append(hmwRows, clikRows...)

//             foundAny := false
//             fmt.Printf("📦 [STEP 2: SCANNING ALL ENTRIES IN THIS BILL]\n")

//             for _, row := range allRows {
//                 // กรองเฉพาะเลขบิลที่ต้องการ
//                 if !strings.Contains(row.DocNo, targetDocNo) {
//                     continue
//                 }
//                 foundAny = true

//                 // --- ส่วนที่เพิ่ม/แก้ไขเพื่อให้เห็นข้อมูลชัดขึ้น ---
//                 company := NormalizeEntityCode(row.Company)
//                 runtimeKey := fmt.Sprintf("%s_%s", company, row.EntityGL)

//                 fmt.Printf("\n🔍 [ENTRY FOUND]")
//                 fmt.Printf("\n   ├─ Source Table: %s", row.Source)
//                 fmt.Printf("\n   ├─ GL Account:   [%s] <--- เช็คเลขนี้!", row.EntityGL)
//                 fmt.Printf("\n   ├─ GL Name:      %s", row.GLAccountName)
//                 fmt.Printf("\n   ├─ Description:  %s", row.Description)
//                 fmt.Printf("\n   ├─ Amount:       %.2f", row.Amount)
//                 fmt.Printf("\n   └─ Runtime Key:  [%s]", runtimeKey)

//                 // ตรวจสอบว่า GL บรรทัดนี้ มีในระบบเราไหม
//                 if mapping, ok := mappingMap[runtimeKey]; ok {
//                     fmt.Printf("\n   ✅ STATUS: MATCHED! -> จะถูกเข้า Grouping: [%s]\n", mapping.AccountName)
//                 } else {
//                     fmt.Printf("\n   ❌ STATUS: NOT MAPPED! -> เลข GL นี้ไม่ได้ตั้งค่าไว้ในระบบ\n")
//                 }
//             }

//             if !foundAny {
//                 fmt.Printf("\n⚠️ [RESULT] ไม่พบเลขบิล [%s] ในฐานข้อมูลเลย\n", targetDocNo)
//             }
//         }
//         return nil
//     })
// }
