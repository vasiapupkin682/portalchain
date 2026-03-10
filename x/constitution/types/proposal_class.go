package types

type ProposalClass int

const (
	ClassSacredViolation ProposalClass = iota
	ClassConstitutional
	ClassNetworkParam
)

func (c ProposalClass) String() string {
	switch c {
	case ClassSacredViolation:
		return "SACRED_VIOLATION"
	case ClassConstitutional:
		return "CONSTITUTIONAL"
	case ClassNetworkParam:
		return "NETWORK_PARAM"
	default:
		return "UNKNOWN"
	}
}
