package poi

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/keeper"
	"portalchain/x/poi/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, report := range genState.Reports {
		k.SetEpochReport(ctx, report)
	}
	for _, rep := range genState.Reputations {
		k.SetReputation(ctx, rep)
	}
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Reports:     k.GetAllEpochReports(ctx),
		Reputations: k.GetAllReputations(ctx),
	}
}
