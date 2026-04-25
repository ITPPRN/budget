package servers

import (
	"context"
	"fmt"
	"time"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"

	"github.com/google/uuid"
)

// skippedTier1Count tracks consecutive Tier 1 skips (stale lock indicator)
var skippedTier1Count int

// recordSyncRun helper — บันทึก sync_run แบบ convenient
// ถ้า tracking ไม่ถูกตั้งค่าจะคืน uuid.Nil และไม่ error
func (s *server) recordSyncRunStart(jobType, year, month, triggeredBy string) uuid.UUID {
	if s.Shd.SyncTrackingRepo == nil {
		return uuid.Nil
	}
	run := &models.SyncRunEntity{
		JobType:     jobType,
		Year:        year,
		Month:       month,
		TriggeredBy: triggeredBy,
	}
	if err := s.Shd.SyncTrackingRepo.CreateRun(context.Background(), run); err != nil {
		logs.Errorf("Failed to record sync run: %v", err)
		return uuid.Nil
	}
	return run.ID
}

func (s *server) recordSyncRunComplete(id uuid.UUID, err error) {
	if s.Shd.SyncTrackingRepo == nil || id == uuid.Nil {
		return
	}
	status := models.SyncStatusSuccess
	errMsg := ""
	if err != nil {
		status = models.SyncStatusFailed
		errMsg = err.Error()
	}
	_ = s.Shd.SyncTrackingRepo.CompleteRun(context.Background(), id, status, 0, 0, 0, errMsg)
}

func (s *server) StartCronJob() {
	logs.Info("⏰ Initializing Cron Jobs...")

	// Startup: clear stale RUNNING sync_runs ที่ค้างจาก previous crash (>2 hours old)
	if s.Shd.SyncTrackingRepo != nil {
		if cleared, err := s.Shd.SyncTrackingRepo.ClearStaleRunningRuns(context.Background(), 2*time.Hour); err != nil {
			logs.Errorf("Startup: Failed to clear stale sync runs: %v", err)
		} else if cleared > 0 {
			logs.Warnf("⚠️ Startup: Cleared %d stale RUNNING sync_runs (previous server crash)", cleared)
		}
	}

	// 1. Job: Tier 1 - Fast Sync (Every 5 mins)
	// Sync only the current month of the current year for real-time reactivity.
	if _, err := s.Cron.AddFunc("0/5 * * * *", func() {
		// TryLock: ถ้ามี sync ใหญ่กำลังรันอยู่ ให้ skip รอบนี้ไป ไม่ต้องรอ queue
		if !s.SyncMutex.TryLock() {
			skippedTier1Count++
			logs.Info("⏰ Tier 1 Job: Skipped (another sync is running)")
			// Alert if too many consecutive skips (possible stuck lock)
			if skippedTier1Count >= 12 { // 12 × 5min = 1 hour of skipping
				logs.Warnf("⚠️ Tier 1 Job: Skipped %d times consecutively — possible stuck sync mutex!", skippedTier1Count)
			}
			return
		}
		skippedTier1Count = 0
		defer s.SyncMutex.Unlock()

		now := time.Now()
		yearStr := now.Format("2006")
		monCode := now.Format("01")

		monthMap := map[string]string{
			"01": "JAN", "02": "FEB", "03": "MAR", "04": "APR", "05": "MAY", "06": "JUN",
			"07": "JUL", "08": "AUG", "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
		}
		mName := monthMap[monCode]

		runID := s.recordSyncRunStart(models.SyncJobTier1Fast, yearStr, mName, "CRON")

		logs.Infof("⏰ Tier 1 Job: Fast-Sync Current Month (%s %s) Started", mName, yearStr)
		err := s.Shd.ActualService.SyncActuals(context.Background(), yearStr, []string{mName})
		s.recordSyncRunComplete(runID, err)
		if err != nil {
			logs.Errorf("Tier 1 Job: Fast-Sync Failed: %v", err)
			return
		}
		logs.Info("⏰ Tier 1 Job: Fast-Sync Completed Successfully")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 1 Cron Job: %v", err))
	}

	// 2. Job: Tier 2 - Full Maintenance Sync (Daily @ 02:00 AM)
	// Synchronize all data from 2025 to Present to ensure full consistency during off-peak hours.
	if _, err := s.Cron.AddFunc("0 2 * * *", func() {
		s.SyncMutex.Lock()
		defer s.SyncMutex.Unlock()

		logs.Info("⏰ Tier 2 Job: Full Maintenance Sync Started")

		currentYear := time.Now().Year()
		startYear := currentYear - 1

		runID := s.recordSyncRunStart(models.SyncJobTier2Full, fmt.Sprintf("%d-%d", startYear, currentYear), "", "CRON")

		var syncErr error
		for year := startYear; year <= currentYear; year++ {
			yearStr := fmt.Sprintf("%d", year)
			logs.Infof("⏰ Tier 2 Job: Syncing Full Year %s...", yearStr)

			// SyncActuals handles months internally batch-by-batch if passed empty months
			if err := s.Shd.ActualService.SyncActuals(context.Background(), yearStr, []string{}); err != nil {
				logs.Errorf("Tier 2 Job: Failed to sync year %s: %v", yearStr, err)
				syncErr = err
			}
		}
		s.recordSyncRunComplete(runID, syncErr)
		if syncErr != nil {
			logs.Error("⏰ Tier 2 Job: Full Maintenance Sync Completed with Errors")
			return
		}
		logs.Info("⏰ Tier 2 Job: Full Maintenance Sync Completed")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 2 Cron Job: %v", err))
	}

	// 3. Retry Job — ทุก 30 นาที: ดึง FAILED runs ใน 24 ชม. ที่ retry < 3 มาทำใหม่
	if s.Shd.SyncTrackingRepo != nil {
		if _, err := s.Cron.AddFunc("*/30 * * * *", func() {
			if !s.SyncMutex.TryLock() {
				return // skip if main sync running
			}
			defer s.SyncMutex.Unlock()

			failed, err := s.Shd.SyncTrackingRepo.GetFailedRunsForRetry(context.Background(), 24*time.Hour, 3)
			if err != nil || len(failed) == 0 {
				return
			}

			logs.Infof("🔁 Retry Job: found %d failed run(s) eligible for retry", len(failed))
			for _, run := range failed {
				logs.Infof("🔁 Retry: %s year=%s month=%s retry=%d", run.JobType, run.Year, run.Month, run.RetryCount+1)
				_ = s.Shd.SyncTrackingRepo.IncrementRetry(context.Background(), run.ID)

				retryID := s.recordSyncRunStart(run.JobType, run.Year, run.Month, "RETRY:"+run.ID.String()[:8])

				var retryErr error
				switch run.JobType {
				case models.SyncJobTier1Fast:
					if run.Year != "" && run.Month != "" {
						retryErr = s.Shd.ActualService.SyncActuals(context.Background(), run.Year, []string{run.Month})
					}
				case models.SyncJobTier2Full, models.SyncJobActualFact:
					if run.Year != "" {
						retryErr = s.Shd.ActualService.SyncActuals(context.Background(), run.Year, []string{})
					}
				case models.SyncJobDW:
					if s.Shd.ExternalSyncService != nil {
						retryErr = s.Shd.ExternalSyncService.SyncFromDW(context.Background())
					}
				}

				s.recordSyncRunComplete(retryID, retryErr)
				if retryErr != nil {
					logs.Errorf("🔁 Retry Failed: %s: %v", run.JobType, retryErr)
				} else {
					logs.Infof("🔁 Retry Success: %s", run.JobType)
				}
			}
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register Retry Cron Job: %v", err))
		}
	}

	// 2. Job: Department Seeding (Run once at startup, or could be a job)
	// For P2P, we keep the "Seed" logic but can trigger it via Cron or just once here.
	go func() {
		logs.Info("System: Starting Department Data Seeding...")
		if err := s.Shd.DepartmentService.ManageDepartments(context.Background()); err != nil {
			logs.Error("Failed to Seed Departments: " + err.Error())
		} else {
			logs.Info("System: Department Data Seeding Completed Successfully")
		}
	}()

	// 3. Initial Broadcast Requests (RabbitMQ)
	// Ask Example Service for data startup
	if s.Shd.ProducerService != nil {
		go func() {
			logs.Info("🚀 P2P: Sending initial data sync requests (Broadcast Begin)...")

			if err := s.Shd.ProducerService.CompanyBegin(&events.MessageCompaniesBeginEvent{}); err != nil {
				logs.Warnf("Failed to request Company Sync: %v", err)
			}
			if err := s.Shd.ProducerService.DepartmentBegin(&events.MessageDepartmentBeginEvent{}); err != nil {
				logs.Warnf("Failed to request Department Sync: %v", err)
			}
			if err := s.Shd.ProducerService.UserBegin(&events.MessageUserBeginEvent{}); err != nil {
				logs.Warnf("Failed to request User Sync: %v", err)
			}

			// Also push our own local data if anyone is listening
			s.Shd.MasterService.BroadcastAllData(context.Background())
		}()
	}

	// 4. Job: DW Auto-Sync (Scheduled Daily @ Midnight)
	if s.Shd.ExternalSyncService != nil {
		// --- Scheduled Job ---
		if _, err := s.Cron.AddFunc("0 0 * * *", func() {
			s.SyncMutex.Lock()
			defer s.SyncMutex.Unlock()

			logs.Info("⏰ Job: DW Auto-Sync Started (Daily @ Midnight)")
			if err := s.Shd.ExternalSyncService.SyncFromDW(context.Background()); err != nil {
				logs.Errorf("Job: DW Auto-Sync Failed: %v", err)
				return
			}

			// Finalize: Refresh Data Inventory Metadata for Admin UI
			if err := s.Shd.ActualService.RefreshDataInventory(context.Background()); err != nil {
				logs.Errorf("Job: DW Auto-Sync Failed to refresh inventory: %v", err)
				return
			}
			logs.Info("⏰ Job: DW Auto-Sync Completed")
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register DW Sync Cron Job: %v", err))
		}

		// 🚀 --- TEMPORARY: IMMEDIATE STARTUP SYNC (FOR TESTING) ---
		// TODO: Remove this block before production deployment.
		// This will run ONCE right now! Delete this block after testing is done.
		go func() {
			s.SyncMutex.Lock()
			defer s.SyncMutex.Unlock()

			ctx := context.Background()
			logs.Info("🚀 IMMEDIATE STARTUP SYNC: STARTING NOW (DW -> MAPPING)...")

			// 1. Sync from Data Warehouse (Raw CLIK/ACHHMW data)
			if err := s.Shd.ExternalSyncService.SyncFromDW(ctx); err != nil {
				logs.Errorf("🚀 Startup Sync: DW Failed: %v", err)
				return
			}

			// 2. Refresh Mapping & Facts
			now := time.Now()
			for y := now.Year(); y <= now.Year(); y++ {
				yStr := fmt.Sprintf("%d", y)
				logs.Infof("🚀 Startup Sync: Mapping Year %s...", yStr)
				if err := s.Shd.ActualService.SyncActuals(ctx, yStr, []string{}); err != nil {
					logs.Errorf("🚀 Startup Sync: Mapping %s Failed: %v", yStr, err)
					return
				}
			}

			// 3. Finalize Metadata
			if err := s.Shd.ActualService.RefreshDataInventory(ctx); err != nil {
				logs.Errorf("🚀 Startup Sync: RefreshDataInventory Failed: %v", err)
				return
			}
			logs.Info("🚀 IMMEDIATE STARTUP SYNC: COMPLETED SUCCESSFULLY")
		}()
		// 🚀 --- END TEMPORARY BLOCK ---
	}

	// Start the Cron scheduler
	s.Cron.Start()
	logs.Info("⏰ Cron Scheduler Started")
}


// func (s *server) StartCronJob() {
//     logs.Info("⚠️  DEBUG MODE: Cron Jobs are DISABLED. Running Targeted Sync Debug...")

//     // 1. รัน Debug ใน Goroutine แยก
//     go func() {
//         // ใช้เวลาหลับสัก 5 วินาที เพื่อรอให้ระบบอื่นๆ (เช่น RabbitMQ/DB) นิ่งก่อน
//         time.Sleep(5 * time.Second)

//         s.SyncMutex.Lock()
//         defer s.SyncMutex.Unlock()

//         ctx := context.Background()

//         // 🎯 ระบุเลขบิลเป้าหมาย (เช็คปี 2026 เดือน APR ในฟังก์ชัน Debug ด้วยนะครับ!)
//         targetDoc := "C00-PVL2602-0082"

//         logs.Infof("🚀 [DEBUG-ONLY] STARTING: Checking Document No. %s...", targetDoc)

//         // เรียกฟังก์ชันที่เราโคลนไว้
//         if err := s.Shd.ActualService.SyncActualsDebug(ctx, targetDoc); err != nil {
//             logs.Errorf("❌ [DEBUG-ONLY] FAILED: %v", err)
//         }

//         logs.Info("🏁 [DEBUG-ONLY] COMPLETED: ตรวจสอบ Log ด้านบนเพื่อหาจุดที่ 'continue' (ข้าม)")
//     }()

//     // 🛑 บรรทัด s.Cron.Start() ต้องถูกคอมเมนต์ไว้ "เท่านั้น" ในตอนดีบัก!
//     logs.Warn("⏰ CRON SCHEDULER IS STOPPED: ระบบจะไม่รัน Sync ปกติจนกว่าจะเปิดบรรทัดนี้")
//     // s.Cron.Start() // <--- ห้ามเปิดเด็ดขาดตอนดีบัก!
// }





// func (s *server) StartCronJob() {
//     logs.Info("⚠️  DEBUG MODE: Cron Jobs are DISABLED. Running Targeted Sync Debug...")

//     go func() {
//         // รอให้ระบบ Network/DB Ready
//         time.Sleep(5 * time.Second)

//         // ล็อค Mutex เพื่อจำลองสภาวะการทำงานจริงของ Sync
//         s.SyncMutex.Lock()
//         defer s.SyncMutex.Unlock()

//         ctx := context.Background()

//         // 🎯 TARGET: บิลใบนี้คือปี 2026 เดือน 02 (FEB)
//         targetDoc := "C00-PVL2602-0082"

//         logs.Infof("🚀 [DEBUG-ONLY] STARTING: Checking Document No. %s...", targetDoc)
//         logs.Info("📅 [DEBUG-INFO] ระบบจะค้นหาใน Year: 2026 | Month: FEB (อ้างอิงตามเลขบิล)")

//         if err := s.Shd.ActualService.SyncActualsDebug(ctx, targetDoc); err != nil {
//             logs.Errorf("❌ [DEBUG-ONLY] FAILED: %v", err)
//         }

//         logs.Info("🏁 [DEBUG-ONLY] COMPLETED: ตรวจสอบ Log ด้านบนเพื่อดูผลลัพธ์ STEP 1 และ STEP 2")
//     }()

//     // 🛑 ปิด Cron ปกติไว้เพื่อไม่ให้ Log ตีกัน
//     logs.Warn("⏰ CRON SCHEDULER IS STOPPED: Running ONLY Targeted Debug Mode")
//     // s.Cron.Start() 
// }