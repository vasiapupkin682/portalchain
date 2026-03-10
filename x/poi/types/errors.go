package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidValidator    = sdkerrors.Register(ModuleName, 1100, "invalid validator address")
	ErrNegativeValue       = sdkerrors.Register(ModuleName, 1101, "value must be non-negative")
	ErrInvalidReliability  = sdkerrors.Register(ModuleName, 1102, "reliability must be between 0 and 1")
	ErrReportNotFound      = sdkerrors.Register(ModuleName, 1103, "epoch report not found")
	ErrReputationNotFound  = sdkerrors.Register(ModuleName, 1104, "reputation not found")
	ErrAgentNotFound       = sdkerrors.Register(ModuleName, 1105, "agent record not found")
	ErrAgentConsentInvalid = sdkerrors.Register(ModuleName, 1106, "agent consent signature is invalid")
	ErrGovAuthority        = sdkerrors.Register(ModuleName, 1107, "only governance can remove agents")
)
