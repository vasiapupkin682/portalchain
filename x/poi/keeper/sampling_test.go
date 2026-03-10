package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	poitypes "portalchain/x/poi/types"
)

// testValidator returns a deterministic bech32 AccAddress string.
func testValidator(b byte) string {
	addr := make([]byte, 20)
	for i := range addr {
		addr[i] = b
	}
	return sdk.AccAddress(addr).String()
}

// seedReport stores an EpochReport and returns it.
func seedReport(t *testing.T, k interface {
	SetEpochReport(ctx sdk.Context, report poitypes.EpochReport)
}, ctx sdk.Context, epoch int64, validator string) poitypes.EpochReport {
	t.Helper()
	report := poitypes.EpochReport{
		Epoch:            epoch,
		Validator:        validator,
		TasksProcessed:   10,
		WeightedTaskSum:  100,
		AvgLatency:       50,
		Reliability:      sdk.NewDecWithPrec(9, 1), // 0.9
		SamplingFailures: 0,
		Timestamp:        1000,
	}
	k.SetEpochReport(ctx, report)
	return report
}

// ---------------------------------------------------------------------------
// TestSamplingRecord_SetGet
// ---------------------------------------------------------------------------

func TestSamplingRecord_SetGet(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	val := testValidator(0x01)

	record := poitypes.SamplingRecord{
		Epoch:     1,
		Validator: val,
		Status:    poitypes.SamplingStatusPending,
		Deadline:  150,
	}
	k.SetSamplingRecord(ctx, record)

	got, found := k.GetSamplingRecord(ctx, 1, val)
	require.True(t, found)
	require.Equal(t, record.Epoch, got.Epoch)
	require.Equal(t, record.Validator, got.Validator)
	require.Equal(t, record.Status, got.Status)
	require.Equal(t, record.Deadline, got.Deadline)
	require.Empty(t, got.VerifiedBy)

	// Non-existent records.
	_, found = k.GetSamplingRecord(ctx, 999, val)
	require.False(t, found)

	_, found = k.GetSamplingRecord(ctx, 1, testValidator(0xFF))
	require.False(t, found)
}

// ---------------------------------------------------------------------------
// TestGetPendingSamplingRecords
// ---------------------------------------------------------------------------

func TestGetPendingSamplingRecords(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)

	val1 := testValidator(0x01)
	val2 := testValidator(0x02)
	val3 := testValidator(0x03)

	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 1, Validator: val1, Status: poitypes.SamplingStatusPending, Deadline: 150,
	})
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 2, Validator: val2, Status: poitypes.SamplingStatusPending, Deadline: 200,
	})
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 3, Validator: val3, Status: poitypes.SamplingStatusVerified, Deadline: 100,
	})

	pending := k.GetPendingSamplingRecords(ctx)
	require.Len(t, pending, 2)

	validators := map[string]bool{}
	for _, r := range pending {
		validators[r.Validator] = true
		require.Equal(t, poitypes.SamplingStatusPending, r.Status)
	}
	require.True(t, validators[val1])
	require.True(t, validators[val2])
}

// ---------------------------------------------------------------------------
// TestExpirePendingSamplings
// ---------------------------------------------------------------------------

func TestExpirePendingSamplings(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = ctx.WithBlockHeight(50)
	val := testValidator(0x01)

	seedReport(t, k, ctx, 1, val)

	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 1, Validator: val, Status: poitypes.SamplingStatusPending, Deadline: 100,
	})

	// Reputation before expiry should be zero (never updated).
	rep, found := k.GetReputation(ctx, val)
	require.False(t, found)
	require.True(t, rep.Value.IsZero())

	// Advance past the deadline.
	ctx = ctx.WithBlockHeight(101)
	k.ExpirePendingSamplings(ctx)

	// Record status must be "failed".
	record, found := k.GetSamplingRecord(ctx, 1, val)
	require.True(t, found)
	require.Equal(t, poitypes.SamplingStatusFailed, record.Status)

	// Report's SamplingFailures must be incremented.
	report, found := k.GetEpochReport(ctx, 1, val)
	require.True(t, found)
	require.Equal(t, int64(1), report.SamplingFailures)

	// Reputation must now be set (UpdateReputation was called).
	rep, found = k.GetReputation(ctx, val)
	require.True(t, found)
	require.False(t, rep.Value.IsZero(), "reputation should be non-zero after expiry penalty")
}

func TestExpirePendingSamplings_NotExpiredYet(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = ctx.WithBlockHeight(50)
	val := testValidator(0x02)

	seedReport(t, k, ctx, 1, val)
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 1, Validator: val, Status: poitypes.SamplingStatusPending, Deadline: 100,
	})

	// Height 50, deadline 100 — not expired.
	k.ExpirePendingSamplings(ctx)

	record, found := k.GetSamplingRecord(ctx, 1, val)
	require.True(t, found)
	require.Equal(t, poitypes.SamplingStatusPending, record.Status, "should still be pending")
}
