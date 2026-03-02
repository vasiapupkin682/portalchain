package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "portalchain/testutil/keeper"
	"portalchain/x/portalchain/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.PortalchainKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
