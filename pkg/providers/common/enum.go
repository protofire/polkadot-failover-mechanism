package common

// FailOverMode enumerates the values for upgrade mode.
type FailOverMode string

const (
	// FailOverModeDistributed ...
	FailOverModeDistributed FailOverMode = "distributed"
	// FailOverModeSingle ...
	FailOverModeSingle FailOverMode = "single"
)
