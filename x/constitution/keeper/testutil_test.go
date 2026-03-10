package keeper_test

import (
	"testing"
	"time"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	constitutionkeeper "portalchain/x/constitution/keeper"
	constitutiontypes "portalchain/x/constitution/types"
)

// ---------------------------------------------------------------------------
// Mock GovKeeper
// ---------------------------------------------------------------------------

type insertQueueCall struct {
	ProposalID uint64
	EndTime    time.Time
}

type removeQueueCall struct {
	ProposalID uint64
	EndTime    time.Time
}

type mockGovKeeper struct {
	proposals       map[uint64]govv1.Proposal
	activeProposals []govv1.Proposal

	setProposalCalls []govv1.Proposal
	insertQueueCalls []insertQueueCall
	removeQueueCalls []removeQueueCall
	refundCalls      []uint64

	tallyPasses bool
	tallyResult govv1.TallyResult
}

var _ constitutiontypes.GovKeeper = (*mockGovKeeper)(nil)

func newMockGovKeeper() *mockGovKeeper {
	return &mockGovKeeper{
		proposals: make(map[uint64]govv1.Proposal),
	}
}

func (m *mockGovKeeper) GetProposal(_ sdk.Context, proposalID uint64) (govv1.Proposal, bool) {
	p, ok := m.proposals[proposalID]
	return p, ok
}

func (m *mockGovKeeper) SetProposal(_ sdk.Context, proposal govv1.Proposal) {
	m.setProposalCalls = append(m.setProposalCalls, proposal)
	m.proposals[proposal.Id] = proposal
}

func (m *mockGovKeeper) Tally(_ sdk.Context, _ govv1.Proposal) (bool, bool, govv1.TallyResult) {
	return m.tallyPasses, false, m.tallyResult
}

func (m *mockGovKeeper) IterateActiveProposalsQueue(_ sdk.Context, endTime time.Time, cb func(govv1.Proposal) bool) {
	for _, p := range m.activeProposals {
		if p.VotingEndTime != nil && !p.VotingEndTime.After(endTime) {
			if cb(p) {
				return
			}
		}
	}
}

func (m *mockGovKeeper) InsertActiveProposalQueue(_ sdk.Context, proposalID uint64, endTime time.Time) {
	m.insertQueueCalls = append(m.insertQueueCalls, insertQueueCall{proposalID, endTime})
}

func (m *mockGovKeeper) RemoveFromActiveProposalQueue(_ sdk.Context, proposalID uint64, endTime time.Time) {
	m.removeQueueCalls = append(m.removeQueueCalls, removeQueueCall{proposalID, endTime})
}

func (m *mockGovKeeper) RefundAndDeleteDeposits(_ sdk.Context, proposalID uint64) {
	m.refundCalls = append(m.refundCalls, proposalID)
}

// ---------------------------------------------------------------------------
// Mock StakingKeeper
// ---------------------------------------------------------------------------

type mockStakingKeeper struct {
	totalBonded     sdk.Int
	delegatorBonded map[string]sdk.Int
	validators      []stakingtypes.Validator
}

var _ constitutiontypes.StakingKeeper = (*mockStakingKeeper)(nil)

func newMockStakingKeeper() *mockStakingKeeper {
	return &mockStakingKeeper{
		totalBonded:     sdk.ZeroInt(),
		delegatorBonded: make(map[string]sdk.Int),
	}
}

func (m *mockStakingKeeper) GetValidator(_ sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	for _, v := range m.validators {
		if v.OperatorAddress == addr.String() {
			return v, true
		}
	}
	return stakingtypes.Validator{}, false
}

func (m *mockStakingKeeper) GetAllValidators(_ sdk.Context) []stakingtypes.Validator {
	return m.validators
}

func (m *mockStakingKeeper) TotalBondedTokens(_ sdk.Context) sdk.Int {
	return m.totalBonded
}

func (m *mockStakingKeeper) GetDelegatorBonded(_ sdk.Context, delegator sdk.AccAddress) sdk.Int {
	if v, ok := m.delegatorBonded[delegator.String()]; ok {
		return v
	}
	return sdk.ZeroInt()
}

// ---------------------------------------------------------------------------
// Mock PoiKeeper
// ---------------------------------------------------------------------------

type mockPoiKeeper struct {
	reputations map[string]bool
	deleteCalls []string
}

var _ constitutiontypes.PoiKeeper = (*mockPoiKeeper)(nil)

func newMockPoiKeeper() *mockPoiKeeper {
	return &mockPoiKeeper{
		reputations: make(map[string]bool),
	}
}

func (m *mockPoiKeeper) HasReputation(_ sdk.Context, validator string) bool {
	return m.reputations[validator]
}

func (m *mockPoiKeeper) DeleteReputation(_ sdk.Context, validator string) {
	delete(m.reputations, validator)
	m.deleteCalls = append(m.deleteCalls, validator)
}

// ---------------------------------------------------------------------------
// Test setup
// ---------------------------------------------------------------------------

func setupKeeper(t *testing.T) (constitutionkeeper.Keeper, sdk.Context, *mockGovKeeper, *mockStakingKeeper, *mockPoiKeeper) {
	t.Helper()

	storeKey := sdk.NewKVStoreKey(constitutiontypes.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	govK := newMockGovKeeper()
	stakingK := newMockStakingKeeper()
	poiK := newMockPoiKeeper()

	k := constitutionkeeper.NewKeeper(cdc, storeKey, govK, stakingK, poiK)

	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	k.SetParams(ctx, constitutiontypes.DefaultParams())

	return k, ctx, govK, stakingK, poiK
}

// makeValidator constructs a stakingtypes.Validator with the given tokens and
// bonded status. addrBytes must be exactly 20 bytes.
func makeValidator(addrBytes []byte, tokens int64, bonded bool) stakingtypes.Validator {
	status := stakingtypes.Unbonded
	if bonded {
		status = stakingtypes.Bonded
	}
	return stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(addrBytes).String(),
		Status:          status,
		Tokens:          sdk.NewInt(tokens),
		DelegatorShares: sdk.NewDec(tokens),
	}
}
