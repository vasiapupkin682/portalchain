package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	modelregistrytypes "portalchain/x/model-registry/types"
)

// CheckAndSlash checks if a validator has exceeded SlashThreshold sampling failures,
// slashes a portion of their staked tokens in the model registry, and optionally
// deregisters them after SlashMaxStrikes.
func (k Keeper) CheckAndSlash(ctx sdk.Context, validator string) {
	params := k.GetParams(ctx)

	report, found := k.GetLatestReport(ctx, validator)
	k.Logger(ctx).Info("CheckAndSlash: latest report",
		"found", found,
		"epoch", report.Epoch,
		"sampling_failures", report.SamplingFailures,
		"threshold", params.SlashThreshold,
	)
	if !found || report.SamplingFailures < params.SlashThreshold {
		return
	}

	k.Logger(ctx).Info("CheckAndSlash: threshold check",
		"failures", report.SamplingFailures,
		"threshold", params.SlashThreshold,
		"will_slash", report.SamplingFailures >= params.SlashThreshold,
	)

	rep, found := k.GetReputation(ctx, validator)
	if !found {
		return
	}

	modelRecord, found := k.GetModelRecord(ctx, validator)
	k.Logger(ctx).Info("CheckAndSlash: model record",
		"found", found,
		"staked", modelRecord.StakedAmount,
	)
	if !found {
		return
	}

	if modelRecord.StakedAmount == "" || modelRecord.StakedAmount == "0" {
		return
	}

	stakedCoin, err := sdk.ParseCoinNormalized(modelRecord.StakedAmount)
	if err != nil || stakedCoin.IsZero() {
		return
	}

	slashAmount := params.SlashPercent.MulInt(stakedCoin.Amount).TruncateInt()
	if slashAmount.IsZero() {
		return
	}

	slashCoin := sdk.NewCoins(sdk.NewCoin(stakedCoin.Denom, slashAmount))

	// Send from model-registry module to distribution module
	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		modelregistrytypes.ModuleName,
		distrtypes.ModuleName,
		slashCoin,
	); err != nil {
		k.Logger(ctx).Error("failed to slash agent", "agent", validator, "error", err)
		return
	}

	// Credit community pool
	feePool := k.distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(slashCoin...)...)
	k.distrKeeper.SetFeePool(ctx, feePool)

	// Update staked amount in model record
	newStaked := stakedCoin.Amount.Sub(slashAmount)
	modelRecord.StakedAmount = sdk.NewCoin(stakedCoin.Denom, newStaked).String()
	k.SetModelRecord(ctx, modelRecord)

	// Increment slash strikes
	rep.SlashStrikes++
	k.SetReputation(ctx, rep)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_slashed",
		sdk.NewAttribute("agent", validator),
		sdk.NewAttribute("slash_amount", slashAmount.String()),
		sdk.NewAttribute("strikes", strconv.FormatInt(rep.SlashStrikes, 10)),
		sdk.NewAttribute("remaining_stake", newStaked.String()),
	))

	k.Logger(ctx).Info("agent slashed",
		"agent", validator,
		"slash_amount", slashAmount,
		"strikes", rep.SlashStrikes,
	)

	if rep.SlashStrikes >= params.SlashMaxStrikes {
		k.DeregisterAgent(ctx, validator)
	}
}

// DeregisterAgent returns remaining stake to the operator and removes the model record.
func (k Keeper) DeregisterAgent(ctx sdk.Context, validator string) {
	modelRecord, found := k.GetModelRecord(ctx, validator)
	if !found {
		return
	}

	if modelRecord.StakedAmount != "" {
		stakedCoin, err := sdk.ParseCoinNormalized(modelRecord.StakedAmount)
		if err == nil && !stakedCoin.IsZero() {
			operatorAddr, err := sdk.AccAddressFromBech32(validator)
			if err == nil {
				_ = k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx,
					modelregistrytypes.ModuleName,
					operatorAddr,
					sdk.NewCoins(stakedCoin),
				)
			}
		}
	}

	k.DeleteModelRecord(ctx, validator)

	rep, _ := k.GetReputation(ctx, validator)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_deregistered_slashing",
		sdk.NewAttribute("agent", validator),
		sdk.NewAttribute("strikes", strconv.FormatInt(rep.SlashStrikes, 10)),
	))

	k.Logger(ctx).Info("agent deregistered due to max slash strikes",
		"agent", validator,
	)
}
