package servers

import (
	"context"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"time"
)

func (s *server) StartCronJob() {
	logs.Info("⏰ Initializing Cron Jobs...")
 
	// 0. Job: Startup Sync (Run immediately on startup)
	go func() {
		s.SyncMutex.Lock()
		defer s.SyncMutex.Unlock()
 
		logs.Info("⏰ Job: Initial Startup Sync Started")
		ctx := context.Background()
		// 1. Full Maintenance Sync from 2025 to Present
		startYear := 2025
		currentYear := time.Now().Year()
 
		for year := startYear; year <= currentYear; year++ {
			yStr := fmt.Sprintf("%d", year)
			logs.Infof("⏰ Job: Startup Sync - Syncing Full Year %s...", yStr)
			if err := s.Shd.ActualService.SyncActuals(ctx, yStr, []string{}); err != nil {
				logs.Errorf("Job: Startup Sync Failed for year %s: %v", yStr, err)
			}
		}
 
		// Refresh Inventory Metadata
		if err := s.Shd.ActualService.RefreshDataInventory(ctx); err != nil {
			logs.Errorf("Job: Startup Sync Metadata Refresh Failed: %v", err)
		}
		logs.Info("⏰ Job: Initial Startup Sync Completed")
	}()

	// 1. Job: Tier 1 - Fast Sync (Every 5 mins)
	// Sync only the current month of the current year for real-time reactivity.
	if _, err := s.Cron.AddFunc("0/5 * * * *", func() {
		now := time.Now()
		yearStr := now.Format("2006")
		monCode := now.Format("01")

		// Month map for name conversion
		monthMap := map[string]string{
			"01": "JAN", "02": "FEB", "03": "MAR", "04": "APR", "05": "MAY", "06": "JUN",
			"07": "JUL", "08": "AUG", "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
		}
		mName := monthMap[monCode]

		s.SyncMutex.Lock()
		defer s.SyncMutex.Unlock()
 
		logs.Infof("⏰ Tier 1 Job: Fast-Sync Current Month (%s %s) Started", mName, yearStr)
		if err := s.Shd.ActualService.SyncActuals(context.Background(), yearStr, []string{mName}); err != nil {
			logs.Errorf("Tier 1 Job: Fast-Sync Failed: %v", err)
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

		startYear := 2025
		currentYear := time.Now().Year()

		for year := startYear; year <= currentYear; year++ {
			yearStr := fmt.Sprintf("%d", year)
			logs.Infof("⏰ Tier 2 Job: Syncing Full Year %s...", yearStr)

			// SyncActuals handles months internally batch-by-batch if passed empty months
			if err := s.Shd.ActualService.SyncActuals(context.Background(), yearStr, []string{}); err != nil {
				logs.Errorf("Tier 2 Job: Failed to sync year %s: %v", yearStr, err)
			}
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
			}

			// Finalize: Refresh Data Inventory Metadata for Admin UI
			if err := s.Shd.ActualService.RefreshDataInventory(context.Background()); err != nil {
				logs.Errorf("Job: DW Auto-Sync Failed to refresh inventory: %v", err)
			}
			logs.Info("⏰ Job: DW Auto-Sync Completed")
		}); err != nil {
			logs.Fatal(fmt.Sprintf("Failed to register DW Sync Cron Job: %v", err))
		}
	}

	// Start the Cron scheduler
	s.Cron.Start()
	logs.Info("⏰ Cron Scheduler Started")
}
