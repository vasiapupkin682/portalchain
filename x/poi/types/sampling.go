package types

const (
	SamplingStatusPending  = "pending"
	SamplingStatusVerified = "verified"
	SamplingStatusFailed   = "failed"

	SamplingVerificationWindow = int64(100) // ~10 minutes at 6s blocks
)

type SamplingRecord struct {
	Epoch      int64  `json:"epoch"`
	Validator  string `json:"validator"`
	Status     string `json:"status"`
	Deadline   int64  `json:"deadline"`
	VerifiedBy string `json:"verified_by"`
}
