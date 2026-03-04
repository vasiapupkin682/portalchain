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

func (k Keeper) UpdateReputation(ctx sdk.Context, report types.EpochReport) error {
	currentRep, _ := k.GetReputation(ctx, report.Validator)

	taskScore := sdk.NewDec(report.TasksProcessed)
	failurePenalty := sdk.NewDec(report.SamplingFailures).Mul(sdk.NewDecWithPrec(1, 1)) // 0.1 per failure

	rawScore := report.Reliability.Mul(taskScore).Sub(failurePenalty)
	if rawScore.IsNegative() {
		rawScore = sdk.ZeroDec()
	}

	// Exponential moving average: new = 0.8 * old + 0.2 * raw_normalized
	alpha := sdk.NewDecWithPrec(2, 1) // 0.2
	oneMinusAlpha := sdk.OneDec().Sub(alpha)

	maxScore := sdk.NewDec(1000)
	normalizedScore := rawScore.Quo(maxScore)
	if normalizedScore.GT(sdk.OneDec()) {
		normalizedScore = sdk.OneDec()
	}

	newRep := oneMinusAlpha.Mul(currentRep.Value).Add(alpha.Mul(normalizedScore))

	currentRep.Validator = report.Validator
	currentRep.Value = newRep
	k.SetReputation(ctx, currentRep)

	if err := k.adjustValidatorPower(ctx, report.Validator, newRep); err != nil {
		return fmt.Errorf("failed to adjust validator power: %w", err)
	}

	return nil
}

func (k Keeper) adjustValidatorPower(ctx sdk.Context, validator string, newRep sdk.Dec) error {
	valAddr, err := sdk.ValAddressFromBech32(validator)
	if err != nil {
		accAddr, accErr := sdk.AccAddressFromBech32(validator)
		if accErr != nil {
			return fmt.Errorf("invalid address: %s", validator)
		}
		valAddr = sdk.ValAddress(accAddr)
	}

	val, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		k.Logger(ctx).Info("validator not found for power adjustment, skipping", "validator", validator)
		return nil
	}

	stake := sdk.NewDecFromInt(val.GetTokens())
	sqrtStake, err := stake.ApproxSqrt()
	if err != nil {
		return fmt.Errorf("failed to compute sqrt: %w", err)
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

	return nil
}

func (k Keeper) ShouldSample(ctx sdk.Context, seed []byte) bool {
	appHash := ctx.BlockHeader().AppHash
	combined := append(appHash, seed...)
	var sum uint64
	for _, b := range combined {
		sum += uint64(b)
	}
	return sum%10 == 0
}
