package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/types"
	poitypes "portalchain/x/poi/types"
)

func (k *msgServer) RegisterModel(goCtx context.Context, msg *types.MsgRegisterModel) (*types.MsgRegisterModelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := k.GetModelRecord(ctx, msg.Operator); found {
		return nil, types.ErrModelAlreadyRegistered
	}

	params := k.GetParams(ctx)
	minStake := params.MinStake

	operatorAddr, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, err
	}

	balance := k.bank.GetBalance(ctx, operatorAddr, minStake.Denom)
	if balance.Amount.LT(minStake.Amount) {
		return nil, types.ErrInsufficientStake.Wrapf(
			"required %s, have %s", minStake.String(), balance.String(),
		)
	}

	if err := k.bank.SendCoinsFromAccountToModule(ctx, operatorAddr, types.ModuleName, sdk.NewCoins(minStake)); err != nil {
		return nil, err
	}

	record := types.ModelRecord{
		Operator:      msg.Operator,
		ModelName:     msg.ModelName,
		Endpoint:      msg.Endpoint,
		Capabilities:  msg.Capabilities,
		PricePerTask:  msg.PricePerTask,
		Active:        true,
		RegisteredAt:  ctx.BlockHeight(),
		UpdatedAt:     ctx.BlockHeight(),
		RepText:       "0.001",
		RepCode:       "0.001",
		RepAnalysis:   "0.001",
		RepGeneral:    "0.001",
		StakedAmount:  minStake.String(),
	}

	k.SetModelRecord(ctx, record)
	// Set initial reputation in poi store so decay works correctly
	initialRep := poitypes.Reputation{
		Validator: msg.Operator,
		Value:     sdk.NewDecWithPrec(1, 3), // 0.001
	}
	k.poiKeeper.SetReputation(ctx, initialRep)

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
