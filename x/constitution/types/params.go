package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ConstitutionParams struct {
	MaxVotingPowerPercent      sdk.Dec `json:"max_voting_power_percent"`
	ConstitutionalQuorum       sdk.Dec `json:"constitutional_quorum"`
	NetworkParamQuorum         sdk.Dec `json:"network_param_quorum"`
	ConstitutionalTimelockDays int64   `json:"constitutional_timelock_days"`
	SacredPrinciplesHash       string  `json:"sacred_principles_hash"`
}

func DefaultParams() ConstitutionParams {
	return ConstitutionParams{
		MaxVotingPowerPercent:      sdk.NewDecWithPrec(15, 2), // 0.15
		ConstitutionalQuorum:       sdk.NewDecWithPrec(66, 2), // 0.66
		NetworkParamQuorum:         sdk.NewDecWithPrec(50, 2), // 0.50
		ConstitutionalTimelockDays: 14,
		SacredPrinciplesHash:       "",
	}
}

func (p ConstitutionParams) Validate() error {
	if p.MaxVotingPowerPercent.IsNegative() || p.MaxVotingPowerPercent.GT(sdk.OneDec()) {
		return ErrSacredPrincipleViolation.Wrap("max_voting_power_percent must be between 0 and 1")
	}
	if p.ConstitutionalQuorum.IsNegative() || p.ConstitutionalQuorum.GT(sdk.OneDec()) {
		return ErrSacredPrincipleViolation.Wrap("constitutional_quorum must be between 0 and 1")
	}
	if p.NetworkParamQuorum.IsNegative() || p.NetworkParamQuorum.GT(sdk.OneDec()) {
		return ErrSacredPrincipleViolation.Wrap("network_param_quorum must be between 0 and 1")
	}
	if p.ConstitutionalTimelockDays < 0 {
		return ErrSacredPrincipleViolation.Wrap("constitutional_timelock_days must be non-negative")
	}
	return nil
}
