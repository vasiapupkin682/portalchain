package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/constitution/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = &msgServer{}

// DeleteOwnRecord allows an agent to delete their own reputation data from x/poi.
// Sacred principle S2: only the owner can delete their own record.
func (k *msgServer) DeleteOwnRecord(goCtx context.Context, msg *types.MsgDeleteOwnRecord) (*types.MsgDeleteOwnRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.poiKeeper.HasReputation(ctx, msg.Address) {
		return nil, types.ErrUnauthorized.Wrap("no reputation record found for this address")
	}

	k.poiKeeper.DeleteReputation(ctx, msg.Address)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"delete_own_record",
			sdk.NewAttribute("address", msg.Address),
		),
	)

	k.Logger(ctx).Info("reputation record deleted by owner",
		"address", msg.Address,
	)

	return &types.MsgDeleteOwnRecordResponse{}, nil
}
