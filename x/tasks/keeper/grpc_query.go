package keeper

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"portalchain/x/tasks/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ListTasks(goCtx context.Context, req *types.QueryListTasksRequest) (*types.QueryListTasksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	tasks := k.GetAllTasks(ctx)
	bz, err := json.Marshal(tasks)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryListTasksResponse{TasksJson: string(bz)}, nil
}

func (k Keeper) GetTask(goCtx context.Context, req *types.QueryGetTaskRequest) (*types.QueryGetTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	task, found := k.getTaskByID(ctx, req.TaskId)
	if !found {
		return nil, status.Error(codes.NotFound, "task not found")
	}
	bz, err := json.Marshal(task)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryGetTaskResponse{TaskJson: string(bz)}, nil
}

func (k Keeper) AgentTasks(goCtx context.Context, req *types.QueryAgentTasksRequest) (*types.QueryAgentTasksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	tasks := k.GetPendingTasksForAgent(ctx, req.Agent)
	bz, err := json.Marshal(tasks)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAgentTasksResponse{TasksJson: string(bz)}, nil
}
