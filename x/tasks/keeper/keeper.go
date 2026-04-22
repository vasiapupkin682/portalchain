package keeper

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/cometbft/cometbft/libs/log"
    "github.com/cosmos/cosmos-sdk/codec"
    storetypes "github.com/cosmos/cosmos-sdk/store/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

    modelregistrytypes "portalchain/x/model-registry/types"
    "portalchain/x/tasks/types"
)

type ModelRegistryKeeper interface {
    GetAllActiveModels(ctx sdk.Context) []modelregistrytypes.ModelRecord
}

type PoiKeeper interface {
    GetReputation(ctx sdk.Context, validator string) (interface{}, bool)
}

type Keeper struct {
    cdc                 codec.BinaryCodec
    storeKey            storetypes.StoreKey
    bank                bankkeeper.Keeper
    modelRegistryKeeper ModelRegistryKeeper
}

func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
    bank bankkeeper.Keeper,
    modelRegistryKeeper ModelRegistryKeeper,
) *Keeper {
    return &Keeper{
        cdc:                 cdc,
        storeKey:            storeKey,
        bank:                bank,
        modelRegistryKeeper: modelRegistryKeeper,
    }
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
    return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte(types.ParamsKey))
    if bz == nil {
        return types.DefaultParams()
    }
    var params types.Params
    if err := json.Unmarshal(bz, &params); err != nil {
        return types.DefaultParams()
    }
    return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
    store := ctx.KVStore(k.storeKey)
    bz, _ := json.Marshal(params)
    store.Set([]byte(types.ParamsKey), bz)
}

func (k Keeper) NextTaskID(ctx sdk.Context) string {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte(types.TaskCounterKey))
    counter := int64(0)
    if bz != nil {
        json.Unmarshal(bz, &counter)
    }
    counter++
    bz, _ = json.Marshal(counter)
    store.Set([]byte(types.TaskCounterKey), bz)
    return fmt.Sprintf("task-%d", counter)
}

func (k Keeper) GetDailyQuota(ctx sdk.Context, address string) types.DailyQuota {
    store := ctx.KVStore(k.storeKey)
    date := time.Unix(ctx.BlockTime().Unix(), 0).UTC().Format("2006-01-02")
    key := []byte(types.QuotaPrefix + address + ":" + date)
    bz := store.Get(key)
    if bz == nil {
        return types.DailyQuota{Address: address, Date: date}
    }
    var quota types.DailyQuota
    json.Unmarshal(bz, &quota)
    return quota
}

func (k Keeper) SetDailyQuota(ctx sdk.Context, quota types.DailyQuota) {
    store := ctx.KVStore(k.storeKey)
    key := []byte(types.QuotaPrefix + quota.Address + ":" + quota.Date)
    bz, _ := json.Marshal(quota)
    store.Set(key, bz)
}

func (k Keeper) IsWithinFreeQuota(ctx sdk.Context, address string, taskType string) bool {
    params := k.GetParams(ctx)
    quota := k.GetDailyQuota(ctx, address)
    switch taskType {
    case "text":
        return quota.TextCount < params.FreeTextLimit
    case "code":
        return quota.CodeCount < params.FreeCodeLimit
    case "analysis":
        return quota.AnalysisCount < params.FreeAnalysisLimit
    }
    return false
}

func (k Keeper) IncrementQuota(ctx sdk.Context, address string, taskType string) {
    quota := k.GetDailyQuota(ctx, address)
    switch taskType {
    case "text":
        quota.TextCount++
    case "code":
        quota.CodeCount++
    case "analysis":
        quota.AnalysisCount++
    }
    k.SetDailyQuota(ctx, quota)
}
