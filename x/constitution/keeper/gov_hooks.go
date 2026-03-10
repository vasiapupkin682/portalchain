package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/constitution/types"
)

// GovHooks implements govtypes.GovHooks to classify proposals on submission.
type GovHooks struct {
	k Keeper
}

func (k Keeper) GovHooks() GovHooks {
	return GovHooks{k: k}
}

// AfterProposalSubmission classifies the proposal, stores its class, and
// snapshots constitutional params to prevent retroactive weakening.
func (h GovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	class := h.k.ClassifyProposalByID(ctx, proposalID)
	if class == types.ClassSacredViolation {
		h.k.Logger(ctx).Error("sacred violation detected after proposal submission",
			"proposal_id", proposalID,
		)
	}
	h.k.SetProposalClass(ctx, proposalID, class)

	if class == types.ClassConstitutional {
		params := h.k.GetParams(ctx)
		h.k.SetProposalSnapshot(ctx, proposalID, params.ConstitutionalTimelockDays, params.ConstitutionalQuorum)
		h.k.Logger(ctx).Info("constitutional params snapshot stored",
			"proposal_id", proposalID,
			"timelock_days", params.ConstitutionalTimelockDays,
			"quorum", params.ConstitutionalQuorum.String(),
		)
	}
}

func (h GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
}

func (h GovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
}

func (h GovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
}

// AfterProposalVotingPeriodEnded fires AFTER gov has tallied and executed
// the proposal (confirmed in SDK v0.47.3 x/gov/abci.go:121). Constitutional
// enforcement is handled in the constitution module's EndBlocker which runs
// BEFORE gov's EndBlocker. This hook is purely diagnostic.
func (h GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	class, found := h.k.GetProposalClass(ctx, proposalID)
	if !found {
		return
	}

	h.k.Logger(ctx).Info("gov hook: proposal voting period ended (post-tally diagnostic)",
		"proposal_id", proposalID,
		"class", class.String(),
	)
}
