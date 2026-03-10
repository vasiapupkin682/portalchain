package keeper_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	constitutiontypes "portalchain/x/constitution/types"
	poitypes "portalchain/x/poi/types"
)

// ---------------------------------------------------------------------------
// ClassSacredViolation
// ---------------------------------------------------------------------------

func TestClassifyProposal_SacredViolation(t *testing.T) {
	k, ctx, _, _, _ := setupKeeper(t)

	// MsgDeleteOwnRecord inside a governance proposal is a sacred violation
	// (S2: only the owner can delete their own record).
	addr := sdk.AccAddress(bytes.Repeat([]byte{0xAA}, 20))
	msgs := []sdk.Msg{
		&constitutiontypes.MsgDeleteOwnRecord{Address: addr.String()},
	}

	class := k.ClassifyProposal(ctx, msgs)
	require.Equal(t, constitutiontypes.ClassSacredViolation, class)
}

// ---------------------------------------------------------------------------
// ClassConstitutional
// ---------------------------------------------------------------------------

func TestClassifyProposal_Constitutional(t *testing.T) {
	k, ctx, _, _, _ := setupKeeper(t)

	authority := sdk.AccAddress(bytes.Repeat([]byte{0xBB}, 20)).String()

	tests := []struct {
		name string
		msgs []sdk.Msg
	}{
		{
			name: "gov MsgUpdateParams",
			msgs: []sdk.Msg{&govv1.MsgUpdateParams{Authority: authority}},
		},
		{
			name: "staking MsgUpdateParams",
			msgs: []sdk.Msg{&stakingtypes.MsgUpdateParams{Authority: authority}},
		},
		{
			name: "MsgRemoveAgent",
			msgs: []sdk.Msg{&poitypes.MsgRemoveAgent{
				Authority:     authority,
				AgentAddress:  sdk.AccAddress(bytes.Repeat([]byte{0xCC}, 20)).String(),
				Reason:        "test",
				AgentConsent:  []byte("sig"),
				ConsentExpiry: 9999999999,
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			class := k.ClassifyProposal(ctx, tc.msgs)
			require.Equal(t, constitutiontypes.ClassConstitutional, class,
				"expected ClassConstitutional for %s", tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// ClassNetworkParam
// ---------------------------------------------------------------------------

func TestClassifyProposal_NetworkParam(t *testing.T) {
	k, ctx, _, _, _ := setupKeeper(t)

	t.Run("empty messages", func(t *testing.T) {
		class := k.ClassifyProposal(ctx, []sdk.Msg{})
		require.Equal(t, constitutiontypes.ClassNetworkParam, class)
	})

	t.Run("unknown message type", func(t *testing.T) {
		// MsgSubmitEpochReport is not in the classifier — falls to default.
		class := k.ClassifyProposal(ctx, []sdk.Msg{
			&poitypes.MsgSubmitEpochReport{},
		})
		require.Equal(t, constitutiontypes.ClassNetworkParam, class)
	})
}

// ---------------------------------------------------------------------------
// Multi-message: highest severity wins
// ---------------------------------------------------------------------------

func TestClassifyProposal_MultiMessage(t *testing.T) {
	k, ctx, _, _, _ := setupKeeper(t)

	authority := sdk.AccAddress(bytes.Repeat([]byte{0xDD}, 20)).String()
	agent := sdk.AccAddress(bytes.Repeat([]byte{0xEE}, 20)).String()

	t.Run("NetworkParam + Constitutional → Constitutional", func(t *testing.T) {
		msgs := []sdk.Msg{
			&poitypes.MsgSubmitEpochReport{}, // NetworkParam (unknown)
			&govv1.MsgUpdateParams{Authority: authority},
		}
		class := k.ClassifyProposal(ctx, msgs)
		require.Equal(t, constitutiontypes.ClassConstitutional, class)
	})

	t.Run("Constitutional + SacredViolation → SacredViolation", func(t *testing.T) {
		msgs := []sdk.Msg{
			&poitypes.MsgRemoveAgent{
				Authority:     authority,
				AgentAddress:  agent,
				Reason:        "test",
				AgentConsent:  []byte("sig"),
				ConsentExpiry: 9999999999,
			},
			&constitutiontypes.MsgDeleteOwnRecord{Address: agent},
		}
		class := k.ClassifyProposal(ctx, msgs)
		require.Equal(t, constitutiontypes.ClassSacredViolation, class)
	})

	t.Run("SacredViolation short-circuits on first match", func(t *testing.T) {
		msgs := []sdk.Msg{
			&constitutiontypes.MsgDeleteOwnRecord{Address: agent},
			&govv1.MsgUpdateParams{Authority: authority},
		}
		class := k.ClassifyProposal(ctx, msgs)
		require.Equal(t, constitutiontypes.ClassSacredViolation, class)
	})
}
