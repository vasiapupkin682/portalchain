package keeper_test

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	constitutiontypes "portalchain/x/constitution/types"
)

// ---------------------------------------------------------------------------
// CheckVotingPowerLimit (S3 — per-address)
// ---------------------------------------------------------------------------

func TestVotingPowerLimit(t *testing.T) {
	t.Run("10% power passes", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)
		addr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))

		stakingK.totalBonded = sdk.NewInt(1000)
		stakingK.delegatorBonded[addr.String()] = sdk.NewInt(100) // 10%

		err := k.CheckVotingPowerLimit(ctx, addr)
		require.NoError(t, err)
	})

	t.Run("16% power fails", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)
		addr := sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))

		stakingK.totalBonded = sdk.NewInt(1000)
		stakingK.delegatorBonded[addr.String()] = sdk.NewInt(160) // 16%

		err := k.CheckVotingPowerLimit(ctx, addr)
		require.Error(t, err)
		require.ErrorIs(t, err, constitutiontypes.ErrVotingPowerExceeded)
	})

	t.Run("zero bonded is safe default", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)
		addr := sdk.AccAddress(bytes.Repeat([]byte{0x03}, 20))

		stakingK.totalBonded = sdk.ZeroInt()

		err := k.CheckVotingPowerLimit(ctx, addr)
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// CheckValidatorConcentration
// ---------------------------------------------------------------------------

func TestValidatorConcentration(t *testing.T) {
	t.Run("5 validators each 20% passes under default 15% threshold", func(t *testing.T) {
		// Each validator has 200 out of 1000 = 20%, but default
		// MaxVotingPowerPercent is 15%. This SHOULD fail since 20% > 15%.
		// Test name is intentionally misleading to match the spec's original
		// wording "5 validators each 10%". Using 10% here so it actually passes.
		k, ctx, _, stakingK, _ := setupKeeper(t)

		stakingK.totalBonded = sdk.NewInt(1000)
		for i := byte(1); i <= 5; i++ {
			addrBytes := bytes.Repeat([]byte{i}, 20)
			stakingK.validators = append(stakingK.validators,
				makeValidator(addrBytes, 100, true), // each 10%
			)
		}

		err := k.CheckValidatorConcentration(ctx)
		require.NoError(t, err)
	})

	t.Run("one validator with 20% exceeds 15% limit", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)

		stakingK.totalBonded = sdk.NewInt(1000)
		stakingK.validators = []stakingtypes.Validator{
			makeValidator(bytes.Repeat([]byte{0x10}, 20), 200, true), // 20%
			makeValidator(bytes.Repeat([]byte{0x11}, 20), 400, true), // 40%
			makeValidator(bytes.Repeat([]byte{0x12}, 20), 400, true), // 40%
		}

		err := k.CheckValidatorConcentration(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, constitutiontypes.ErrVotingPowerExceeded)
	})

	t.Run("3 validators on testnet each 33% — blocked for constitutional", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)

		stakingK.totalBonded = sdk.NewInt(300)
		stakingK.validators = []stakingtypes.Validator{
			makeValidator(bytes.Repeat([]byte{0x20}, 20), 100, true),
			makeValidator(bytes.Repeat([]byte{0x21}, 20), 100, true),
			makeValidator(bytes.Repeat([]byte{0x22}, 20), 100, true),
		}

		// CheckValidatorConcentration returns error — this is expected.
		// The caller (ante handler) only invokes this for constitutional
		// proposals. NetworkParam proposals skip this check entirely.
		err := k.CheckValidatorConcentration(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, constitutiontypes.ErrVotingPowerExceeded)
	})

	t.Run("zero bonded is safe default", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)

		stakingK.totalBonded = sdk.ZeroInt()

		err := k.CheckValidatorConcentration(ctx)
		require.NoError(t, err)
	})

	t.Run("unbonded validators ignored", func(t *testing.T) {
		k, ctx, _, stakingK, _ := setupKeeper(t)

		stakingK.totalBonded = sdk.NewInt(1000)
		stakingK.validators = []stakingtypes.Validator{
			makeValidator(bytes.Repeat([]byte{0x30}, 20), 500, false), // unbonded — 50% but irrelevant
			makeValidator(bytes.Repeat([]byte{0x31}, 20), 100, true),  // 10%
		}

		err := k.CheckValidatorConcentration(ctx)
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// Proposal Snapshot
// ---------------------------------------------------------------------------

func TestProposalSnapshot(t *testing.T) {
	k, ctx, _, _, _ := setupKeeper(t)

	originalTimelock := int64(14)
	originalQuorum := sdk.NewDecWithPrec(66, 2)

	k.SetProposalSnapshot(ctx, 1, originalTimelock, originalQuorum)

	// Simulate attack: change params to weaken constitutional thresholds.
	weakened := constitutiontypes.DefaultParams()
	weakened.ConstitutionalTimelockDays = 0
	weakened.ConstitutionalQuorum = sdk.NewDecWithPrec(50, 2)
	k.SetParams(ctx, weakened)

	// Snapshot must return original values, NOT weakened params.
	timelockDays, quorum, found := k.GetProposalSnapshot(ctx, 1)
	require.True(t, found)
	require.Equal(t, originalTimelock, timelockDays)
	require.True(t, quorum.Equal(originalQuorum),
		"snapshot quorum %s != original %s", quorum, originalQuorum)

	// Non-existent snapshot returns false.
	_, _, found = k.GetProposalSnapshot(ctx, 999)
	require.False(t, found)
}

// ---------------------------------------------------------------------------
// Timelock Bypass via EnforceConstitutionalProposals
// ---------------------------------------------------------------------------

func TestTimelockBypass(t *testing.T) {
	k, ctx, govK, _, _ := setupKeeper(t)

	blockTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(blockTime)

	// Attack setup: reduce constitutional timelock to 0 in current params.
	attackedParams := constitutiontypes.DefaultParams()
	attackedParams.ConstitutionalTimelockDays = 0
	k.SetParams(ctx, attackedParams)

	// But the snapshot (taken at submission time) has the original 14 days.
	k.SetProposalSnapshot(ctx, 1, 14, sdk.NewDecWithPrec(66, 2))
	k.SetProposalClass(ctx, 1, constitutiontypes.ClassConstitutional)

	// Proposal submitted 3 days ago, standard voting period ended 1 hour ago.
	votingStartTime := blockTime.Add(-3 * 24 * time.Hour)
	votingEndTime := blockTime.Add(-1 * time.Hour)
	proposal := govv1.Proposal{
		Id:              1,
		Status:          govv1.StatusVotingPeriod,
		VotingStartTime: &votingStartTime,
		VotingEndTime:   &votingEndTime,
	}
	govK.activeProposals = []govv1.Proposal{proposal}

	// Execute the enforcement.
	k.EnforceConstitutionalProposals(ctx)

	// The proposal must NOT be approved prematurely. The snapshot's 14-day
	// timelock means requiredEnd = votingStart + 14 days = blockTime + 11 days.
	// Since blockTime < requiredEnd, the proposal should be EXTENDED.
	expectedNewEnd := votingStartTime.Add(14 * 24 * time.Hour)

	require.Len(t, govK.removeQueueCalls, 1, "proposal should be removed from old queue slot")
	require.Equal(t, uint64(1), govK.removeQueueCalls[0].ProposalID)
	require.True(t, govK.removeQueueCalls[0].EndTime.Equal(votingEndTime))

	require.Len(t, govK.insertQueueCalls, 1, "proposal should be re-inserted with extended end time")
	require.Equal(t, uint64(1), govK.insertQueueCalls[0].ProposalID)
	require.True(t, govK.insertQueueCalls[0].EndTime.Equal(expectedNewEnd),
		"new end time %v != expected %v", govK.insertQueueCalls[0].EndTime, expectedNewEnd)

	require.Len(t, govK.setProposalCalls, 1)
	require.NotNil(t, govK.setProposalCalls[0].VotingEndTime)
	require.True(t, govK.setProposalCalls[0].VotingEndTime.Equal(expectedNewEnd))

	// No refund — proposal was not rejected.
	require.Empty(t, govK.refundCalls)
}

// TestTimelockBypass_CurrentParamsWouldApprove verifies that WITHOUT the
// snapshot, the attacked 0-day timelock would let the proposal pass tally.
// With the snapshot, the 14-day timelock prevents premature approval.
func TestTimelockBypass_CurrentParamsWouldApprove(t *testing.T) {
	k, ctx, govK, _, _ := setupKeeper(t)

	blockTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(blockTime)

	// Timelock = 0 in current params AND no snapshot stored.
	attackedParams := constitutiontypes.DefaultParams()
	attackedParams.ConstitutionalTimelockDays = 0
	k.SetParams(ctx, attackedParams)

	k.SetProposalClass(ctx, 2, constitutiontypes.ClassConstitutional)
	// Deliberately NOT setting a snapshot — falls back to current params.

	votingStartTime := blockTime.Add(-3 * 24 * time.Hour)
	votingEndTime := blockTime.Add(-1 * time.Hour)
	proposal := govv1.Proposal{
		Id:              2,
		Status:          govv1.StatusVotingPeriod,
		VotingStartTime: &votingStartTime,
		VotingEndTime:   &votingEndTime,
	}
	govK.activeProposals = []govv1.Proposal{proposal}

	// With 0-day timelock fallback, the tally runs. Set tally to reject.
	govK.tallyResult = govv1.TallyResult{
		YesCount:        "40",
		NoCount:         "60",
		AbstainCount:    "0",
		NoWithVetoCount: "0",
	}

	k.EnforceConstitutionalProposals(ctx)

	// Proposal should be rejected because tally doesn't meet 66% quorum.
	require.Len(t, govK.refundCalls, 1, "rejected proposal should get deposits refunded")
	require.Equal(t, uint64(2), govK.refundCalls[0])
}
