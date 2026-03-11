package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/types"
)

func (k *msgServer) DeregisterModel(goCtx context.Context, msg *types.MsgDeregisterModel) (*types.MsgDeregisterModelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, found := k.GetModelRecord(ctx, msg.Operator)
	if !found {
		return nil, types.ErrModelNotFound
	}

	if msg.ModelName != "" && record.ModelName != msg.ModelName {
		return nil, types.ErrModelNotFound.Wrapf(
			"model %q does not match registered model %q for operator %s",
			msg.ModelName, record.ModelName, msg.Operator,
		)
	}

	if record.StakedAmount != "" && record.StakedAmount != "0" {
		stakedCoins, err := sdk.ParseCoinsNormalized(record.StakedAmount)
		if err == nil && !stakedCoins.IsZero() {
			operatorAddr, addrErr := sdk.AccAddressFromBech32(msg.Operator)
			if addrErr == nil {
				if err := k.bank.SendCoinsFromModuleToAccount(ctx, types.ModuleName, operatorAddr, stakedCoins); err != nil {
					k.Logger(ctx).Error("failed to return stake on deregister", "operator", msg.Operator, "err", err)
				}
			}
		}
	}

	k.DeleteModelRecord(ctx, msg.Operator)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"model_deregistered",
			sdk.NewAttribute("operator", msg.Operator),
			sdk.NewAttribute("model_name", record.ModelName),
		),
	)

	k.Logger(ctx).Info("model deregistered",
		"operator", msg.Operator,
		"model_name", record.ModelName,
	)

	return &types.MsgDeregisterModelResponse{}, nil
}
