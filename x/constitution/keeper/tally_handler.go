package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"portalchain/x/constitution/types"
)

// EnforceConstitutionalProposals is called from the constitution module's
// EndBlocker, which runs BEFORE gov's EndBlocker. It iterates all active
// proposals whose voting period has ended and, for constitutional proposals,
// applies the 66% supermajority and timelock requirements.
//
// If a constitutional proposal fails the supermajority: reject it and remove
// it from the active queue so gov never processes it.
//
// If the timelock has not elapsed: extend the voting period so gov skips it
// this block and it remains in the queue for future evaluation.
func (k Keeper) EnforceConstitutionalProposals(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockTime := ctx.BlockHeader().Time

	// Collect proposals to process. We must not mutate the iterator's
	// underlying store while iterating, so collect first, then act.
	type pendingAction struct {
		proposal govv1.Proposal
		class    types.ProposalClass
	}
	var actions []pendingAction

	k.govKeeper.IterateActiveProposalsQueue(ctx, blockTime, func(proposal govv1.Proposal) bool {
		class, found := k.GetProposalClass(ctx, proposal.Id)
		if !found || class == types.ClassNetworkParam {
			return false // let gov handle it normally
		}
		actions = append(actions, pendingAction{proposal: proposal, class: class})
		return false
	})

	for _, a := range actions {
		proposal := a.proposal

		switch a.class {
		case types.ClassSacredViolation:
			// Should have been blocked by AnteHandler, but as a safety net
			// reject it now.
			k.rejectProposal(ctx, proposal, "sacred principle violation")

		case types.ClassConstitutional:
			k.enforceConstitutional(ctx, proposal, params)
		}
	}
}

// enforceConstitutional applies the supermajority and timelock checks to a
// constitutional proposal using params SNAPSHOTTED at classification time.
// This prevents retroactive weakening of timelock/quorum via fast proposals.
func (k Keeper) enforceConstitutional(ctx sdk.Context, proposal govv1.Proposal, currentParams types.ConstitutionParams) {
	blockTime := ctx.BlockHeader().Time

	// Use snapshotted params from classification time. Fall back to current
	// params for proposals classified before the snapshot feature was added.
	timelockDays, quorum, snapFound := k.GetProposalSnapshot(ctx, proposal.Id)
	if !snapFound {
		timelockDays = currentParams.ConstitutionalTimelockDays
		quorum = currentParams.ConstitutionalQuorum
	}

	// Step 1: Check timelock — the voting period must have been at least
	// timelockDays long (using the snapshot, not current params).
	if proposal.VotingStartTime != nil {
		timelockDuration := time.Duration(timelockDays) * 24 * time.Hour
		requiredEnd := proposal.VotingStartTime.Add(timelockDuration)

		if blockTime.Before(requiredEnd) {
			oldEndTime := *proposal.VotingEndTime
			proposal.VotingEndTime = &requiredEnd

			k.govKeeper.RemoveFromActiveProposalQueue(ctx, proposal.Id, oldEndTime)
			k.govKeeper.SetProposal(ctx, proposal)
			k.govKeeper.InsertActiveProposalQueue(ctx, proposal.Id, requiredEnd)

			k.Logger(ctx).Info("constitutional timelock not elapsed, extended voting period",
				"proposal_id", proposal.Id,
				"snapshot_timelock_days", timelockDays,
				"new_voting_end", requiredEnd.String(),
			)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"constitutional_timelock_extended",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
					sdk.NewAttribute("new_voting_end", requiredEnd.String()),
				),
			)
			return
		}
	}

	// Step 2: Timelock has passed. Tally votes and check supermajority
	// against the SNAPSHOTTED quorum threshold.
	passes, _, tallyResults := k.govKeeper.Tally(ctx, proposal)

	yesVotes, err := sdk.NewDecFromStr(tallyResults.YesCount)
	if err != nil {
		k.rejectProposal(ctx, proposal, "invalid tally yes count")
		return
	}

	totalVotes := sdk.ZeroDec()
	for _, count := range []string{
		tallyResults.YesCount,
		tallyResults.NoCount,
		tallyResults.AbstainCount,
		tallyResults.NoWithVetoCount,
	} {
		v, parseErr := sdk.NewDecFromStr(count)
		if parseErr != nil {
			continue
		}
		totalVotes = totalVotes.Add(v)
	}

	if totalVotes.IsZero() {
		k.rejectProposal(ctx, proposal, "no votes cast")
		return
	}

	ratio := yesVotes.Quo(totalVotes)
	meetsSupermajority := ratio.GTE(quorum)

	k.Logger(ctx).Info("constitutional tally result",
		"proposal_id", proposal.Id,
		"yes_ratio", ratio.String(),
		"snapshot_quorum", quorum.String(),
		"standard_passes", passes,
		"meets_supermajority", meetsSupermajority,
	)

	if !meetsSupermajority {
		k.rejectProposal(ctx, proposal, fmt.Sprintf(
			"constitutional quorum not met: %.2f%% < %.2f%% required (snapshot)",
			ratio.MulInt64(100).MustFloat64(),
			quorum.MulInt64(100).MustFloat64(),
		))
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"constitutional_proposal_approved",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute("yes_ratio", ratio.String()),
		),
	)
}

// rejectProposal sets a proposal's status to Rejected, removes it from the
// active queue, and refunds deposits — all before gov's EndBlocker runs.
func (k Keeper) rejectProposal(ctx sdk.Context, proposal govv1.Proposal, reason string) {
	oldEndTime := *proposal.VotingEndTime
	proposal.Status = govv1.StatusRejected

	k.govKeeper.SetProposal(ctx, proposal)
	k.govKeeper.RemoveFromActiveProposalQueue(ctx, proposal.Id, oldEndTime)
	k.govKeeper.RefundAndDeleteDeposits(ctx, proposal.Id)

	k.Logger(ctx).Info("constitutional proposal rejected",
		"proposal_id", proposal.Id,
		"reason", reason,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"constitutional_proposal_rejected",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute("reason", reason),
		),
	)
}
