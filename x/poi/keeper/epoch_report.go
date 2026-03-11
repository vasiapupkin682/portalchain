package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

func epochReportStoreKey(epoch int64, validator string) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(epoch))
	return append(append([]byte(types.EpochReportPrefix), buf...), []byte(":"+validator)...)
}

func (k Keeper) SetEpochReport(ctx sdk.Context, report types.EpochReport) {
	store := ctx.KVStore(k.storeKey)
	key := epochReportStoreKey(report.Epoch, report.Validator)
	bz := k.cdc.MustMarshal(&report)
	store.Set(key, bz)
}

func (k Keeper) GetEpochReport(ctx sdk.Context, epoch int64, validator string) (types.EpochReport, bool) {
	store := ctx.KVStore(k.storeKey)
	key := epochReportStoreKey(epoch, validator)
	bz := store.Get(key)
	if bz == nil {
		return types.EpochReport{}, false
	}
	var report types.EpochReport
	k.cdc.MustUnmarshal(bz, &report)
	return report, true
}

func (k Keeper) GetReportsByValidator(ctx sdk.Context, validator string) []types.EpochReport {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.EpochReportPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.EpochReport
	suffix := []byte(":" + validator)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		if len(key) >= len(suffix) && string(key[len(key)-len(suffix):]) == string(suffix) {
			var report types.EpochReport
			k.cdc.MustUnmarshal(iterator.Value(), &report)
			reports = append(reports, report)
		}
	}
	return reports
}

func (k Keeper) GetAllEpochReports(ctx sdk.Context) []types.EpochReport {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.EpochReportPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.EpochReport
	for ; iterator.Valid(); iterator.Next() {
		var report types.EpochReport
		k.cdc.MustUnmarshal(iterator.Value(), &report)
		reports = append(reports, report)
	}
	return reports
}

func (k Keeper) UpdateReputation(ctx sdk.Context, report types.EpochReport) {
	currentRep, _ := k.GetReputation(ctx, report.Validator)

	// report.Reliability is sdk.Dec (gogoproto customtype), use it directly
	taskScore := sdk.NewDec(report.TasksProcessed)
	failurePenalty := sdk.NewDec(report.SamplingFailures).Mul(sdk.NewDecWithPrec(1, 1)) // 0.1 per failure

	rawScore := report.Reliability.Mul(taskScore).Sub(failurePenalty)
	if rawScore.IsNegative() {
		rawScore = sdk.ZeroDec()
	}

	// Exponential moving average: new = 0.95 * old + 0.05 * raw_normalized
	// alpha=0.05: reputation changes slowly, single bad report has minimal impact
	alpha := sdk.NewDecWithPrec(5, 2) // 0.05
	oneMinusAlpha := sdk.OneDec().Sub(alpha)

	// maxScore=100: agent processing 50-100 tasks per epoch gets 0.5-1.0 normalized score
	maxScore := sdk.NewDec(100)
	normalizedScore := rawScore.Quo(maxScore)
	if normalizedScore.GT(sdk.OneDec()) {
		normalizedScore = sdk.OneDec()
	}

	newRep := oneMinusAlpha.Mul(currentRep.Value).Add(alpha.Mul(normalizedScore))

	currentRep.Validator = report.Validator
	currentRep.Value = newRep
	k.SetReputation(ctx, currentRep)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"poi_reputation_updated",
			sdk.NewAttribute("validator", report.Validator),
			sdk.NewAttribute("old_reputation", currentRep.Value.String()),
			sdk.NewAttribute("new_reputation", newRep.String()),
			sdk.NewAttribute("raw_score", rawScore.String()),
		),
	)

	// Power adjustment is best-effort; never roll back a report over it
	k.adjustValidatorPower(ctx, report.Validator, newRep)

	// Update category reputation in model-registry (if operator has a registered model)
	taskType := report.TaskType
	if taskType == "" {
		taskType = "general"
	}
	k.UpdateModelCategoryRep(ctx, report.Validator, taskType, normalizedScore)
}

func (k Keeper) adjustValidatorPower(ctx sdk.Context, validator string, newRep sdk.Dec) {
	valAddr, err := sdk.ValAddressFromBech32(validator)
	if err != nil {
		accAddr, accErr := sdk.AccAddressFromBech32(validator)
		if accErr != nil {
			k.Logger(ctx).Error("adjustValidatorPower: unparseable address", "validator", validator, "err", err)
			return
		}
		valAddr = sdk.ValAddress(accAddr)
	}

	val, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		k.Logger(ctx).Info("validator not found for power adjustment, skipping", "validator", validator)
		return
	}

	stake := sdk.NewDecFromInt(val.GetTokens())
	sqrtStake, err := stake.ApproxSqrt()
	if err != nil {
		k.Logger(ctx).Error("adjustValidatorPower: sqrt failed", "validator", validator, "err", err)
		return
	}

	effectivePower := sqrtStake.Mul(sdk.OneDec().Add(newRep))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"poi_power_adjustment",
			sdk.NewAttribute("validator", validator),
			sdk.NewAttribute("effective_power", effectivePower.String()),
			sdk.NewAttribute("reputation", newRep.String()),
		),
	)
}

func (k Keeper) ShouldSample(ctx sdk.Context, seed []byte) bool {
	appHash := ctx.BlockHeader().AppHash

	// Copy to avoid mutating the block header's backing array.
	combined := make([]byte, len(appHash)+len(seed))
	copy(combined, appHash)
	copy(combined[len(appHash):], seed)

	var sum uint64
	for _, b := range combined {
		sum += uint64(b)
	}

	result := sum%10 == 0

	hashHex := fmt.Sprintf("%x", appHash)
	if len(appHash) > 8 {
		hashHex = fmt.Sprintf("%x", appHash[:8])
	}

	k.Logger(ctx).Info("ShouldSample evaluated",
		"app_hash", hashHex,
		"seed", fmt.Sprintf("%x", seed),
		"sum", sum,
		"mod10", sum%10,
		"selected", result,
	)

	return result
}
