package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"portalchain/x/poi/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) EpochReport(goCtx context.Context, req *types.QueryEpochReportRequest) (*types.QueryEpochReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	report, found := k.GetEpochReport(ctx, req.Epoch, req.Validator)
	if !found {
		return nil, status.Error(codes.NotFound, "epoch report not found")
	}

	return &types.QueryEpochReportResponse{Report: report}, nil
}

func (k Keeper) ValidatorReputation(goCtx context.Context, req *types.QueryValidatorReputationRequest) (*types.QueryValidatorReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	rep, found := k.GetReputation(ctx, req.Validator)
	if !found {
		return nil, status.Error(codes.NotFound, "reputation not found")
	}

	return &types.QueryValidatorReputationResponse{Reputation: rep}, nil
}

func (k Keeper) ReportsByValidator(goCtx context.Context, req *types.QueryReportsByValidatorRequest) (*types.QueryReportsByValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	reports := k.GetReportsByValidator(ctx, req.Validator)

	return &types.QueryReportsByValidatorResponse{
		Reports: reports,
	}, nil
}
