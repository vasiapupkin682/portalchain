package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	poikeeper "portalchain/x/poi/keeper"
	poitypes "portalchain/x/poi/types"
)

// ---------------------------------------------------------------------------
// Mocks for x/poi keeper dependencies
// ---------------------------------------------------------------------------

type mockAccountKeeper struct {
	accounts    map[string]authtypes.AccountI
	moduleAddrs map[string]sdk.AccAddress
}

var _ poitypes.AccountKeeper = (*mockAccountKeeper)(nil)

func (m *mockAccountKeeper) GetAccount(_ sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return m.accounts[addr.String()]
}

func (m *mockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return m.moduleAddrs[name]
}

type mockStakingKeeper struct{}

var _ poitypes.StakingKeeper = (*mockStakingKeeper)(nil)

func (m *mockStakingKeeper) GetValidator(_ sdk.Context, _ sdk.ValAddress) (stakingtypes.Validator, bool) {
	return stakingtypes.Validator{}, false
}
func (m *mockStakingKeeper) GetAllValidators(_ sdk.Context) []stakingtypes.Validator { return nil }

type mockBankKeeper struct{}

var _ poitypes.BankKeeper = (*mockBankKeeper)(nil)

func (m *mockBankKeeper) SpendableCoins(_ sdk.Context, _ sdk.AccAddress) sdk.Coins { return nil }

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ sdk.Context, _ string, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ sdk.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

// ---------------------------------------------------------------------------
// Test setup
// ---------------------------------------------------------------------------

const testChainID = "portalchain-test-1"

func setupPoiKeeper(t *testing.T) (*poikeeper.Keeper, sdk.Context, *mockAccountKeeper) {
	t.Helper()

	storeKey := sdk.NewKVStoreKey(poitypes.StoreKey)
	modelRegistryStoreKey := sdk.NewKVStoreKey("modelregistry")

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(modelRegistryStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	accountK := &mockAccountKeeper{
		accounts:    make(map[string]authtypes.AccountI),
		moduleAddrs: make(map[string]sdk.AccAddress),
	}

	k := poikeeper.NewKeeper(cdc, storeKey, modelRegistryStoreKey, accountK, &mockStakingKeeper{}, &mockBankKeeper{}, distrkeeper.Keeper{})

	ctx := sdk.NewContext(stateStore, tmproto.Header{
		ChainID: testChainID,
		Time:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return k, ctx, accountK
}

// signConsent builds the consent payload and signs it with the given private key.
func signConsent(
	t *testing.T,
	privKey *secp256k1.PrivKey,
	agentAddr, authority, reason, chainID string,
	consentExpiry int64,
) []byte {
	t.Helper()
	expiryStr := strconv.FormatInt(consentExpiry, 10)
	payload := sha256.Sum256([]byte(agentAddr + authority + reason + chainID + expiryStr))
	sig, err := privKey.Sign(payload[:])
	require.NoError(t, err)
	return sig
}

// seedAgent creates a reputation record and a mock account with a public key.
func seedAgent(
	t *testing.T,
	k *poikeeper.Keeper,
	ctx sdk.Context,
	accountK *mockAccountKeeper,
	privKey *secp256k1.PrivKey,
) sdk.AccAddress {
	t.Helper()

	pubKey := privKey.PubKey()
	agentAddr := sdk.AccAddress(pubKey.Address())

	// Store reputation record.
	k.SetReputation(ctx, poitypes.Reputation{
		Validator: agentAddr.String(),
		Value:     sdk.NewDec(100),
	})

	// Create account with public key.
	acc := authtypes.NewBaseAccountWithAddress(agentAddr)
	require.NoError(t, acc.SetPubKey(pubKey))
	accountK.accounts[agentAddr.String()] = acc

	return agentAddr
}

// ---------------------------------------------------------------------------
// TestRemoveAgent_ConsentExpiry
// ---------------------------------------------------------------------------

func TestRemoveAgent_ConsentExpiry(t *testing.T) {
	govAddr := sdk.AccAddress(bytes.Repeat([]byte{0xFF}, 20))

	t.Run("valid consent with future expiry succeeds", func(t *testing.T) {
		k, ctx, accountK := setupPoiKeeper(t)
		accountK.moduleAddrs["gov"] = govAddr

		privKey := secp256k1.GenPrivKey()
		agentAddr := seedAgent(t, k, ctx, accountK, privKey)

		consentExpiry := ctx.BlockTime().Add(1 * time.Hour).Unix()
		sig := signConsent(t, privKey, agentAddr.String(), govAddr.String(),
			"inactive", testChainID, consentExpiry)

		msg := &poitypes.MsgRemoveAgent{
			Authority:     govAddr.String(),
			AgentAddress:  agentAddr.String(),
			Reason:        "inactive",
			AgentConsent:  sig,
			ConsentExpiry: consentExpiry,
		}

		srv := poikeeper.NewMsgServerImpl(*k)
		_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
		require.NoError(t, err)

		// Reputation should be deleted.
		require.False(t, k.HasReputation(ctx, agentAddr.String()))
	})

	t.Run("expired consent is rejected", func(t *testing.T) {
		k, ctx, accountK := setupPoiKeeper(t)
		accountK.moduleAddrs["gov"] = govAddr

		privKey := secp256k1.GenPrivKey()
		agentAddr := seedAgent(t, k, ctx, accountK, privKey)

		consentExpiry := ctx.BlockTime().Add(-1 * time.Hour).Unix() // past
		sig := signConsent(t, privKey, agentAddr.String(), govAddr.String(),
			"inactive", testChainID, consentExpiry)

		msg := &poitypes.MsgRemoveAgent{
			Authority:     govAddr.String(),
			AgentAddress:  agentAddr.String(),
			Reason:        "inactive",
			AgentConsent:  sig,
			ConsentExpiry: consentExpiry,
		}

		srv := poikeeper.NewMsgServerImpl(*k)
		_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrAgentConsentInvalid)
		require.Contains(t, err.Error(), "expired")
	})

	t.Run("invalid signature is rejected", func(t *testing.T) {
		k, ctx, accountK := setupPoiKeeper(t)
		accountK.moduleAddrs["gov"] = govAddr

		agentPrivKey := secp256k1.GenPrivKey()
		agentAddr := seedAgent(t, k, ctx, accountK, agentPrivKey)

		consentExpiry := ctx.BlockTime().Add(1 * time.Hour).Unix()

		// Sign with a DIFFERENT key — the signature won't match.
		wrongKey := secp256k1.GenPrivKey()
		badSig := signConsent(t, wrongKey, agentAddr.String(), govAddr.String(),
			"inactive", testChainID, consentExpiry)

		msg := &poitypes.MsgRemoveAgent{
			Authority:     govAddr.String(),
			AgentAddress:  agentAddr.String(),
			Reason:        "inactive",
			AgentConsent:  badSig,
			ConsentExpiry: consentExpiry,
		}

		srv := poikeeper.NewMsgServerImpl(*k)
		_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrAgentConsentInvalid)
		require.Contains(t, err.Error(), "does not match")
	})

	t.Run("non-gov authority is rejected", func(t *testing.T) {
		k, ctx, accountK := setupPoiKeeper(t)
		accountK.moduleAddrs["gov"] = govAddr

		privKey := secp256k1.GenPrivKey()
		agentAddr := seedAgent(t, k, ctx, accountK, privKey)

		fakeAuthority := sdk.AccAddress(bytes.Repeat([]byte{0xAA}, 20))
		consentExpiry := ctx.BlockTime().Add(1 * time.Hour).Unix()
		sig := signConsent(t, privKey, agentAddr.String(), fakeAuthority.String(),
			"inactive", testChainID, consentExpiry)

		msg := &poitypes.MsgRemoveAgent{
			Authority:     fakeAuthority.String(),
			AgentAddress:  agentAddr.String(),
			Reason:        "inactive",
			AgentConsent:  sig,
			ConsentExpiry: consentExpiry,
		}

		srv := poikeeper.NewMsgServerImpl(*k)
		_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrGovAuthority)
	})

	t.Run("agent not found is rejected", func(t *testing.T) {
		k, ctx, accountK := setupPoiKeeper(t)
		accountK.moduleAddrs["gov"] = govAddr

		// Agent has NO reputation record — skip seedAgent's SetReputation.
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		agentAddr := sdk.AccAddress(pubKey.Address())

		acc := authtypes.NewBaseAccountWithAddress(agentAddr)
		require.NoError(t, acc.SetPubKey(pubKey))
		accountK.accounts[agentAddr.String()] = acc

		consentExpiry := ctx.BlockTime().Add(1 * time.Hour).Unix()
		sig := signConsent(t, privKey, agentAddr.String(), govAddr.String(),
			"inactive", testChainID, consentExpiry)

		msg := &poitypes.MsgRemoveAgent{
			Authority:     govAddr.String(),
			AgentAddress:  agentAddr.String(),
			Reason:        "inactive",
			AgentConsent:  sig,
			ConsentExpiry: consentExpiry,
		}

		srv := poikeeper.NewMsgServerImpl(*k)
		_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poitypes.ErrAgentNotFound)
	})
}

// ---------------------------------------------------------------------------
// TestRemoveAgent_ReplayAttack
// ---------------------------------------------------------------------------

func TestRemoveAgent_ReplayAttack(t *testing.T) {
	govAddr := sdk.AccAddress(bytes.Repeat([]byte{0xFF}, 20))

	k, ctx, accountK := setupPoiKeeper(t)
	accountK.moduleAddrs["gov"] = govAddr

	privKey := secp256k1.GenPrivKey()
	agentAddr := seedAgent(t, k, ctx, accountK, privKey)

	consentExpiry := ctx.BlockTime().Add(1 * time.Hour).Unix()
	sig := signConsent(t, privKey, agentAddr.String(), govAddr.String(),
		"inactive", testChainID, consentExpiry)

	msg := &poitypes.MsgRemoveAgent{
		Authority:     govAddr.String(),
		AgentAddress:  agentAddr.String(),
		Reason:        "inactive",
		AgentConsent:  sig,
		ConsentExpiry: consentExpiry,
	}

	srv := poikeeper.NewMsgServerImpl(*k)

	// First removal succeeds.
	_, err := srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.False(t, k.HasReputation(ctx, agentAddr.String()))

	// Replay with the exact same message fails — the record no longer exists.
	_, err = srv.RemoveAgent(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.ErrorIs(t, err, poitypes.ErrAgentNotFound)
}
