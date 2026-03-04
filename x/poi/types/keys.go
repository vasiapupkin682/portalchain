package types

const (
	ModuleName = "poi"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
	MemStoreKey = "mem_poi"

	EpochReportPrefix = "EpochReport:"
	ReputationPrefix  = "Reputation:"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func EpochReportKey(epoch int64, validator string) []byte {
	return []byte(EpochReportPrefix + string(rune(epoch)) + ":" + validator)
}

func ReputationKey(validator string) []byte {
	return []byte(ReputationPrefix + validator)
}
