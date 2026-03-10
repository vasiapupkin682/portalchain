package constitution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/constitution/keeper"
	"portalchain/x/constitution/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	k.SetParams(ctx, gs.Params)
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
