package modelregistry

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/keeper"
	"portalchain/x/model-registry/types"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, types.DefaultParams())
	for _, record := range genState.Models {
		k.SetModelRecord(ctx, record)
	}
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Models: k.GetAllModels(ctx),
	}
}
