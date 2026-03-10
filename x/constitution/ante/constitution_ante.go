package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	constitutionkeeper "portalchain/x/constitution/keeper"
	"portalchain/x/constitution/types"
)

type ConstitutionAnteDecorator struct {
	constitutionKeeper constitutionkeeper.Keeper
}

func NewConstitutionAnteDecorator(keeper constitutionkeeper.Keeper) ConstitutionAnteDecorator {
	return ConstitutionAnteDecorator{constitutionKeeper: keeper}
}

func (d ConstitutionAnteDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *govv1.MsgSubmitProposal:
			if err := d.handleSubmitProposal(ctx, m); err != nil {
				return ctx, err
			}

		case *govv1.MsgVote:
			if err := d.handleVote(ctx, m.Voter, m.ProposalId); err != nil {
				return ctx, err
			}

		case *govv1.MsgVoteWeighted:
			if err := d.handleVote(ctx, m.Voter, m.ProposalId); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

func (d ConstitutionAnteDecorator) handleSubmitProposal(ctx sdk.Context, submitProposal *govv1.MsgSubmitProposal) error {
	// S3: Per-address voting power check on proposer (all proposals).
	proposerAddr, err := sdk.AccAddressFromBech32(submitProposal.Proposer)
	if err != nil {
		return err
	}
	if err := d.constitutionKeeper.CheckVotingPowerLimit(ctx, proposerAddr); err != nil {
		return err
	}

	proposalMsgs, err := submitProposal.GetMsgs()
	if err != nil {
		return err
	}

	class := d.constitutionKeeper.ClassifyProposal(ctx, proposalMsgs)

	// S4: Reject sacred violations outright.
	if class == types.ClassSacredViolation {
		return types.ErrSacredPrincipleViolation.Wrap(
			"this proposal violates an immutable sacred principle and cannot be submitted",
		)
	}

	// Validator concentration check only for constitutional proposals.
	// NetworkParam proposals pass through — on a small testnet with 3-5
	// validators, each naturally holds >15% and blocking all governance
	// would be counterproductive.
	if class == types.ClassConstitutional {
		if err := d.constitutionKeeper.CheckValidatorConcentration(ctx); err != nil {
			return err
		}
	}

	return nil
}

// handleVote enforces S3 on MsgVote and MsgVoteWeighted, with the
// validator concentration check scoped to constitutional proposals only.
func (d ConstitutionAnteDecorator) handleVote(ctx sdk.Context, voter string, proposalID uint64) error {
	voterAddr, err := sdk.AccAddressFromBech32(voter)
	if err != nil {
		return err
	}

	// S3: Per-address voting power check (all proposals).
	if err := d.constitutionKeeper.CheckVotingPowerLimit(ctx, voterAddr); err != nil {
		return err
	}

	// Validator concentration check only if this is a constitutional proposal.
	class, found := d.constitutionKeeper.GetProposalClass(ctx, proposalID)
	if found && class == types.ClassConstitutional {
		if err := d.constitutionKeeper.CheckValidatorConcentration(ctx); err != nil {
			return err
		}
	}

	return nil
}
