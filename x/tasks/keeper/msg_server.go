package keeper

import (
    "context"
    "crypto/sha256"
    "fmt"
    "math/rand"

    sdk "github.com/cosmos/cosmos-sdk/types"

    modelregistrytypes "portalchain/x/model-registry/types"
    "portalchain/x/tasks/types"
)

type msgServer struct{ *Keeper }

func NewMsgServer(k *Keeper) types.MsgServer { return &msgServer{k} }

func (s *msgServer) CreateTask(goCtx context.Context, msg *types.MsgCreateTask) (*types.MsgCreateTaskResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    params := s.GetParams(ctx)
    isFree := s.IsWithinFreeQuota(ctx, msg.Creator, msg.TaskType)

    // Charge if not free
    if !isFree {
        var price sdk.Coin
        switch msg.TaskType {
        case "text":
            price = params.PricePerText
        case "code":
            price = params.PricePerCode
        case "analysis":
            price = params.PricePerAnalysis
        default:
            price = params.PricePerText
        }
        creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
        if err != nil {
            return nil, types.ErrInvalidTask.Wrap("invalid creator address")
        }
        if err := s.bank.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, sdk.NewCoins(price)); err != nil {
            return nil, types.ErrInsufficientFunds.Wrap(err.Error())
        }
    }

    // Select agent by reputation weight
    models := s.modelRegistryKeeper.GetAllActiveModels(ctx)
    if len(models) == 0 {
        return nil, types.ErrNoAgentsAvailable
    }
    agent := selectAgent(models, msg.TaskType)

    // Create task
    taskID := s.NextTaskID(ctx)
    task := types.Task{
        ID:         taskID,
        Creator:    msg.Creator,
        QueryHash:  msg.QueryHash,
        QueryURL:   msg.QueryUrl,
        TaskType:   msg.TaskType,
        Agent:      agent,
        Status:     types.TaskStatusAssigned,
        CreatedAt:  ctx.BlockHeight(),
        Deadline:   ctx.BlockHeight() + params.TaskDeadlineBlocks,
        FreeQuota:  isFree,
    }
    if !isFree {
        switch msg.TaskType {
        case "text":
            task.Reward = params.PricePerText
        case "code":
            task.Reward = params.PricePerCode
        case "analysis":
            task.Reward = params.PricePerAnalysis
        }
    }

    s.SetTask(ctx, task)
    s.IncrementQuota(ctx, msg.Creator, msg.TaskType)

    ctx.EventManager().EmitEvent(sdk.NewEvent(
        "task_created",
        sdk.NewAttribute("task_id", taskID),
        sdk.NewAttribute("creator", msg.Creator),
        sdk.NewAttribute("agent", agent),
        sdk.NewAttribute("task_type", msg.TaskType),
        sdk.NewAttribute("free", fmt.Sprintf("%v", isFree)),
    ))

    return &types.MsgCreateTaskResponse{TaskId: taskID, Agent: agent}, nil
}

func (s *msgServer) SubmitResult(goCtx context.Context, msg *types.MsgSubmitResult) (*types.MsgSubmitResultResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    task, found := s.GetTask(ctx, msg.TaskId)
    if !found {
        return nil, types.ErrTaskNotFound
    }
    if task.Agent != msg.Agent {
        return nil, types.ErrUnauthorized.Wrap("only assigned agent can submit result")
    }
    if task.Status != types.TaskStatusAssigned {
        return nil, types.ErrTaskAlreadyDone
    }

    task.ResultHash = msg.ResultHash
    task.ResultURL = msg.ResultUrl
    task.Status = types.TaskStatusCompleted
    s.SetTask(ctx, task)

    // Pay agent if paid task
    if !task.FreeQuota && task.Reward.IsPositive() {
        agentAddr, err := sdk.AccAddressFromBech32(msg.Agent)
        if err == nil {
            s.bank.SendCoinsFromModuleToAccount(ctx, types.ModuleName, agentAddr, sdk.NewCoins(task.Reward))
        }
    }

    ctx.EventManager().EmitEvent(sdk.NewEvent(
        "task_completed",
        sdk.NewAttribute("task_id", msg.TaskId),
        sdk.NewAttribute("agent", msg.Agent),
        sdk.NewAttribute("result_hash", msg.ResultHash),
    ))

    return &types.MsgSubmitResultResponse{}, nil
}

func selectAgent(models []modelregistrytypes.ModelRecord, taskType string) string {
    // Simple weighted random by category reputation
    type weighted struct {
        endpoint string
        weight   float64
    }
    var candidates []weighted
    for _, m := range models {
        rep := 0.001 // base weight
        switch taskType {
        case "text":
            if r := parseRep(m.RepText); r > 0 { rep = r }
        case "code":
            if r := parseRep(m.RepCode); r > 0 { rep = r }
        case "analysis":
            if r := parseRep(m.RepAnalysis); r > 0 { rep = r }
        }
        candidates = append(candidates, weighted{m.Operator, rep})
    }
    total := 0.0
    for _, c := range candidates { total += c.weight }
    r := rand.Float64() * total
    cumulative := 0.0
    for _, c := range candidates {
        cumulative += c.weight
        if r <= cumulative { return c.endpoint }
    }
    return candidates[len(candidates)-1].endpoint
}

func parseRep(s string) float64 {
    var f float64
    fmt.Sscanf(s, "%f", &f)
    return f
}

// Compute query hash helper
func QueryHash(query string) string {
    h := sha256.Sum256([]byte(query))
    return fmt.Sprintf("%x", h)
}
