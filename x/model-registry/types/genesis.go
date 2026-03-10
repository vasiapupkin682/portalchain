package types

import "strings"

// GenesisState defines the model registry module's genesis state.
type GenesisState struct {
	Models []ModelRecord `json:"models"`
}

// DefaultGenesis returns the default genesis state (empty registry).
func DefaultGenesis() GenesisState {
	return GenesisState{
		Models: []ModelRecord{},
	}
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	for _, m := range gs.Models {
		if m.Operator == "" {
			return ErrUnauthorized.Wrap("genesis model has empty operator")
		}
		if m.ModelName == "" {
			return ErrInvalidCapabilities.Wrap("genesis model has empty model name")
		}
		if m.Endpoint == "" {
			return ErrInvalidEndpoint.Wrap("genesis model has empty endpoint")
		}
		if !(strings.HasPrefix(m.Endpoint, "http://") || strings.HasPrefix(m.Endpoint, "https://")) {
			return ErrInvalidEndpoint
		}
		if len(m.Capabilities) == 0 {
			return ErrInvalidCapabilities
		}
	}
	return nil
}
