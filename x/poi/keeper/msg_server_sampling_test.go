package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	poikeeper "portalchain/x/poi/keeper"
	poitypes "portalchain/x/poi/types"
)

// AppHash values that deterministically control ShouldSample for epoch=1.
//
//	seed = []byte("1") = [0x31] → byte sum = 49
//	AppHash{1}: total sum = 1+49 = 50, 50%10 = 0 → sampled
//	AppHash{0}: total sum = 0+49 = 49, 49%10 = 9 → NOT sampled
var (
	appHashSampled    = []byte{1}
	appHashNotSampled = []byte{0}
	testEpoch         = int64(1)
)

func submitTestReport(t *testing.T, srv poitypes.FullMsgServer, ctx sdk.Context, validator string) {
	t.Helper()
	msg := &poitypes.MsgSubmitEpochReport{
		Epoch:            testEpoch,
		Validator:        validator,
		TasksProcessed:   10,
		WeightedTaskSum:  100,
		AvgLatency:       50,
		Reliability:      sdk.NewDecWithPrec(9, 1), // 0.9
		SamplingFailures: 0,
		Timestamp:        1000,
	}
	_, err := srv.SubmitEpochReport(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
}

// withHeader returns a context with the given block height and AppHash,
// preserving other header fields (ChainID, Time) from the original context.
func withHeader(ctx sdk.Context, height int64, appHash []byte) sdk.Context {
	h := ctx.BlockHeader()
	return ctx.WithBlockHeader(tmproto.Header{
		Height:  height,
		AppHash: appHash,
		ChainID: h.ChainID,
		Time:    h.Time,
	})
}

// ---------------------------------------------------------------------------
// TestSubmitEpochReport_SamplingDeferred
// ---------------------------------------------------------------------------

func TestSubmitEpochReport_SamplingDeferred(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = withHeader(ctx, 50, appHashSampled)

	srv := poikeeper.NewMsgServerImpl(*k)
	val := testValidator(0x10)

	submitTestReport(t, srv, ctx, val)

	// Sampling record must exist with pending status.
	record, found := k.GetSamplingRecord(ctx, testEpoch, val)
	require.True(t, found, "sampling record should be created")
	require.Equal(t, poitypes.SamplingStatusPending, record.Status)
	require.Equal(t, int64(50+poitypes.SamplingVerificationWindow), record.Deadline)

	// Reputation must NOT have been updated yet (deferred).
	rep, found := k.GetReputation(ctx, val)
	require.False(t, found, "reputation should not exist yet for sampled report")
	require.True(t, rep.Value.IsZero())
}

// ---------------------------------------------------------------------------
// TestSubmitEpochReport_NoSampling
// ---------------------------------------------------------------------------

func TestSubmitEpochReport_NoSampling(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = withHeader(ctx, 50, appHashNotSampled)

	srv := poikeeper.NewMsgServerImpl(*k)
	val := testValidator(0x11)

	submitTestReport(t, srv, ctx, val)

	// No sampling record should exist.
	_, found := k.GetSamplingRecord(ctx, testEpoch, val)
	require.False(t, found, "no sampling record for non-sampled report")

	// Reputation must have been updated immediately.
	rep, found := k.GetReputation(ctx, val)
	require.True(t, found, "reputation should exist after non-sampled report")
	require.False(t, rep.Value.IsZero(), "reputation should be non-zero")
}

// ---------------------------------------------------------------------------
// TestVerifySampling_Passed
// ---------------------------------------------------------------------------

func TestVerifySampling_Passed(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = ctx.WithBlockHeight(50)

	srv := poikeeper.NewMsgServerImpl(*k)
	val := testValidator(0x20)
	verifier := testValidator(0x21)

	seedReport(t, k, ctx, testEpoch, val)
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: testEpoch, Validator: val,
		Status: poitypes.SamplingStatusPending, Deadline: 200,
	})

	msg := &poitypes.MsgVerifySampling{
		Verifier:  verifier,
		Epoch:     testEpoch,
		Validator: val,
		Passed:    true,
	}
	_, err := srv.VerifySampling(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	record, found := k.GetSamplingRecord(ctx, testEpoch, val)
	require.True(t, found)
	require.Equal(t, poitypes.SamplingStatusVerified, record.Status)
	require.Equal(t, verifier, record.VerifiedBy)

	// Reputation should have been updated (the deferred update now applied).
	rep, found := k.GetReputation(ctx, val)
	require.True(t, found)
	require.False(t, rep.Value.IsZero())

	// SamplingFailures must remain 0 (passed verification).
	report, _ := k.GetEpochReport(ctx, testEpoch, val)
	require.Equal(t, int64(0), report.SamplingFailures)
}

// ---------------------------------------------------------------------------
// TestVerifySampling_Failed
// ---------------------------------------------------------------------------

func TestVerifySampling_Failed(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = ctx.WithBlockHeight(50)

	srv := poikeeper.NewMsgServerImpl(*k)
	val := testValidator(0x30)
	verifier := testValidator(0x31)

	seedReport(t, k, ctx, testEpoch, val)
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: testEpoch, Validator: val,
		Status: poitypes.SamplingStatusPending, Deadline: 200,
	})

	msg := &poitypes.MsgVerifySampling{
		Verifier:  verifier,
		Epoch:     testEpoch,
		Validator: val,
		Passed:    false,
	}
	_, err := srv.VerifySampling(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	record, found := k.GetSamplingRecord(ctx, testEpoch, val)
	require.True(t, found)
	require.Equal(t, poitypes.SamplingStatusFailed, record.Status)
	require.Equal(t, verifier, record.VerifiedBy)

	// SamplingFailures must be incremented to 1.
	report, found := k.GetEpochReport(ctx, testEpoch, val)
	require.True(t, found)
	require.Equal(t, int64(1), report.SamplingFailures)

	// Reputation was updated with penalty.
	rep, found := k.GetReputation(ctx, val)
	require.True(t, found)
	require.False(t, rep.Value.IsZero())
}

// ---------------------------------------------------------------------------
// TestVerifySampling_Errors
// ---------------------------------------------------------------------------

func TestVerifySampling_Errors(t *testing.T) {
	k, ctx, _ := setupPoiKeeper(t)
	ctx = ctx.WithBlockHeight(50)

	srv := poikeeper.NewMsgServerImpl(*k)
	val := testValidator(0x40)
	verifier := testValidator(0x41)

	t.Run("non-existent record → ErrSamplingNotFound", func(t *testing.T) {
		msg := &poitypes.MsgVerifySampling{
			Verifier: verifier, Epoch: 999, Validator: val, Passed: true,
		}
		_, err := srv.VerifySampling(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrSamplingNotFound)
	})

	// Seed report and pending record for remaining tests.
	seedReport(t, k, ctx, testEpoch, val)
	k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
		Epoch: testEpoch, Validator: val,
		Status: poitypes.SamplingStatusPending, Deadline: 60,
	})

	t.Run("expired record → ErrSamplingExpired", func(t *testing.T) {
		expiredCtx := ctx.WithBlockHeight(61)
		msg := &poitypes.MsgVerifySampling{
			Verifier: verifier, Epoch: testEpoch, Validator: val, Passed: true,
		}
		_, err := srv.VerifySampling(sdk.WrapSDKContext(expiredCtx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrSamplingExpired)
	})

	t.Run("self-verification → ErrSelfVerification", func(t *testing.T) {
		msg := &poitypes.MsgVerifySampling{
			Verifier: val, Epoch: testEpoch, Validator: val, Passed: true,
		}
		_, err := srv.VerifySampling(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrSelfVerification)
	})

	t.Run("already resolved → ErrSamplingAlreadyResolved", func(t *testing.T) {
		k.SetSamplingRecord(ctx, poitypes.SamplingRecord{
			Epoch: testEpoch, Validator: val,
			Status: poitypes.SamplingStatusVerified, Deadline: 200,
			VerifiedBy: verifier,
		})
		msg := &poitypes.MsgVerifySampling{
			Verifier: verifier, Epoch: testEpoch, Validator: val, Passed: true,
		}
		_, err := srv.VerifySampling(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrSamplingAlreadyResolved)
	})
}
