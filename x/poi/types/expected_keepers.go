package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAddress(name string) sdk.AccAddress
}

type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool)
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
}
