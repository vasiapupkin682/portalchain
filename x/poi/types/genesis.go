package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Reports:     []EpochReport{},
		Reputations: []Reputation{},
	}
}

func (gs GenesisState) Validate() error {
	for _, report := range gs.Reports {
		if report.Validator == "" {
			return ErrInvalidValidator
		}
	}
	for _, rep := range gs.Reputations {
		if rep.Validator == "" {
			return ErrInvalidValidator
		}
	}
	return nil
}
