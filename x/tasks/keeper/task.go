package keeper

import (
    "encoding/json"

    sdk "github.com/cosmos/cosmos-sdk/types"

    "portalchain/x/tasks/types"
)

func taskKey(id string) []byte {
    return []byte(types.TaskPrefix + id)
}

func (k Keeper) SetTask(ctx sdk.Context, task types.Task) {
    store := ctx.KVStore(k.storeKey)
    bz, _ := json.Marshal(task)
    store.Set(taskKey(task.ID), bz)
}

func (k Keeper) GetTask(ctx sdk.Context, id string) (types.Task, bool) {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(taskKey(id))
    if bz == nil {
        return types.Task{}, false
    }
    var task types.Task
    json.Unmarshal(bz, &task)
    return task, true
}

func (k Keeper) GetAllTasks(ctx sdk.Context) []types.Task {
    store := ctx.KVStore(k.storeKey)
    prefix := []byte(types.TaskPrefix)
    iter := store.Iterator(prefix, sdk.PrefixEndBytes(prefix))
    defer iter.Close()
    var tasks []types.Task
    for ; iter.Valid(); iter.Next() {
        var task types.Task
        json.Unmarshal(iter.Value(), &task)
        tasks = append(tasks, task)
    }
    return tasks
}

func (k Keeper) GetPendingTasksForAgent(ctx sdk.Context, agent string) []types.Task {
    all := k.GetAllTasks(ctx)
    var result []types.Task
    for _, t := range all {
        if t.Agent == agent && t.Status == types.TaskStatusAssigned {
            result = append(result, t)
        }
    }
    return result
}
