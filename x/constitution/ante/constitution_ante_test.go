package ante_test

import (
	"bytes"
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

	constitutionante "portalchain/x/constitution/ante"
	constitutionkeeper "portalchain/x/constitution/keeper"
	constitutiontypes "portalchain/x/constitution/types"
)

// ---------------------------------------------------------------------------
// Minimal mocks (duplicated here because ante is a different package)
// ---------------------------------------------------------------------------

type mockGovKeeper struct {
	proposals map[uint64]govv1.Proposal
}

var _ constitutiontypes.GovKeeper = (*mockGovKeeper)(nil)

func (m *mockGovKeeper) GetProposal(_ sdk.Context, id uint64) (govv1.Proposal, bool) {
	p, ok := m.proposals[id]
	return p, ok
}
func (m *mockGovKeeper) SetProposal(_ sdk.Context, p govv1.Proposal) {}
func (m *mockGovKeeper) Tally(_ sdk.Context, _ govv1.Proposal) (bool, bool, govv1.TallyResult) {
	return false, false, govv1.TallyResult{}
}
func (m *mockGovKeeper) IterateActiveProposalsQueue(_ sdk.Context, _ time.Time, _ func(govv1.Proposal) bool) {
}
func (m *mockGovKeeper) InsertActiveProposalQueue(_ sdk.Context, _ uint64, _ time.Time)     {}
func (m *mockGovKeeper) RemoveFromActiveProposalQueue(_ sdk.Context, _ uint64, _ time.Time) {}
func (m *mockGovKeeper) RefundAndDeleteDeposits(_ sdk.Context, _ uint64)                    {}

type mockStakingKeeper struct {
	totalBonded     sdk.Int
	delegatorBonded map[string]sdk.Int
	validators      []stakingtypes.Validator
}

var _ constitutiontypes.StakingKeeper = (*mockStakingKeeper)(nil)

func (m *mockStakingKeeper) GetValidator(_ sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	return stakingtypes.Validator{}, false
}
func (m *mockStakingKeeper) GetAllValidators(_ sdk.Context) []stakingtypes.Validator {
	return m.validators
}
func (m *mockStakingKeeper) TotalBondedTokens(_ sdk.Context) sdk.Int { return m.totalBonded }
func (m *mockStakingKeeper) GetDelegatorBonded(_ sdk.Context, delegator sdk.AccAddress) sdk.Int {
	if v, ok := m.delegatorBonded[delegator.String()]; ok {
		return v
	}
	return sdk.ZeroInt()
}

type mockPoiKeeper struct{}

var _ constitutiontypes.PoiKeeper = (*mockPoiKeeper)(nil)

func (m *mockPoiKeeper) HasReputation(_ sdk.Context, _ string) bool { return false }
func (m *mockPoiKeeper) DeleteReputation(_ sdk.Context, _ string)   {}

// ---------------------------------------------------------------------------
// Mock Tx
// ---------------------------------------------------------------------------

type mockTx struct {
	msgs []sdk.Msg
}

func (t mockTx) GetMsgs() []sdk.Msg   { return t.msgs }
func (t mockTx) ValidateBasic() error { return nil }

// ---------------------------------------------------------------------------
// Test setup
// ---------------------------------------------------------------------------

func setupAnteKeeper(t *testing.T) (constitutionkeeper.Keeper, sdk.Context, *mockStakingKeeper) {
	t.Helper()

	storeKey := sdk.NewKVStoreKey(constitutiontypes.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	govK := &mockGovKeeper{proposals: make(map[uint64]govv1.Proposal)}
	stakingK := &mockStakingKeeper{
		totalBonded:     sdk.NewInt(1000),
		delegatorBonded: make(map[string]sdk.Int),
	}
	poiK := &mockPoiKeeper{}

	k := constitutionkeeper.NewKeeper(cdc, storeKey, govK, stakingK, poiK)

	ctx := sdk.NewContext(stateStore, tmproto.Header{
		Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	k.SetParams(ctx, constitutiontypes.DefaultParams())

	return k, ctx, stakingK
}

func passNext(_ sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return sdk.Context{}, nil
}

// ---------------------------------------------------------------------------
// TestAnteHandler_SacredRejection
// ---------------------------------------------------------------------------

func TestAnteHandler_SacredRejection(t *testing.T) {
	k, ctx, stakingK := setupAnteKeeper(t)
	decorator := constitutionante.NewConstitutionAnteDecorator(k)

	proposerAddr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	stakingK.delegatorBonded[proposerAddr.String()] = sdk.NewInt(100) // 10%

	// Wrap MsgDeleteOwnRecord in a governance proposal → sacred violation.
	innerMsg := &constitutiontypes.MsgDeleteOwnRecord{Address: proposerAddr.String()}
	submitProposal, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{innerMsg},
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000))),
		proposerAddr.String(),
		"", "Remove record", "Test sacred violation",
	)
	require.NoError(t, err)

	tx := mockTx{msgs: []sdk.Msg{submitProposal}}
	_, err = decorator.AnteHandle(ctx, tx, false, passNext)

	require.Error(t, err)
	require.ErrorIs(t, err, constitutiontypes.ErrSacredPrincipleViolation)
}

// ---------------------------------------------------------------------------
// TestAnteHandler_VotingPowerS3
// ---------------------------------------------------------------------------

func TestAnteHandler_VotingPowerS3(t *testing.T) {
	t.Run("MsgSubmitProposal from 16% address → error", func(t *testing.T) {
		k, ctx, stakingK := setupAnteKeeper(t)
		decorator := constitutionante.NewConstitutionAnteDecorator(k)

		proposerAddr := sdk.AccAddress(bytes.Repeat([]byte{0x10}, 20))
		stakingK.delegatorBonded[proposerAddr.String()] = sdk.NewInt(160) // 16%

		submitProposal := &govv1.MsgSubmitProposal{
			Proposer: proposerAddr.String(),
			Messages: []*codectypes.Any{},
		}
		tx := mockTx{msgs: []sdk.Msg{submitProposal}}

		_, err := decorator.AnteHandle(ctx, tx, false, passNext)
		require.Error(t, err)
		require.ErrorIs(t, err, constitutiontypes.ErrVotingPowerExceeded)
	})

	t.Run("MsgVote from 16% address → error", func(t *testing.T) {
		k, ctx, stakingK := setupAnteKeeper(t)
		decorator := constitutionante.NewConstitutionAnteDecorator(k)

		voterAddr := sdk.AccAddress(bytes.Repeat([]byte{0x11}, 20))
		stakingK.delegatorBonded[voterAddr.String()] = sdk.NewInt(160) // 16%

		vote := &govv1.MsgVote{
			ProposalId: 1,
			Voter:      voterAddr.String(),
			Option:     govv1.OptionYes,
		}
		tx := mockTx{msgs: []sdk.Msg{vote}}

		_, err := decorator.AnteHandle(ctx, tx, false, passNext)
		require.Error(t, err)
		require.ErrorIs(t, err, constitutiontypes.ErrVotingPowerExceeded)
	})

	t.Run("MsgVote from 14% address → passes", func(t *testing.T) {
		k, ctx, stakingK := setupAnteKeeper(t)
		decorator := constitutionante.NewConstitutionAnteDecorator(k)

		voterAddr := sdk.AccAddress(bytes.Repeat([]byte{0x12}, 20))
		stakingK.delegatorBonded[voterAddr.String()] = sdk.NewInt(140) // 14%

		vote := &govv1.MsgVote{
			ProposalId: 1,
			Voter:      voterAddr.String(),
			Option:     govv1.OptionYes,
		}
		tx := mockTx{msgs: []sdk.Msg{vote}}

		_, err := decorator.AnteHandle(ctx, tx, false, passNext)
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// TestAnteHandler_PassThrough
// ---------------------------------------------------------------------------

func TestAnteHandler_PassThrough(t *testing.T) {
	t.Run("non-governance message passes through unchanged", func(t *testing.T) {
		k, ctx, _ := setupAnteKeeper(t)
		decorator := constitutionante.NewConstitutionAnteDecorator(k)

		// MsgDeleteOwnRecord as a DIRECT tx message (not inside MsgSubmitProposal)
		// is not a governance message, so the ante handler ignores it.
		directMsg := &constitutiontypes.MsgDeleteOwnRecord{
			Address: sdk.AccAddress(bytes.Repeat([]byte{0x20}, 20)).String(),
		}
		tx := mockTx{msgs: []sdk.Msg{directMsg}}

		_, err := decorator.AnteHandle(ctx, tx, false, passNext)
		require.NoError(t, err)
	})

	t.Run("MsgSubmitProposal with NetworkParam class passes", func(t *testing.T) {
		k, ctx, stakingK := setupAnteKeeper(t)
		decorator := constitutionante.NewConstitutionAnteDecorator(k)

		proposerAddr := sdk.AccAddress(bytes.Repeat([]byte{0x21}, 20))
		stakingK.delegatorBonded[proposerAddr.String()] = sdk.NewInt(100) // 10%

		// Empty messages → ClassNetworkParam
		submitProposal := &govv1.MsgSubmitProposal{
			Proposer: proposerAddr.String(),
			Messages: []*codectypes.Any{},
		}
		tx := mockTx{msgs: []sdk.Msg{submitProposal}}

		_, err := decorator.AnteHandle(ctx, tx, false, passNext)
		require.NoError(t, err)
	})
}
