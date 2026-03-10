package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrAgentConsentRequired     = sdkerrors.Register(ModuleName, 2001, "proposal removing agent requires agent's own signature")
	ErrUnauthorized             = sdkerrors.Register(ModuleName, 2002, "unauthorized: only the record owner can perform this action")
	ErrVotingPowerExceeded      = sdkerrors.Register(ModuleName, 2003, "single address exceeds maximum allowed voting power percentage")
	ErrSacredPrincipleViolation = sdkerrors.Register(ModuleName, 2004, "proposal violates a sacred constitutional principle")
)
