package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterModel{}, "modelregistry/MsgRegisterModel", nil)
	cdc.RegisterConcrete(&MsgUpdateModel{}, "modelregistry/MsgUpdateModel", nil)
	cdc.RegisterConcrete(&MsgDeregisterModel{}, "modelregistry/MsgDeregisterModel", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterModel{},
		&MsgUpdateModel{},
		&MsgDeregisterModel{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
