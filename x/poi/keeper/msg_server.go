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

func NewMsgServerImpl(keeper Keeper) types.FullMsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.FullMsgServer = &msgServer{}

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

	epochStr := strconv.FormatInt(msg.Epoch, 10)
	sampled := k.ShouldSample(ctx, []byte(epochStr))

	if sampled {
		// Defer reputation update until verification completes.
		record := types.SamplingRecord{
			Epoch:     msg.Epoch,
			Validator: msg.Validator,
			Status:    types.SamplingStatusPending,
			Deadline:  ctx.BlockHeight() + types.SamplingVerificationWindow,
		}
		k.SetSamplingRecord(ctx, record)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent("sampling_selected",
				sdk.NewAttribute("epoch", epochStr),
				sdk.NewAttribute("validator", msg.Validator),
				sdk.NewAttribute("deadline", strconv.FormatInt(record.Deadline, 10)),
			),
		)

		k.Logger(ctx).Info("report selected for sampling — reputation deferred",
			"epoch", msg.Epoch,
			"validator", msg.Validator,
			"deadline", record.Deadline,
		)
	} else {
		// Not sampled — update reputation immediately.
		k.UpdateReputation(ctx, report)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_epoch_report",
			sdk.NewAttribute("epoch", epochStr),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("tasks_processed", strconv.FormatInt(msg.TasksProcessed, 10)),
			sdk.NewAttribute("reliability", msg.Reliability.String()),
			sdk.NewAttribute("sampled", strconv.FormatBool(sampled)),
		),
	)

	return &types.MsgSubmitEpochReportResponse{}, nil
}

func (k msgServer) VerifySampling(goCtx context.Context, msg *types.MsgVerifySampling) (*types.MsgVerifySamplingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, found := k.GetSamplingRecord(ctx, msg.Epoch, msg.Validator)
	if !found {
		return nil, types.ErrSamplingNotFound.Wrapf(
			"no sampling record for epoch %d validator %s", msg.Epoch, msg.Validator,
		)
	}

	if record.Status != types.SamplingStatusPending {
		return nil, types.ErrSamplingAlreadyResolved.Wrapf(
			"sampling record status is %q, expected %q", record.Status, types.SamplingStatusPending,
		)
	}

	if ctx.BlockHeight() > record.Deadline {
		return nil, types.ErrSamplingExpired.Wrapf(
			"deadline was block %d, current block is %d", record.Deadline, ctx.BlockHeight(),
		)
	}

	if msg.Verifier == msg.Validator {
		return nil, types.ErrSelfVerification.Wrap("verifier cannot be the same as the validator being verified")
	}

	report, found := k.GetEpochReport(ctx, msg.Epoch, msg.Validator)
	if !found {
		return nil, types.ErrReportNotFound.Wrapf(
			"epoch report not found for epoch %d validator %s", msg.Epoch, msg.Validator,
		)
	}

	if msg.Passed {
		record.Status = types.SamplingStatusVerified
		record.VerifiedBy = msg.Verifier
		k.UpdateReputation(ctx, report)
	} else {
		record.Status = types.SamplingStatusFailed
		record.VerifiedBy = msg.Verifier
		report.SamplingFailures++
		k.SetEpochReport(ctx, report)
		k.UpdateReputation(ctx, report)
	}

	k.SetSamplingRecord(ctx, record)

	resultStr := "passed"
	if !msg.Passed {
		resultStr = "failed"
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"sampling_verified",
			sdk.NewAttribute("epoch", strconv.FormatInt(msg.Epoch, 10)),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("verifier", msg.Verifier),
			sdk.NewAttribute("result", resultStr),
		),
	)

	k.Logger(ctx).Info("sampling verification completed",
		"epoch", msg.Epoch,
		"validator", msg.Validator,
		"verifier", msg.Verifier,
		"result", resultStr,
	)

	return &types.MsgVerifySamplingResponse{}, nil
}
