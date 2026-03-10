package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrModelAlreadyRegistered = sdkerrors.Register(ModuleName, 1201, "model already registered for this operator")
	ErrModelNotFound          = sdkerrors.Register(ModuleName, 1202, "model record not found")
	ErrInvalidEndpoint        = sdkerrors.Register(ModuleName, 1203, "endpoint must start with http:// or https://")
	ErrUnauthorized           = sdkerrors.Register(ModuleName, 1204, "only the operator can update or deregister their model")
	ErrInvalidCapabilities    = sdkerrors.Register(ModuleName, 1205, "capabilities cannot be empty")
)
