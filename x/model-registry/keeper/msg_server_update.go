package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/types"
)

func (k *msgServer) UpdateModel(goCtx context.Context, msg *types.MsgUpdateModel) (*types.MsgUpdateModelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, found := k.GetModelRecord(ctx, msg.Operator)
	if !found {
		return nil, types.ErrModelNotFound
	}

	if msg.Endpoint != "" {
		record.Endpoint = msg.Endpoint
	}
	if len(msg.Capabilities) > 0 {
		record.Capabilities = msg.Capabilities
	}
	if msg.PricePerTask != "" {
		record.PricePerTask = msg.PricePerTask
	}
	record.Active = msg.Active
	record.UpdatedAt = ctx.BlockHeight()

	k.SetModelRecord(ctx, record)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"model_updated",
			sdk.NewAttribute("operator", msg.Operator),
			sdk.NewAttribute("updated_at", strconv.FormatInt(record.UpdatedAt, 10)),
		),
	)

	k.Logger(ctx).Info("model updated",
		"operator", msg.Operator,
	)

	return &types.MsgUpdateModelResponse{}, nil
}
