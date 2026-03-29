package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Params defines the parameters for the poi module tokenomics and slashing.
type Params struct {
	RewardInterval          int64   // blocks between rewards
	RewardPercent           sdk.Dec // fraction of community pool per interval (e.g. 0.001 = 0.1%)
	MinReputationForReward  sdk.Dec // minimum reputation to receive reward
	SlashThreshold          int64   // sampling failures before slash
	SlashPercent            sdk.Dec // fraction of stake to slash (e.g. 0.10 = 10%)
	SlashMaxStrikes         int64   // strikes before deregister
}

// DefaultParams returns default poi module parameters.
func DefaultParams() Params {
	return Params{
		RewardInterval:         100,
		RewardPercent:          sdk.NewDecWithPrec(5, 3),   // 0.005
		MinReputationForReward: sdk.NewDecWithPrec(1, 3),   // 0.001
		SlashThreshold:         3,
		SlashPercent:            sdk.NewDecWithPrec(10, 2), // 10%
		SlashMaxStrikes:         3,
	}
}
