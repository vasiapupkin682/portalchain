package portalchain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "portalchain/testutil/keeper"
	"portalchain/testutil/nullify"
	"portalchain/x/portalchain"
	"portalchain/x/portalchain/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.PortalchainKeeper(t)
	portalchain.InitGenesis(ctx, *k, genesisState)
	got := portalchain.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
