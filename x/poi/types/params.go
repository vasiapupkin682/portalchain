package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Params defines the parameters for the poi module tokenomics.
type Params struct {
	RewardInterval          int64   // blocks between rewards
	RewardPercent           sdk.Dec // fraction of community pool per interval (e.g. 0.001 = 0.1%)
	MinReputationForReward  sdk.Dec // minimum reputation to receive reward
}

// DefaultParams returns default poi module parameters.
func DefaultParams() Params {
	return Params{
		RewardInterval:         100,
		RewardPercent:          sdk.NewDecWithPrec(1, 3), // 0.001
		MinReputationForReward: sdk.NewDecWithPrec(1, 2), // 0.01
	}
}
