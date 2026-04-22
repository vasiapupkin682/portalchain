package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
    cdc.RegisterConcrete(&MsgCreateTask{}, "tasks/MsgCreateTask", nil)
    cdc.RegisterConcrete(&MsgSubmitResult{}, "tasks/MsgSubmitResult", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateTask{},
		&MsgSubmitResult{},
	)
}

var (
	Amino = codec.NewLegacyAmino()
)

func init() {
	RegisterCodec(Amino)
}
