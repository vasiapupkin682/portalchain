package keeper

import (
	"portalchain/x/portalchain/types"
)

var _ types.QueryServer = Keeper{}
