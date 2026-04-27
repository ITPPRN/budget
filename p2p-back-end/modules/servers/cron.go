package servers

import (
	"context"
	"fmt"
	"time"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
	syncQueue "p2p-back-end/modules/external_sync/queue"
)

// enqueueDWPerMonth fans out a DW sync request across the rolling window of
// past 17 months (current year-1 through current month) — one queue entry per
// month. Worker still drains them serially, but failures are isolated per
// month and the retry job can re-enqueue only what failed.
func (s *server) enqueueDWPerMonth(triggeredBy string) {
	monthCodes := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	now := time.Now()
	currentYear := now.Year()
	startYear := currentYear - 1
	currentMonth := int(now.Month())
	for year := startYear; year <= currentYear; year++ {
		endMonth := 12
		if year == currentYear {
			endMonth = currentMonth
		}
		yearStr := fmt.Sprintf("%d", year)
		for m := 1; m <= endMonth; m++ {
			s.enqueueOrLog(models.SyncJobDW, yearStr, []string{monthCodes[m-1]}, triggeredBy)
		}
	}
}

// enqueueOrLog enqueues a sync job and logs the outcome.
// All cron handlers route through here so the single Worker serializes execution.
func (s *server) enqueueOrLog(jobType, year string, months []string, triggeredBy string) {
	if s.Shd.SyncQueue == nil {
		logs.Warnf("⏰ %s: queue not configured, skipping", jobType)
		return
	}
	job := &syncQueue.Job{
		JobType:     jobType,
		Year:        year,
		Months:      months,
		TriggeredBy: triggeredBy,
	}
	enqueued, err := s.Shd.SyncQueue.Enqueue(context.Background(), job)
	if err != nil {
		logs.Errorf("⏰ %s: enqueue failed: %v", jobType, err)
		return
	}
	if !enqueued {
		logs.Infof("⏰ %s: skipped (identical job already queued or running)", jobType)
		return
	}
	logs.Infof("⏰ %s: queued (id=%s, year=%s, months=%v, by=%s)",
		jobType, job.ID, year, months, triggeredBy)
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

	// 1. Job: Tier 1 - Fast Sync (Every 30 mins) — SKIP if queue busy.
	// Tier 1 is the lowest priority and we don't want to pile it up while bigger jobs run.
	if _, err := s.Cron.AddFunc("0/30 * * * *", func() {
		if s.Shd.SyncQueue == nil {
			return
		}
		busy, err := s.Shd.SyncQueue.IsBusy(context.Background())
		if err != nil {
			logs.Errorf("⏰ Tier 1: IsBusy check failed: %v", err)
			return
		}
		if busy {
			logs.Info("⏰ Tier 1: skipped (queue busy with higher-priority job)")
			return
		}
		now := time.Now()
		yearStr := now.Format("2006")
		monCode := now.Format("01")
		monthMap := map[string]string{
			"01": "JAN", "02": "FEB", "03": "MAR", "04": "APR", "05": "MAY", "06": "JUN",
			"07": "JUL", "08": "AUG", "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
		}
		s.enqueueOrLog(models.SyncJobTier1Fast, yearStr, []string{monthMap[monCode]}, "CRON")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 1 Cron Job: %v", err))
	}

	// 2. Job: Tier 2 - Full Maintenance Sync (Daily @ 02:00 AM) — enqueue full year
	if _, err := s.Cron.AddFunc("0 2 * * *", func() {
		yearStr := fmt.Sprintf("%d", time.Now().Year())
		s.enqueueOrLog(models.SyncJobTier2Full, yearStr, []string{}, "CRON")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 2 Cron Job: %v", err))
	}

	// 3. Retry Job — every 30 mins: scan for FAILED runs and re-enqueue them
	if s.Shd.SyncTrackingRepo != nil {
		if _, err := s.Cron.AddFunc("*/30 * * * *", func() {
			failed, err := s.Shd.SyncTrackingRepo.GetFailedRunsForRetry(context.Background(), 24*time.Hour, 3)
			if err != nil || len(failed) == 0 {
				return
			}
			logs.Infof("🔁 Retry Job: found %d failed run(s) eligible for retry", len(failed))
			for _, run := range failed {
				_ = s.Shd.SyncTrackingRepo.IncrementRetry(context.Background(), run.ID)
				triggeredBy := "RETRY:" + run.ID.String()[:8]
				switch run.JobType {
				case models.SyncJobTier1Fast:
					if run.Year != "" && run.Month != "" {
						s.enqueueOrLog(run.JobType, run.Year, []string{run.Month}, triggeredBy)
					}
				case models.SyncJobTier2Full, models.SyncJobActualFact, models.SyncJobManual:
					if run.Year != "" {
						s.enqueueOrLog(run.JobType, run.Year, []string{}, triggeredBy)
					}
				case models.SyncJobDW:
					if run.Year != "" && run.Month != "" {
						s.enqueueOrLog(run.JobType, run.Year, []string{run.Month}, triggeredBy)
					}
				}
			}
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register Retry Cron Job: %v", err))
		}

		// 4. Cleanup Job — delete TIER1_FAST sync_runs older than 24h (daily @ 03:00)
		if _, err := s.Cron.AddFunc("0 3 * * *", func() {
			deleted, err := s.Shd.SyncTrackingRepo.DeleteOldRunsByJobType(
				context.Background(), models.SyncJobTier1Fast, 24*time.Hour,
			)
			if err != nil {
				logs.Errorf("🧹 Cleanup Job: Failed to delete old TIER1_FAST runs: %v", err)
				return
			}
			logs.Infof("🧹 Cleanup Job: Deleted %d old TIER1_FAST sync_runs (>24h)", deleted)
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register Cleanup Cron Job: %v", err))
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

	// 4. Job: DW Auto-Sync (Daily @ Midnight) — fan out into per-month jobs.
	// Each month is its own queue entry so a TCP reset on one month does not
	// cascade; the retry job will re-enqueue only the failed (year, month).
	if s.Shd.ExternalSyncService != nil {
		if _, err := s.Cron.AddFunc("0 0 * * *", func() {
			s.enqueueDWPerMonth("CRON")
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register DW Sync Cron Job: %v", err))
		}

		// 🚀 IMMEDIATE STARTUP SYNC — fan out the same per-month DW jobs as the
		// midnight cron. Each per-month job re-projects ActualFact for that month
		// inside SyncFromDW, so a trailing full-year ACTUAL_FACT job would be
		// redundant; we skip it. DedupKey on the queue prevents collisions if
		// midnight CRON has already enqueued the same months.
		go func() {
			s.enqueueDWPerMonth("STARTUP")
		}()
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
