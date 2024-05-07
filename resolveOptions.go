package discovery

// ResolveOptions represents flag values used by Discovery
type ResolveOptions int

const (
	// RoNone represents no resolve options
	RoNone ResolveOptions = 0
	// RoDontResolve is used to indicate that the item should not be automatically resolved
	RoDontResolve ResolveOptions = 1 << 0
	// RoInstanceItem is used to indicate that the item should be created for exclusive
	// use by the caller, and not shared with other callers
	RoInstanceItem ResolveOptions = 1 << 1
	// RoUseAOItem is used to indicate that a mapped AO implementation should be used to wrap
	// the requested item (if and only if the call results in a resolution)
	RoUseAOItem ResolveOptions = 1 << 2
)
