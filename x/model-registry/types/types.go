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
}
