package servers

import (
	"context"
	"fmt"
	"time"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
)

func (s *server) StartCronJob() {
	logs.Info("⏰ Initializing Cron Jobs...")

	// 1. Job: Tier 1 - Fast Sync (Every 5 mins)
	// Sync only the current month of the current year for real-time reactivity.
	if _, err := s.Cron.AddFunc("0/5 * * * *", func() {
		// TryLock: ถ้ามี sync ใหญ่กำลังรันอยู่ ให้ skip รอบนี้ไป ไม่ต้องรอ queue
		if !s.SyncMutex.TryLock() {
			logs.Info("⏰ Tier 1 Job: Skipped (another sync is running)")
			return
		}
		defer s.SyncMutex.Unlock()

		now := time.Now()
		yearStr := now.Format("2006")
		monCode := now.Format("01")

		monthMap := map[string]string{
			"01": "JAN", "02": "FEB", "03": "MAR", "04": "APR", "05": "MAY", "06": "JUN",
			"07": "JUL", "08": "AUG", "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
		}
		mName := monthMap[monCode]

		logs.Infof("⏰ Tier 1 Job: Fast-Sync Current Month (%s %s) Started", mName, yearStr)
		if err := s.Shd.ActualService.SyncActuals(context.Background(), yearStr, []string{mName}); err != nil {
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
		if syncErr != nil {
			logs.Error("⏰ Tier 2 Job: Full Maintenance Sync Completed with Errors")
			return
		}
		logs.Info("⏰ Tier 2 Job: Full Maintenance Sync Completed")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 2 Cron Job: %v", err))
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