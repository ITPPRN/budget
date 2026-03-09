package servers

import (
	"fmt"
	"p2p-back-end/logs"
	"time"

	"go.uber.org/zap"
)

func (s *server) StartCronJob() {
	logs.Info("⏰ Initializing Cron Jobs...")

	// 1. Job: Sync Data (Actuals Sync) - Replacing manual ticker
	if _, err := s.Cron.AddFunc("0/5 * * * *", func() {
		logs.Info("⏰ Job: Auto-Sync Central Actuals Started (Triggered via Master/Budget Sync placeholder)")
		// Currently handled by the unified BudgetSrv.SyncActuals
		if err := s.BudgetSrv.SyncActuals(time.Now().Format("2006"), []string{}); err != nil {
			logs.Error("Auto-Sync Central Actuals Failed", zap.Error(err))
		}
		logs.Info("⏰ Job: Auto-Sync Central Actuals Completed Successfully")
	}); err != nil {
		logs.Fatal(fmt.Sprintf("Failed to register Cron Job: %v", err))
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
