package types

import errorsmod "cosmossdk.io/errors"

var (
    ErrInvalidTask       = errorsmod.Register(ModuleName, 1400, "invalid task")
    ErrTaskNotFound      = errorsmod.Register(ModuleName, 1401, "task not found")
    ErrUnauthorized      = errorsmod.Register(ModuleName, 1402, "unauthorized")
    ErrNoAgentsAvailable = errorsmod.Register(ModuleName, 1403, "no agents available")
    ErrInsufficientFunds = errorsmod.Register(ModuleName, 1404, "insufficient funds")
    ErrTaskAlreadyDone   = errorsmod.Register(ModuleName, 1405, "task already completed")
)
