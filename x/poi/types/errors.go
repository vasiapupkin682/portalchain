package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidValidator        = sdkerrors.Register(ModuleName, 1100, "invalid validator address")
	ErrNegativeValue           = sdkerrors.Register(ModuleName, 1101, "value must be non-negative")
	ErrInvalidReliability      = sdkerrors.Register(ModuleName, 1102, "reliability must be between 0 and 1")
	ErrReportNotFound          = sdkerrors.Register(ModuleName, 1103, "epoch report not found")
	ErrReputationNotFound      = sdkerrors.Register(ModuleName, 1104, "reputation not found")
	ErrSamplingNotFound        = sdkerrors.Register(ModuleName, 1108, "sampling record not found")
	ErrSamplingExpired         = sdkerrors.Register(ModuleName, 1109, "sampling verification deadline has passed")
	ErrSelfVerification        = sdkerrors.Register(ModuleName, 1110, "cannot verify own sampling record")
	ErrSamplingAlreadyResolved = sdkerrors.Register(ModuleName, 1111, "sampling record already resolved")
)
