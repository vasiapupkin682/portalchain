package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
)

var _ sdk.Msg = &MsgRegisterModel{}
var _ sdk.Msg = &MsgUpdateModel{}
var _ sdk.Msg = &MsgDeregisterModel{}

func (msg *MsgRegisterModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgRegisterModel) ValidateBasic() error {
	if msg.Operator == "" {
		return sdkerrors.Wrap(ErrUnauthorized, "operator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Operator); err != nil {
		return sdkerrors.Wrapf(ErrUnauthorized, "invalid operator address: %s", err)
	}
	if msg.ModelName == "" {
		return sdkerrors.Wrap(ErrInvalidCapabilities, "model name cannot be empty")
	}
	if msg.Endpoint == "" {
		return sdkerrors.Wrap(ErrInvalidEndpoint, "endpoint cannot be empty")
	}
	if !strings.HasPrefix(msg.Endpoint, "http://") && !strings.HasPrefix(msg.Endpoint, "https://") {
		return ErrInvalidEndpoint
	}
	if len(msg.Capabilities) == 0 {
		return ErrInvalidCapabilities
	}
	for _, c := range msg.Capabilities {
		if strings.TrimSpace(c) == "" {
			return ErrInvalidCapabilities
		}
	}
	return nil
}

func (msg *MsgUpdateModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateModel) ValidateBasic() error {
	if msg.Operator == "" {
		return sdkerrors.Wrap(ErrUnauthorized, "operator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Operator); err != nil {
		return sdkerrors.Wrapf(ErrUnauthorized, "invalid operator address: %s", err)
	}
	if msg.Endpoint != "" && !strings.HasPrefix(msg.Endpoint, "http://") && !strings.HasPrefix(msg.Endpoint, "https://") {
		return ErrInvalidEndpoint
	}
	if len(msg.Capabilities) > 0 {
		for _, c := range msg.Capabilities {
			if strings.TrimSpace(c) == "" {
				return ErrInvalidCapabilities
			}
		}
	}
	return nil
}

func (msg *MsgDeregisterModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgDeregisterModel) ValidateBasic() error {
	if msg.Operator == "" {
		return sdkerrors.Wrap(ErrUnauthorized, "operator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Operator); err != nil {
		return sdkerrors.Wrapf(ErrUnauthorized, "invalid operator address: %s", err)
	}
	return nil
}
