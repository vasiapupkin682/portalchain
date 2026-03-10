package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

func samplingStoreKey(epoch int64, validator string) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(epoch))
	return append(append([]byte(types.SamplingPrefix), buf...), []byte(":"+validator)...)
}

func (k Keeper) SetSamplingRecord(ctx sdk.Context, record types.SamplingRecord) {
	store := ctx.KVStore(k.storeKey)
	key := samplingStoreKey(record.Epoch, record.Validator)
	bz, err := json.Marshal(record)
	if err != nil {
		panic(fmt.Errorf("failed to marshal sampling record: %w", err))
	}
	store.Set(key, bz)
}

func (k Keeper) GetSamplingRecord(ctx sdk.Context, epoch int64, validator string) (types.SamplingRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	key := samplingStoreKey(epoch, validator)
	bz := store.Get(key)
	if bz == nil {
		return types.SamplingRecord{}, false
	}
	var record types.SamplingRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.SamplingRecord{}, false
	}
	return record, true
}

func (k Keeper) GetPendingSamplingRecords(ctx sdk.Context) []types.SamplingRecord {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SamplingPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var records []types.SamplingRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.SamplingRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		if record.Status == types.SamplingStatusPending {
			records = append(records, record)
		}
	}
	return records
}

// ExpirePendingSamplings is called from EndBlock. It finds all pending
// sampling records whose deadline has passed, treats them as failures
// (no verifier responded in time), and applies a reputation penalty.
func (k Keeper) ExpirePendingSamplings(ctx sdk.Context) {
	records := k.GetPendingSamplingRecords(ctx)
	currentHeight := ctx.BlockHeight()

	for _, record := range records {
		if record.Deadline >= currentHeight {
			continue
		}

		report, found := k.GetEpochReport(ctx, record.Epoch, record.Validator)
		if found {
			report.SamplingFailures++
			k.SetEpochReport(ctx, report)
			k.UpdateReputation(ctx, report)
		}

		record.Status = types.SamplingStatusFailed
		k.SetSamplingRecord(ctx, record)

		k.Logger(ctx).Info("sampling record expired — treated as failure",
			"epoch", record.Epoch,
			"validator", record.Validator,
			"deadline", record.Deadline,
			"current_height", currentHeight,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"sampling_expired",
				sdk.NewAttribute("epoch", strconv.FormatInt(record.Epoch, 10)),
				sdk.NewAttribute("validator", record.Validator),
			),
		)
	}
}
