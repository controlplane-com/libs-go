/* auto-generated */

package volumeSet

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type CustomEncryptionRegion struct {
	KeyId string `json:"keyId,omitempty"`
}

type FileSystem struct {
	Name              string   `json:"name"`
	AccessMode        string   `json:"accessMode"`
	CommandsSupported []string `json:"commandsSupported,omitempty"`
}

type FileSystemType string

const (
	FileSystemTypeExt4   FileSystemType = "ext4"
	FileSystemTypeXfs    FileSystemType = "xfs"
	FileSystemTypeShared FileSystemType = "shared"
)

type MountResources struct {
	MaxCpu    string `json:"maxCpu,omitempty"`
	MinCpu    string `json:"minCpu,omitempty"`
	MinMemory string `json:"minMemory,omitempty"`
	MaxMemory string `json:"maxMemory,omitempty"`
}

type PerformanceClassFeaturesSupported string

const (
	PerformanceClassFeaturesSupportedAutomaticExpansion PerformanceClassFeaturesSupported = "automatic-expansion"
	PerformanceClassFeaturesSupportedSnapshots          PerformanceClassFeaturesSupported = "snapshots"
)

type PerformanceClass struct {
	Name              string                              `json:"name"`
	MinCapacity       float32                             `json:"minCapacity"`
	MaxCapacity       float32                             `json:"maxCapacity"`
	FeaturesSupported []PerformanceClassFeaturesSupported `json:"featuresSupported,omitempty"`
}

type PerformanceClassName string

const (
	PerformanceClassNameGeneralPurposeSsd PerformanceClassName = "general-purpose-ssd"
	PerformanceClassNameHighThroughputSsd PerformanceClassName = "high-throughput-ssd"
	PerformanceClassNameShared            PerformanceClassName = "shared"
)

type PersistentVolumeStatusLifecycle string

const (
	PersistentVolumeStatusLifecycleCreating  PersistentVolumeStatusLifecycle = "creating"
	PersistentVolumeStatusLifecycleUnused    PersistentVolumeStatusLifecycle = "unused"
	PersistentVolumeStatusLifecycleUnbound   PersistentVolumeStatusLifecycle = "unbound"
	PersistentVolumeStatusLifecycleBound     PersistentVolumeStatusLifecycle = "bound"
	PersistentVolumeStatusLifecycleDeleted   PersistentVolumeStatusLifecycle = "deleted"
	PersistentVolumeStatusLifecycleRepairing PersistentVolumeStatusLifecycle = "repairing"
)

type PersistentVolumeStatusAttributes map[string]string

type PersistentVolumeStatus struct {
	Lifecycle           PersistentVolumeStatusLifecycle  `json:"lifecycle,omitempty"`
	StorageDeviceId     string                           `json:"storageDeviceId,omitempty"`
	OldStorageDeviceIds []string                         `json:"oldStorageDeviceIds,omitempty"`
	ResourceName        string                           `json:"resourceName,omitempty"`
	Index               float32                          `json:"index"`
	CurrentSize         float32                          `json:"currentSize"`
	CurrentBytesUsed    *float32                         `json:"currentBytesUsed,omitempty"`
	CurrentBytesFree    *float32                         `json:"currentBytesFree,omitempty"`
	Iops                *float32                         `json:"iops,omitempty"`
	Throughput          *float32                         `json:"throughput,omitempty"`
	Driver              string                           `json:"driver"`
	VolumeSnapshots     []VolumeSnapshot                 `json:"volumeSnapshots,omitempty"`
	Attributes          PersistentVolumeStatusAttributes `json:"attributes,omitempty"`
	Zone                string                           `json:"zone,omitempty"`
}

type SnapshotSpec struct {
	CreateFinalSnapshot bool   `json:"createFinalSnapshot,omitempty"`
	RetentionDuration   string `json:"retentionDuration,omitempty"`
	Schedule            string `json:"schedule,omitempty"`
}

type VolumeSetTags map[string]any

type VolumeSet struct {
	Id           string          `json:"id,omitempty"`
	Name         base.Name       `json:"name,omitempty"`
	Kind         base.Kind       `json:"kind,omitempty"`
	Version      *float32        `json:"version,omitempty"`
	Description  string          `json:"description,omitempty"`
	Tags         VolumeSetTags   `json:"tags,omitempty"`
	Created      string          `json:"created,omitempty"`
	LastModified string          `json:"lastModified,omitempty"`
	Links        base.Links      `json:"links,omitempty"`
	Spec         VolumeSetSpec   `json:"spec"`
	Status       VolumeSetStatus `json:"status,omitempty"`
	Gvc          any             `json:"gvc,omitempty"`
}

type VolumeSetSpecCustomEncryptionRegions map[string]CustomEncryptionRegion

type VolumeSetSpecCustomEncryption struct {
	Regions VolumeSetSpecCustomEncryptionRegions `json:"regions,omitempty"`
}

type VolumeSetSpecAutoscalingPredictive struct {
	Enabled                bool     `json:"enabled,omitempty"`
	LookbackHours          *float32 `json:"lookbackHours,omitempty"`
	ProjectionHours        *float32 `json:"projectionHours,omitempty"`
	MinDataPoints          *float32 `json:"minDataPoints,omitempty"`
	MinGrowthRateGBPerHour *float32 `json:"minGrowthRateGBPerHour,omitempty"`
	ScalingFactor          *float32 `json:"scalingFactor,omitempty"`
}

type VolumeSetSpecAutoscaling struct {
	MaxCapacity       *float32                            `json:"maxCapacity,omitempty"`
	MinFreePercentage *float32                            `json:"minFreePercentage,omitempty"`
	ScalingFactor     *float32                            `json:"scalingFactor,omitempty"`
	Predictive        *VolumeSetSpecAutoscalingPredictive `json:"predictive,omitempty"`
}

type VolumeSetSpecMountOptionsResources struct {
	MaxCpu    string `json:"maxCpu,omitempty"`
	MinCpu    string `json:"minCpu,omitempty"`
	MinMemory string `json:"minMemory,omitempty"`
	MaxMemory string `json:"maxMemory,omitempty"`
}

type VolumeSetSpecMountOptions struct {
	Resources VolumeSetSpecMountOptionsResources `json:"resources,omitempty"`
}

type VolumeSetSpec struct {
	InitialCapacity    float32                        `json:"initialCapacity"`
	PerformanceClass   PerformanceClassName           `json:"performanceClass,omitempty"`
	StorageClassSuffix string                         `json:"storageClassSuffix,omitempty"`
	FileSystemType     FileSystemType                 `json:"fileSystemType,omitempty"`
	CustomEncryption   *VolumeSetSpecCustomEncryption `json:"customEncryption,omitempty"`
	Snapshots          *SnapshotSpec                  `json:"snapshots,omitempty"`
	Autoscaling        *VolumeSetSpecAutoscaling      `json:"autoscaling,omitempty"`
	MountOptions       *VolumeSetSpecMountOptions     `json:"mountOptions,omitempty"`
}

type VolumeSetStatus struct {
	ParentId       string                    `json:"parentId,omitempty"`
	UsedByWorkload string                    `json:"usedByWorkload,omitempty"`
	WorkloadLinks  []string                  `json:"workloadLinks,omitempty"`
	BindingId      string                    `json:"bindingId,omitempty"`
	Locations      []VolumeSetStatusLocation `json:"locations,omitempty"`
}

type VolumeSetStatusLocation struct {
	Name               string                   `json:"name"`
	Volumes            []PersistentVolumeStatus `json:"volumes,omitempty"`
	DesiredVolumeCount *float32                 `json:"desiredVolumeCount,omitempty"`
	ClusterId          string                   `json:"clusterId,omitempty"`
}

type VolumeSnapshotTags map[string]string

type VolumeSnapshot struct {
	Name    string               `json:"name"`
	Id      string               `json:"id,omitempty"`
	Created string               `json:"created"`
	Expires string               `json:"expires,omitempty"`
	Size    *float32             `json:"size,omitempty"`
	Tags    []VolumeSnapshotTags `json:"tags,omitempty"`
}
