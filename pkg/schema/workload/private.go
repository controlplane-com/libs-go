package workload

// private.go holds the internal/private workload types — the go-libs home of the
// Joi schema marked `.meta({ private: true })` in nodelibs/schema/src/workload.ts:1387.
// WorkloadConfig and its exclusive sub-types are not part of the public `workload`
// resource shape (they are not reachable from any public struct in models.go); they
// are kept out of models.go so the canonical public models can ship on their own.
// These types carry no ValidationRule of their own.

type WorkloadConfigScheduling struct {
	Fingerprint string   `json:"fingerprint,omitzero"`
	Version     *float64 `json:"version,omitzero"`
}

type PodZoneMap map[string]string

type WorkloadConfigPodZoneMap map[string]string

type WorkloadConfigLocationPodZoneMap map[string]PodZoneMap

type WorkloadConfigProxy struct {
	MinCpu *float64 `json:"minCpu,omitzero"`
}

type WorkloadConfigSubsets struct {
	Enabled *bool `json:"enabled,omitzero"`
}

// Capabilities (master ccfd1ae08) and Seccomp (master 40bb798c0) are config-driven
// actuator overrides. Like the rest of WorkloadConfig they are private and unvalidated,
// so drop/keep stay plain []string rather than the LinuxCapabilities enum.
type WorkloadConfigCapabilities struct {
	Drop []string `json:"drop,omitzero"`
	Keep []string `json:"keep,omitzero"`
}

type WorkloadConfigSeccomp struct {
	Unconfined *bool `json:"unconfined,omitzero"`
}

type WorkloadConfig struct {
	Scheduling         *WorkloadConfigScheduling        `json:"scheduling,omitzero"`
	ThinProvision      *float64                         `json:"thinProvision,omitzero"`
	PodZoneMap         WorkloadConfigPodZoneMap         `json:"podZoneMap,omitzero"`
	LocationPodZoneMap WorkloadConfigLocationPodZoneMap `json:"locationPodZoneMap,omitzero"`
	Proxy              *WorkloadConfigProxy             `json:"proxy,omitzero"`
	Subsets            *WorkloadConfigSubsets           `json:"subsets,omitzero"`
	Capabilities       *WorkloadConfigCapabilities      `json:"capabilities,omitzero"`
	Seccomp            *WorkloadConfigSeccomp           `json:"seccomp,omitzero"`
}
