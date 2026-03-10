package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/model-registry/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func modelKey(operator string) []byte {
	return []byte(types.ModelRegistryPrefix + operator)
}

func (k Keeper) SetModelRecord(ctx sdk.Context, record types.ModelRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(record)
	if err != nil {
		panic(fmt.Errorf("failed to marshal model record: %w", err))
	}
	store.Set(modelKey(record.Operator), bz)
}

func (k Keeper) GetModelRecord(ctx sdk.Context, operator string) (types.ModelRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(modelKey(operator))
	if bz == nil {
		return types.ModelRecord{}, false
	}
	var record types.ModelRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.ModelRecord{}, false
	}
	return record, true
}

func (k Keeper) DeleteModelRecord(ctx sdk.Context, operator string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(modelKey(operator))
}

func (k Keeper) GetAllModels(ctx sdk.Context) []types.ModelRecord {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ModelRegistryPrefix)
	iter := store.Iterator(prefix, sdk.PrefixEndBytes(prefix))
	defer iter.Close()

	var result []types.ModelRecord
	for ; iter.Valid(); iter.Next() {
		var record types.ModelRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		result = append(result, record)
	}
	return result
}

func (k Keeper) GetAllActiveModels(ctx sdk.Context) []types.ModelRecord {
	all := k.GetAllModels(ctx)
	var result []types.ModelRecord
	for _, r := range all {
		if r.Active {
			result = append(result, r)
		}
	}
	return result
}
