package servers

import (
	"fmt"
	"p2p-back-end/logs"
	"time"
)

func (s *server) StartCronJob() {
	logs.Info("⏰ Initializing Cron Jobs...")

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

		logs.Infof("⏰ Tier 1 Job: Fast-Sync Current Month (%s %s) Started", mName, yearStr)
		if err := s.ActualSrv.SyncActuals(yearStr, []string{mName}); err != nil {
			logs.Errorf("Tier 1 Job: Fast-Sync Failed: %v", err)
		}
		logs.Info("⏰ Tier 1 Job: Fast-Sync Completed Successfully")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Tier 1 Cron Job: %v", err))
	}

	// 2. Job: Tier 2 - Full Maintenance Sync (Daily @ 02:00 AM)
	// Synchronize all data from 2025 to Present to ensure full consistency during off-peak hours.
	if _, err := s.Cron.AddFunc("0 2 * * *", func() {
		logs.Info("⏰ Tier 2 Job: Full Maintenance Sync Started")

		startYear := 2025
		currentYear := time.Now().Year()

		for year := startYear; year <= currentYear; year++ {
			yearStr := fmt.Sprintf("%d", year)
			logs.Infof("⏰ Tier 2 Job: Syncing Full Year %s...", yearStr)

			// SyncActuals handles months internally batch-by-batch if passed empty months
			if err := s.ActualSrv.SyncActuals(yearStr, []string{}); err != nil {
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
		if err := s.DeptSrv.ManageDepartments(); err != nil {
			logs.Error("Failed to Seed Departments: " + err.Error())
		} else {
			logs.Info("System: Department Data Seeding Completed Successfully")
		}
	}()

	// 3. Initial Broadcast Requests (RabbitMQ)
	// Ask Example Service for data startup
	if s.ProducerSrv != nil {
		go func() {
			logs.Info("🚀 P2P: Sending initial data sync requests (Broadcast Begin)...")

			if err := s.ProducerSrv.RequestCompanySync(); err != nil {
				logs.Warnf("Failed to request Company Sync: %v", err)
			}
			if err := s.ProducerSrv.RequestDepartmentSync(); err != nil {
				logs.Warnf("Failed to request Department Sync: %v", err)
			}
			if err := s.ProducerSrv.RequestUserSync(); err != nil {
				logs.Warnf("Failed to request User Sync: %v", err)
			}

			// Also push our own local data if anyone is listening
			s.MasterSrv.BroadcastAllData()
		}()
	}

	// Start the Cron scheduler
	s.Cron.Start()
	logs.Info("⏰ Cron Scheduler Started")
}
