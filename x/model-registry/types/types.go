package types

// ModelRecord represents an AI model registered by an operator on-chain.
type ModelRecord struct {
	Operator     string   `json:"operator"`
	ModelName    string   `json:"model_name"`
	Endpoint     string   `json:"endpoint"`
	Capabilities []string `json:"capabilities"`
	PricePerTask string   `json:"price_per_task"`
	Active       bool     `json:"active"`
	RegisteredAt int64    `json:"registered_at"`
	UpdatedAt    int64    `json:"updated_at"`
	// Category-based reputations (sdk.Dec as string, default "0.0")
	RepText     string `json:"rep_text"`
	RepCode     string `json:"rep_code"`
	RepAnalysis string `json:"rep_analysis"`
	RepGeneral  string `json:"rep_general"`
	StakedAmount string `json:"staked_amount"` // e.g. "100portal"
}
