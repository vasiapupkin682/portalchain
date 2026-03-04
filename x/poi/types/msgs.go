package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitEpochReport = "submit_epoch_report"

var _ sdk.Msg = &MsgSubmitEpochReport{}

func NewMsgSubmitEpochReport(
	epoch int64,
	validator string,
	tasksProcessed int64,
	weightedTaskSum int64,
	avgLatency int64,
	reliability sdk.Dec,
	samplingFailures int64,
	timestamp int64,
) *MsgSubmitEpochReport {
	return &MsgSubmitEpochReport{
		Epoch:            epoch,
		Validator:        validator,
		TasksProcessed:   tasksProcessed,
		WeightedTaskSum:  weightedTaskSum,
		AvgLatency:       avgLatency,
		Reliability:      reliability,
		SamplingFailures: samplingFailures,
		Timestamp:        timestamp,
	}
}

func (msg *MsgSubmitEpochReport) Route() string {
	return RouterKey
}

func (msg *MsgSubmitEpochReport) Type() string {
	return TypeMsgSubmitEpochReport
}

func (msg *MsgSubmitEpochReport) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgSubmitEpochReport) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitEpochReport) ValidateBasic() error {
	if msg.Validator == "" {
		return sdkerrors.Wrap(ErrInvalidValidator, "validator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return sdkerrors.Wrapf(ErrInvalidValidator, "invalid validator address: %s", err)
	}
	if msg.TasksProcessed < 0 {
		return sdkerrors.Wrap(ErrNegativeValue, "tasks_processed must be non-negative")
	}
	if msg.WeightedTaskSum < 0 {
		return sdkerrors.Wrap(ErrNegativeValue, "weighted_task_sum must be non-negative")
	}
	if msg.AvgLatency < 0 {
		return sdkerrors.Wrap(ErrNegativeValue, "avg_latency must be non-negative")
	}
	if msg.SamplingFailures < 0 {
		return sdkerrors.Wrap(ErrNegativeValue, "sampling_failures must be non-negative")
	}
	if msg.Reliability.IsNegative() {
		return sdkerrors.Wrap(ErrInvalidReliability, "reliability must be >= 0")
	}
	if msg.Reliability.GT(sdk.OneDec()) {
		return sdkerrors.Wrap(ErrInvalidReliability, "reliability must be <= 1")
	}
	return nil
}
