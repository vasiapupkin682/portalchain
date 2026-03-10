package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeleteOwnRecord = "delete_own_record"

var _ sdk.Msg = &MsgDeleteOwnRecord{}

func NewMsgDeleteOwnRecord(address string) *MsgDeleteOwnRecord {
	return &MsgDeleteOwnRecord{Address: address}
}

func (msg *MsgDeleteOwnRecord) Route() string { return RouterKey }
func (msg *MsgDeleteOwnRecord) Type() string  { return TypeMsgDeleteOwnRecord }

func (msg *MsgDeleteOwnRecord) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgDeleteOwnRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteOwnRecord) ValidateBasic() error {
	if msg.Address == "" {
		return sdkerrors.Wrap(ErrUnauthorized, "address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return sdkerrors.Wrapf(ErrUnauthorized, "invalid address: %s", err)
	}
	return nil
}
