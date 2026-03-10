package keeper

import (
	"portalchain/x/model-registry/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = &msgServer{}
