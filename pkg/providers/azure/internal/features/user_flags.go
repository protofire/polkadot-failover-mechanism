package features

// UserFeatures represents provider features
type UserFeatures struct {
	PolkadotFailOverFeature PolkadotFailOverFeatures
}

// PolkadotFailOverFeatures represents provider VMs features
type PolkadotFailOverFeatures struct {
	DeleteVmsWithAPIInSingleMode bool
}
