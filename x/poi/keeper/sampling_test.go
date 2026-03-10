package keeper_test

import (
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	poikeeper "portalchain/x/poi/keeper"
	poitypes "portalchain/x/poi/types"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockStakingKeeper struct{}

var _ poitypes.StakingKeeper = (*mockStakingKeeper)(nil)

func (*mockStakingKeeper) GetValidator(_ sdk.Context, _ sdk.ValAddress) (stakingtypes.Validator, bool) {
	return stakingtypes.Validator{}, false
}
func (*mockStakingKeeper) GetAllValidators(_ sdk.Context) []stakingtypes.Validator { return nil }

type mockBankKeeper struct{}

var _ poitypes.BankKeeper = (*mockBankKeeper)(nil)

func (*mockBankKeeper) SpendableCoins(_ sdk.Context, _ sdk.AccAddress) sdk.Coins { return nil }

// ---------------------------------------------------------------------------
// Test helper
// ---------------------------------------------------------------------------

func setupPoiKeeper(t *testing.T, height int64, appHash []byte) (*poikeeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := sdk.NewKVStoreKey(poitypes.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := poikeeper.NewKeeper(cdc, storeKey, &mockStakingKeeper{}, &mockBankKeeper{})

	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Height:  height,
		AppHash: appHash,
	}, false, log.NewNopLogger())

	return k, ctx
}

// testValidator returns a deterministic bech32 AccAddress string.
func testValidator(b byte) string {
	addr := make([]byte, 20)
	for i := range addr {
		addr[i] = b
	}
	return sdk.AccAddress(addr).String()
}

// seedReport stores an EpochReport and returns it.
func seedReport(t *testing.T, k *poikeeper.Keeper, ctx sdk.Context, epoch int64, validator string) poitypes.EpochReport {
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
	k, ctx := setupPoiKeeper(t, 50, nil)
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

	// Non-existent record.
	_, found = k.GetSamplingRecord(ctx, 999, val)
	require.False(t, found)

	_, found = k.GetSamplingRecord(ctx, 1, testValidator(0xFF))
	require.False(t, found)
}

// ---------------------------------------------------------------------------
// TestGetPendingSamplingRecords
// ---------------------------------------------------------------------------

func TestGetPendingSamplingRecords(t *testing.T) {
	k, ctx := setupPoiKeeper(t, 50, nil)

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
	k, ctx := setupPoiKeeper(t, 50, nil)
	val := testValidator(0x01)

	// Seed an epoch report.
	seedReport(t, k, ctx, 1, val)

	// Pending record with Deadline=100.
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 1, Validator: val, Status: poitypes.SamplingStatusPending, Deadline: 100,
	})

	// Reputation before expiry should be zero (never updated).
	rep, found := k.GetReputation(ctx, val)
	require.False(t, found)
	require.True(t, rep.Value.IsZero())

	// Advance to height 101 — past the deadline.
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
	k, ctx := setupPoiKeeper(t, 50, nil)
	val := testValidator(0x02)

	seedReport(t, k, ctx, 1, val)
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: 1, Validator: val, Status: poitypes.SamplingStatusPending, Deadline: 100,
	})

	// Height is 50, deadline is 100 — not expired.
	k.ExpirePendingSamplings(ctx)

	record, found := k.GetSamplingRecord(ctx, 1, val)
	require.True(t, found)
	require.Equal(t, poitypes.SamplingStatusPending, record.Status, "should still be pending")
}
