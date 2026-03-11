package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

// DistributeRewards runs every RewardInterval blocks, takes a percentage of the
// community pool (DAAI only), and distributes it to eligible AI agents proportionally
// to their PoI reputation score.
func (k Keeper) DistributeRewards(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// Only run every RewardInterval blocks
	if params.RewardInterval <= 0 || ctx.BlockHeight()%params.RewardInterval != 0 {
		return
	}

	// Get community pool balance
	feePool := k.distrKeeper.GetFeePool(ctx)
	communityPool := feePool.CommunityPool

	// Find daai in community pool
	daaiDec := sdk.ZeroDec()
	for _, coin := range communityPool {
		if coin.Denom == "daai" {
			daaiDec = coin.Amount
			break
		}
	}
	if daaiDec.IsZero() {
		return
	}

	// Calculate reward pool = communityPool * RewardPercent
	rewardPool := daaiDec.Mul(params.RewardPercent)
	if rewardPool.LT(sdk.OneDec()) {
		return // too small to distribute
	}

	// Get all reputations
	reputations := k.GetAllReputations(ctx)

	// Filter by MinReputationForReward
	var eligible []types.Reputation
	totalScore := sdk.ZeroDec()
	for _, rep := range reputations {
		if rep.Value.GTE(params.MinReputationForReward) {
			eligible = append(eligible, rep)
			totalScore = totalScore.Add(rep.Value)
		}
	}

	if len(eligible) == 0 || totalScore.IsZero() {
		return
	}

	// Distribute proportionally
	rewardPoolInt := rewardPool.TruncateInt()
	distributed := sdk.ZeroInt()

	for _, rep := range eligible {
		// share = reputation / totalScore
		share := rep.Value.Quo(totalScore)
		agentReward := share.MulInt(rewardPoolInt).TruncateInt()

		if agentReward.IsZero() {
			continue
		}

		agentAddr, err := sdk.AccAddressFromBech32(rep.Validator)
		if err != nil {
			continue
		}

		// Send from community pool to agent
		reward := sdk.NewCoins(sdk.NewCoin("daai", agentReward))
		err = k.distrKeeper.DistributeFromFeePool(ctx, reward, agentAddr)
		if err != nil {
			k.Logger(ctx).Error("failed to distribute reward",
				"agent", rep.Validator,
				"amount", agentReward,
				"error", err,
			)
			continue
		}

		distributed = distributed.Add(agentReward)

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"agent_reward",
			sdk.NewAttribute("agent", rep.Validator),
			sdk.NewAttribute("amount", agentReward.String()),
			sdk.NewAttribute("reputation", rep.Value.String()),
		))

		k.Logger(ctx).Info("agent reward distributed",
			"agent", rep.Validator,
			"amount", agentReward,
			"reputation", rep.Value.String(),
		)
	}

	k.Logger(ctx).Info("reward distribution complete",
		"block", ctx.BlockHeight(),
		"total_distributed", distributed,
		"eligible_agents", len(eligible),
	)
}
