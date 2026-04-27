package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"p2p-back-end/modules/entities/models"
)

// buildJobFromRequest is the core of the manual-trigger endpoint: it canonicalises
// inbound (job_type, year, months) and rejects malformed requests. We test it
// directly because routing through Fiber + auth middleware would obscure the
// per-job-type defaulting rules we care about.

func TestBuildJobFromRequest_ActualFactRequiresYear(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{JobType: models.SyncJobActualFact}, "ADMIN")
	assert.Nil(t, job, "ACTUAL_FACT without year must be rejected")
}

func TestBuildJobFromRequest_ManualRequiresYear(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{JobType: models.SyncJobManual}, "ADMIN")
	assert.Nil(t, job, "MANUAL without year must be rejected")
}

func TestBuildJobFromRequest_EmptyJobTypeDefaultsToActualFact(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{Year: "2026"}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, models.SyncJobActualFact, job.JobType)
		assert.Equal(t, "2026", job.Year)
	}
}

func TestBuildJobFromRequest_ActualFactPassesMonthsThrough(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{
		JobType: models.SyncJobActualFact,
		Year:    "2026",
		Months:  []string{"APR", "MAY"},
	}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, "2026", job.Year)
		assert.Equal(t, []string{"APR", "MAY"}, job.Months)
	}
}

func TestBuildJobFromRequest_Tier1DefaultsToCurrentYearAndMonth(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{JobType: models.SyncJobTier1Fast}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, fmt.Sprintf("%d", time.Now().Year()), job.Year)
		require := assert.New(t)
		require.Len(job.Months, 1)
		// Should match the current month abbreviation
		abbr := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
		require.Equal(abbr[time.Now().Month()-1], job.Months[0])
	}
}

func TestBuildJobFromRequest_Tier1RespectsExplicitMonth(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{
		JobType: models.SyncJobTier1Fast,
		Year:    "2026",
		Months:  []string{"FEB"},
	}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, "2026", job.Year)
		assert.Equal(t, []string{"FEB"}, job.Months)
	}
}

func TestBuildJobFromRequest_Tier2AlwaysFullYear(t *testing.T) {
	// Even if user passes months, TIER2_FULL must clear them — it's full-year by definition.
	job := buildJobFromRequest(&triggerRequest{
		JobType: models.SyncJobTier2Full,
		Year:    "2026",
		Months:  []string{"APR", "MAY"},
	}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, "2026", job.Year)
		assert.Empty(t, job.Months, "TIER2_FULL must always sync the full year")
	}
}

func TestBuildJobFromRequest_Tier2DefaultsToCurrentYear(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{JobType: models.SyncJobTier2Full}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, fmt.Sprintf("%d", time.Now().Year()), job.Year)
		assert.Empty(t, job.Months)
	}
}

func TestBuildJobFromRequest_DWAcceptsAnyInput(t *testing.T) {
	// DW_SYNC ignores year/months but must still produce a valid job.
	job := buildJobFromRequest(&triggerRequest{JobType: models.SyncJobDW}, "ADMIN")
	if assert.NotNil(t, job) {
		assert.Equal(t, models.SyncJobDW, job.JobType)
	}
}

func TestBuildJobFromRequest_UnknownJobTypeRejected(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{JobType: "NOT_A_JOB", Year: "2026"}, "ADMIN")
	assert.Nil(t, job, "unknown job_type must be rejected")
}

func TestBuildJobFromRequest_TriggeredByCarriedThrough(t *testing.T) {
	job := buildJobFromRequest(&triggerRequest{Year: "2026"}, "ADMIN:abc-123")
	if assert.NotNil(t, job) {
		assert.Equal(t, "ADMIN:abc-123", job.TriggeredBy)
	}
}
