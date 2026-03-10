package types

type GenesisState struct {
	Params ConstitutionParams `json:"params"`
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
