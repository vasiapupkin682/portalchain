package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	accountKeeper types.AccountKeeper
	stakingKeeper types.StakingKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
