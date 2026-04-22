package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg *MsgCreateTask) Route() string { return ModuleName }
func (msg *MsgCreateTask) Type() string  { return "create_task" }
func (msg *MsgCreateTask) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrInvalidTask.Wrap("creator required")
	}
	if msg.QueryHash == "" {
		return ErrInvalidTask.Wrap("query hash required")
	}
	if msg.TaskType == "" {
		return ErrInvalidTask.Wrap("task type required")
	}
	return nil
}
func (msg *MsgCreateTask) GetSignBytes() []byte { return sdk.MustSortJSON([]byte("{}")) }
func (msg *MsgCreateTask) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{addr}
}

func (msg *MsgSubmitResult) Route() string { return ModuleName }
func (msg *MsgSubmitResult) Type() string  { return "submit_result" }
func (msg *MsgSubmitResult) ValidateBasic() error {
	if msg.Agent == "" {
		return ErrInvalidTask.Wrap("agent required")
	}
	if msg.TaskId == "" {
		return ErrInvalidTask.Wrap("task id required")
	}
	if msg.ResultHash == "" {
		return ErrInvalidTask.Wrap("result hash required")
	}
	return nil
}
func (msg *MsgSubmitResult) GetSignBytes() []byte { return sdk.MustSortJSON([]byte("{}")) }
func (msg *MsgSubmitResult) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Agent)
	return []sdk.AccAddress{addr}
}
