package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) SubmitEpochReport(goCtx context.Context, msg *types.MsgSubmitEpochReport) (*types.MsgSubmitEpochReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	report := types.EpochReport{
		Epoch:            msg.Epoch,
		Validator:        msg.Validator,
		TasksProcessed:   msg.TasksProcessed,
		WeightedTaskSum:  msg.WeightedTaskSum,
		AvgLatency:       msg.AvgLatency,
		Reliability:      msg.Reliability,
		SamplingFailures: msg.SamplingFailures,
		Timestamp:        msg.Timestamp,
	}

	k.SetEpochReport(ctx, report)
	k.UpdateReputation(ctx, report)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_epoch_report",
			sdk.NewAttribute("epoch", strconv.FormatInt(msg.Epoch, 10)),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("tasks_processed", strconv.FormatInt(msg.TasksProcessed, 10)),
			sdk.NewAttribute("reliability", msg.Reliability.String()),
		),
	)

	return &types.MsgSubmitEpochReportResponse{}, nil
}
