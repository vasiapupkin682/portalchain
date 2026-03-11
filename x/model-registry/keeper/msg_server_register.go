package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/types"
)

func (k *msgServer) RegisterModel(goCtx context.Context, msg *types.MsgRegisterModel) (*types.MsgRegisterModelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := k.GetModelRecord(ctx, msg.Operator); found {
		return nil, types.ErrModelAlreadyRegistered
	}

	record := types.ModelRecord{
		Operator:     msg.Operator,
		ModelName:    msg.ModelName,
		Endpoint:     msg.Endpoint,
		Capabilities: msg.Capabilities,
		PricePerTask: msg.PricePerTask,
		Active:       true,
		RegisteredAt: ctx.BlockHeight(),
		UpdatedAt:    ctx.BlockHeight(),
		RepText:      "0.0",
		RepCode:      "0.0",
		RepAnalysis:  "0.0",
		RepGeneral:   "0.0",
	}

	k.SetModelRecord(ctx, record)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"model_registered",
			sdk.NewAttribute("operator", msg.Operator),
			sdk.NewAttribute("model_name", msg.ModelName),
			sdk.NewAttribute("endpoint", msg.Endpoint),
		),
	)

	k.Logger(ctx).Info("model registered",
		"operator", msg.Operator,
		"model_name", msg.ModelName,
		"endpoint", msg.Endpoint,
	)

	return &types.MsgRegisterModelResponse{}, nil
}
