package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool)
	SetProposal(ctx sdk.Context, proposal govv1.Proposal)
	Tally(ctx sdk.Context, proposal govv1.Proposal) (passes bool, burnDeposits bool, tallyResults govv1.TallyResult)
	IterateActiveProposalsQueue(ctx sdk.Context, endTime time.Time, cb func(proposal govv1.Proposal) bool)
	InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time)
	RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time)
	RefundAndDeleteDeposits(ctx sdk.Context, proposalID uint64)
}

type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool)
	GetAllValidators(ctx sdk.Context) []stakingtypes.Validator
	TotalBondedTokens(ctx sdk.Context) sdk.Int
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int
}

type PoiKeeper interface {
	DeleteReputation(ctx sdk.Context, validator string)
	HasReputation(ctx sdk.Context, validator string) bool
}
