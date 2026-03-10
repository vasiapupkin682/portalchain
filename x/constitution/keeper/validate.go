package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"portalchain/x/constitution/types"
	poitypes "portalchain/x/poi/types"
)

// ClassifyProposal examines ALL messages inside a governance proposal and
// returns the highest-severity constitutional class. It scans every message
// rather than returning on first match, preventing a lower-severity message
// from masking a higher-severity one later in the array.
//
// Severity order: ClassSacredViolation (0) > ClassConstitutional (1) > ClassNetworkParam (2)
func (k Keeper) ClassifyProposal(ctx sdk.Context, proposalMsgs []sdk.Msg) types.ProposalClass {
	highest := types.ClassNetworkParam

	for _, msg := range proposalMsgs {
		var class types.ProposalClass

		switch msg.(type) {

		// S2: MsgDeleteOwnRecord is a direct user action — only the record
		// owner may execute it. If it appears inside a governance proposal,
		// someone is attempting to delete another agent's data through
		// governance, which violates the S2 sacred principle.
		case *types.MsgDeleteOwnRecord:
			class = types.ClassSacredViolation

		// S1: Agent removal requires supermajority governance (Constitutional).
		// The actual consent signature is verified in x/poi's msg_server.
		// Constitution ensures this always goes through the 66% + timelock path.
		case *poitypes.MsgRemoveAgent:
			class = types.ClassConstitutional

		case *govv1.MsgUpdateParams:
			class = types.ClassConstitutional
		case *stakingtypes.MsgUpdateParams:
			class = types.ClassConstitutional

		default:
			continue
		}

		if class < highest {
			highest = class
		}
		if highest == types.ClassSacredViolation {
			return highest
		}
	}

	return highest
}

// ClassifyProposalByID fetches a proposal from the gov keeper and classifies it.
func (k Keeper) ClassifyProposalByID(ctx sdk.Context, proposalID uint64) types.ProposalClass {
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return types.ClassNetworkParam
	}

	msgs, err := proposal.GetMsgs()
	if err != nil {
		k.Logger(ctx).Error("failed to get proposal messages", "proposal_id", proposalID, "err", err)
		return types.ClassNetworkParam
	}

	return k.ClassifyProposal(ctx, msgs)
}
