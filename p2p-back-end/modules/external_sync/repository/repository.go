package repository

import (
	"context"
	"fmt"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// batchFetchTimeout caps a single LIMIT-bounded SELECT against the DW so a
// half-open TCP connection cannot hang the worker forever (the per-month and
// per-statement timeouts are higher-level safety nets — this is the tightest
// one, since each batch fetches at most batchSize rows and should finish in
// seconds under healthy network/DB conditions).
const batchFetchTimeout = 90 * time.Second

// batchHeartbeatEvery emits progress logs from inside the streaming loop so the
// operator can distinguish "running slowly" from "hung" without inspecting DB
// state. Tuned to be frequent enough for monitoring but rare enough not to
// spam the log file.
const batchHeartbeatEvery = 30 * time.Second

// monthRange returns [start, end) covering the given (year, month) — used to
// avoid EXTRACT(YEAR/MONTH FROM CAST(...)) which prevents index usage on
// Posting_Date. Date-range comparison is sargable and uses the index.
func monthRange(year, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

type externalSyncRepository struct {
	localDb *gorm.DB
	dwDb    *gorm.DB
}

func NewExternalSyncRepository(localDb *gorm.DB, dwDb *gorm.DB) models.ExternalSyncRepository {
	// Robustness: Ensure local tables have correct schema/PK.
	// We previously dropped these once to clear legacy ghost columns.
	if err := localDb.AutoMigrate(&models.AchHmwGleEntity{}, &models.ClikGleEntity{}); err != nil {
		fmt.Printf("[WARN] AutoMigrate failed: %v\n", err)
	}

	return &externalSyncRepository{
		localDb: localDb,
		dwDb:    dwDb,
	}
}

// PingDW does a low-level liveness check on the DW connection. Used by
// SyncFromDW as a fail-fast gate so we don't enqueue 17 monthly jobs that all
// hang on a half-open TCP connection — if the DW is unreachable, abort the
// whole sync up-front and let the retry job pick it up later.
func (r *externalSyncRepository) PingDW(ctx context.Context) error {
	sqlDB, err := r.dwDb.DB()
	if err != nil {
		return fmt.Errorf("extSyncRepo.PingDW: get sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("extSyncRepo.PingDW: %w", err)
	}
	return nil
}

// FetchHMWInBatches streams DW raw rows for one month using cursor pagination on `id`.
// We avoid FindInBatches because, when used with `.Table(...)` + a custom SELECT alias,
// GORM cannot reliably detect the destination's primary key and errors with
// "primary key required". Manual cursor pagination is portable and fast.
func (r *externalSyncRepository) FetchHMWInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.AchHmwGleEntity) error) error {
	if batchSize <= 0 {
		batchSize = 5000
	}
	start, end := monthRange(year, month)

	lastID := 0
	totalFetched := 0
	lastHeartbeat := time.Now()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var batch []models.AchHmwGleEntity

		// Per-batch timeout. Each LIMIT-bounded query should finish in seconds; if it
		// blocks past batchFetchTimeout the connection is half-open or the DW is sick,
		// so we fail fast rather than letting the worker hang for the per-month cap.
		batchCtx, cancel := context.WithTimeout(ctx, batchFetchTimeout)
		err := r.dwDb.WithContext(batchCtx).
			Table("achhmw_gle_api").
			Select(`
				id, "Entry_No", "System_Created_Entry", CAST("Posting_Date" AS DATE) AS "Posting_Date",
				"Document_Type", "Document_No", "G_L_Account_No", "G_L_Account_Name", "Description",
				"Description_2", "Job_No", "Global_Dimension_1_Code", "Global_Dimension_2_Code",
				"Vendor_Bank_Account", "IC_Partner_Code", "Gen_Posting_Type", "Gen_Bus_Posting_Group",
				"Gen_Prod_Posting_Group", "Quantity", "Amount", "Debit_Amount", "Credit_Amount",
				"Additional_Currency_Amount", "VAT_Amount", "Bal_Account_Type", "Bal_Account_No",
				"User_ID", "Source_Code", "Reason_Code", "Reversed", "Reversed_by_Entry_No",
				"Car_ID", "Reversed_Entry_No", "FA_Entry_Type", "FA_Entry_No", "Dimension_Set_ID",
				"Cutomer_No", "Cutomer_Name", "Vendor_Name", "Serial_No", "Serial_Description",
				company, branch, _id, __v
			`).
			Where(`"Posting_Date" >= ? AND "Posting_Date" < ?`, start, end).
			Where(`id > ?`, lastID).
			Order("id ASC").
			Limit(batchSize).
			Scan(&batch).Error
		cancel()
		if err != nil {
			return fmt.Errorf("extSyncRepo.FetchHMWInBatches: %w", err)
		}
		if len(batch) == 0 {
			return nil
		}
		if err := handle(batch); err != nil {
			return err
		}
		totalFetched += len(batch)
		if time.Since(lastHeartbeat) >= batchHeartbeatEvery {
			logs.Infof("[DW Sync][Heartbeat] HMW %d-%02d: %d rows fetched (cursor=%d)",
				year, month, totalFetched, batch[len(batch)-1].ID)
			lastHeartbeat = time.Now()
		}
		lastID = batch[len(batch)-1].ID
		if len(batch) < batchSize {
			return nil
		}
	}
}

// FetchCLIKInBatches — same cursor pattern as FetchHMWInBatches.
func (r *externalSyncRepository) FetchCLIKInBatches(ctx context.Context, year int, month int, batchSize int, handle func([]models.ClikGleEntity) error) error {
	if batchSize <= 0 {
		batchSize = 5000
	}
	start, end := monthRange(year, month)

	lastID := 0
	totalFetched := 0
	lastHeartbeat := time.Now()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var batch []models.ClikGleEntity

		batchCtx, cancel := context.WithTimeout(ctx, batchFetchTimeout)
		err := r.dwDb.WithContext(batchCtx).
			Table("general_ledger_entries_clik").
			Select(`
				id, "Entry_No", "System_Created_Entry", CAST("Posting_Date" AS DATE) AS "Posting_Date",
				"Document_Type", "Document_No", "G_L_Account_No", "G_L_Account_Name", "Description",
				"Description_2", "Job_No", "Global_Dimension_1_Code", "Global_Dimension_2_Code",
				"Vendor_Bank_Account", "IC_Partner_Code", "Gen_Posting_Type", "Gen_Bus_Posting_Group",
				"Gen_Prod_Posting_Group", "Quantity", "Amount", "Debit_Amount", "Credit_Amount",
				"Additional_Currency_Amount", "VAT_Amount", "Bal_Account_Type", "Bal_Account_No",
				"User_ID", "Source_Code", "Reason_Code", "Reversed", "Reversed_by_Entry_No",
				"Car_ID", "Reversed_Entry_No", "FA_Entry_Type", "FA_Entry_No", "Dimension_Set_ID",
				"Cutomer_No", "Cutomer_Name", "Vendor_Name", "Serial_No", "Serial_Description",
				'CLIK' as company
			`).
			Where(`"Posting_Date" >= ? AND "Posting_Date" < ?`, start, end).
			Where(`id > ?`, lastID).
			Order("id ASC").
			Limit(batchSize).
			Scan(&batch).Error
		cancel()
		if err != nil {
			return fmt.Errorf("extSyncRepo.FetchCLIKInBatches: %w", err)
		}
		if len(batch) == 0 {
			return nil
		}
		if err := handle(batch); err != nil {
			return err
		}
		totalFetched += len(batch)
		if time.Since(lastHeartbeat) >= batchHeartbeatEvery {
			logs.Infof("[DW Sync][Heartbeat] CLIK %d-%02d: %d rows fetched (cursor=%d)",
				year, month, totalFetched, batch[len(batch)-1].ID)
			lastHeartbeat = time.Now()
		}
		lastID = batch[len(batch)-1].ID
		if len(batch) < batchSize {
			return nil
		}
	}
}

// DeleteHMWByYearMonth ลบข้อมูล HMW raw ของเดือน/ปีที่ระบุก่อน insert ใหม่
// เพื่อป้องกัน duplicate rows (idempotent sync). ใช้ date-range เพื่อให้ใช้ index ได้
func (r *externalSyncRepository) DeleteHMWByYearMonth(ctx context.Context, year int, month int) error {
	start, end := monthRange(year, month)
	err := r.localDb.WithContext(ctx).
		Exec(`DELETE FROM achhmw_gle_api WHERE "Posting_Date" >= ? AND "Posting_Date" < ?`, start, end).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.DeleteHMWByYearMonth: %w", err)
	}
	return nil
}

// DeleteCLIKByYearMonth ลบข้อมูล CLIK raw ของเดือน/ปีที่ระบุก่อน insert ใหม่
func (r *externalSyncRepository) DeleteCLIKByYearMonth(ctx context.Context, year int, month int) error {
	start, end := monthRange(year, month)
	err := r.localDb.WithContext(ctx).
		Exec(`DELETE FROM general_ledger_entries_clik WHERE "Posting_Date" >= ? AND "Posting_Date" < ?`, start, end).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.DeleteCLIKByYearMonth: %w", err)
	}
	return nil
}

// UpsertHMWLocal inserts HMW rows. Caller must DELETE the target year-month first
// (via DeleteHMWByYearMonth) to guarantee idempotency / no duplicates.
// Uses DoNothing on conflict as last-resort safety net (PK id collision should not happen
// with auto-increment, but if it does we don't overwrite data silently).
func (r *externalSyncRepository) UpsertHMWLocal(ctx context.Context, data []models.AchHmwGleEntity) error {
	if len(data) == 0 {
		return nil
	}
	// Reset PK so DB generates new ids; prevents accidental PK reuse if DW returns duplicate id fields.
	for i := range data {
		data[i].ID = 0
	}
	err := r.localDb.WithContext(ctx).
		Table("achhmw_gle_api").
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(data, 2000).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.UpsertHMWLocal: %w", err)
	}
	return nil
}

// UpsertCLIKLocal inserts CLIK rows. Same semantics as UpsertHMWLocal.
func (r *externalSyncRepository) UpsertCLIKLocal(ctx context.Context, data []models.ClikGleEntity) error {
	if len(data) == 0 {
		return nil
	}
	for i := range data {
		data[i].ID = 0
		data[i].Company = "CLIK"
	}
	err := r.localDb.WithContext(ctx).
		Table("general_ledger_entries_clik").
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(data, 2000).Error
	if err != nil {
		return fmt.Errorf("extSyncRepo.UpsertCLIKLocal: %w", err)
	}
	return nil
}
