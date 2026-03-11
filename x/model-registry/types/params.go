package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Params defines the model registry module parameters.
type Params struct {
	MinStake sdk.Coin `json:"min_stake"` // minimum stake to register
}

// DefaultParams returns default model registry parameters (100portal for testnet).
func DefaultParams() Params {
	return Params{
		MinStake: sdk.NewCoin("daai", sdk.NewInt(100)),
	}
}
