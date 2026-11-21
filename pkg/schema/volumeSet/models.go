/* hand-maintained — canonical source of truth for the `volumeset` schema.
 * `validate:` tags + the type-attached Validator methods in validate.go drive
 * validation (replaces the schema_volumeset.go builder rule-trees). See
 * nodelibs/schema/src/volumeSet.ts.
 *
 * Leaf value types carry their own rule (validate.go):
 *   VsCpu/VsMemory → workload Cpu/Memory rules; SnapshotSchedule → ScheduleType
 *   rule + the snapshot-frequency custom; StatusWorkloadLink →
 *   LocalLink+toKind('workload'); PersistentVolumeStatusAttributes →
 *   pattern(String, String.allow('')). The storageClassSuffix / retentionDuration
 *   regexes are the StorageClassSuffix / SnapshotRetentionDuration nominal types.
 * Cross-field .custom()s (volumeSetSpecCustom, mountResourcesCustom,
 * volumeSetStatusDuplicateLocations, volumeSetTopCustom) are object-level
 * .Refine()s layered in each type's ValidationRule. */

package volumeSet

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// --- leaf types ---

type VsCpu string              // → CpuRule()
type VsMemory string           // → MemoryRule()
type SnapshotSchedule string   // → scheduleTypeRule() + snapshotSchedule frequency custom
type StatusWorkloadLink string // → LocalLink.custom(toKind('workload'))

// StorageClassSuffix is the strongly-typed replacement for the
// `pattern=vsStorageClassSuffix` tag on VolumeSetSpec.StorageClassSuffix.
type StorageClassSuffix string

// SnapshotRetentionDuration is the strongly-typed replacement for the
// `pattern=vsSnapshotRetention` tag on SnapshotSpec.RetentionDuration.
type SnapshotRetentionDuration string

type CustomEncryptionRegion struct {
	KeyId string `json:"keyId,omitzero"`
}

type FileSystem struct {
	Name              string   `json:"name"`
	AccessMode        string   `json:"accessMode"`
	CommandsSupported []string `json:"commandsSupported,omitzero"`
}

type FileSystemType string

const (
	FileSystemTypeExt4   FileSystemType = "ext4"
	FileSystemTypeXfs    FileSystemType = "xfs"
	FileSystemTypeShared FileSystemType = "shared"
)

func (FileSystemType) EnumValues() []string { return []string{"ext4", "xfs", "shared"} }

type MountResources struct {
	MaxCpu    string `json:"maxCpu,omitzero"`
	MinCpu    string `json:"minCpu,omitzero"`
	MinMemory string `json:"minMemory,omitzero"`
	MaxMemory string `json:"maxMemory,omitzero"`
}

type PerformanceClassFeaturesSupported string

const (
	PerformanceClassFeaturesSupportedAutomaticExpansion PerformanceClassFeaturesSupported = "automatic-expansion"
	PerformanceClassFeaturesSupportedSnapshots          PerformanceClassFeaturesSupported = "snapshots"
)

type PerformanceClass struct {
	Name              string                              `json:"name"`
	MinCapacity       float64                             `json:"minCapacity"`
	MaxCapacity       float64                             `json:"maxCapacity"`
	FeaturesSupported []PerformanceClassFeaturesSupported `json:"featuresSupported,omitzero"`
}

type PerformanceClassName string

const (
	PerformanceClassNameGeneralPurposeSsd PerformanceClassName = "general-purpose-ssd"
	PerformanceClassNameHighThroughputSsd PerformanceClassName = "high-throughput-ssd"
	PerformanceClassNameShared            PerformanceClassName = "shared"
)

func (PerformanceClassName) EnumValues() []string {
	return []string{"general-purpose-ssd", "high-throughput-ssd", "shared"}
}

type PersistentVolumeStatusLifecycle string

const (
	PersistentVolumeStatusLifecycleCreating  PersistentVolumeStatusLifecycle = "creating"
	PersistentVolumeStatusLifecycleUnused    PersistentVolumeStatusLifecycle = "unused"
	PersistentVolumeStatusLifecycleUnbound   PersistentVolumeStatusLifecycle = "unbound"
	PersistentVolumeStatusLifecycleBound     PersistentVolumeStatusLifecycle = "bound"
	PersistentVolumeStatusLifecycleDeleted   PersistentVolumeStatusLifecycle = "deleted"
	PersistentVolumeStatusLifecycleRepairing PersistentVolumeStatusLifecycle = "repairing"
)

func (PersistentVolumeStatusLifecycle) EnumValues() []string {
	return []string{"creating", "unused", "unbound", "bound", "deleted", "repairing"}
}

// PersistentVolumeStatusAttributes → pattern(String, String().allow(”)) via
// RegisterType (the map value's allow(”) cannot be tagged).
type PersistentVolumeStatusAttributes map[string]string

// DiskType is volumeSet.ts's shared DiskTypes list (volumeSet.ts:238-244): one type for
// every field that spells `.valid(...DiskTypes)`, so a new disk type is added once.
type DiskType string

const (
	DiskTypeGp3                DiskType = "gp3"
	DiskTypePdBalanced         DiskType = "pd-balanced"
	DiskTypePdSsd              DiskType = "pd-ssd"
	DiskTypeHyperdiskBalanced  DiskType = "hyperdisk-balanced"
	DiskTypePremiumV2Lrs       DiskType = "PremiumV2_LRS"
	DiskTypeDoBlockStorage     DiskType = "do-block-storage"
	DiskTypeHcloudVolume       DiskType = "hcloud-volume"
	DiskTypeLinodeBlockStorage DiskType = "linode-block-storage"
	DiskTypeOciBvBalanced      DiskType = "oci-bv-balanced"
	DiskTypeHostpath           DiskType = "hostpath"
	DiskTypeJuicefs            DiskType = "juicefs"
)

func (DiskType) EnumValues() []string {
	return []string{"gp3", "pd-balanced", "pd-ssd", "hyperdisk-balanced", "PremiumV2_LRS",
		"do-block-storage", "hcloud-volume", "linode-block-storage", "oci-bv-balanced",
		"juicefs", "hostpath"}
}

// PersistentVolumeStatusDiskType names the field's type as the generator spelled it.
type PersistentVolumeStatusDiskType = DiskType

type PersistentVolumeStatus struct {
	Lifecycle           PersistentVolumeStatusLifecycle  `json:"lifecycle,omitzero" validate:"default=creating"`
	StorageDeviceId     string                           `json:"storageDeviceId,omitzero"`
	OldStorageDeviceIds []string                         `json:"oldStorageDeviceIds,omitzero"`
	ResourceName        string                           `json:"resourceName,omitzero"`
	Index               float64                          `json:"index" validate:"required,integer,min=0"`
	CurrentSize         float64                          `json:"currentSize" validate:"required,min=0"`
	CurrentBytesUsed    *float64                         `json:"currentBytesUsed,omitzero" validate:"min=0"`
	CurrentBytesFree    *float64                         `json:"currentBytesFree,omitzero" validate:"min=0"`
	Iops                *float64                         `json:"iops,omitzero"`
	Throughput          *float64                         `json:"throughput,omitzero"`
	Driver              string                           `json:"driver" validate:"required"`
	DiskType            PersistentVolumeStatusDiskType   `json:"diskType,omitzero" validate:"optional"`
	VolumeSnapshots     []VolumeSnapshot                 `json:"volumeSnapshots,omitzero"`
	Attributes          PersistentVolumeStatusAttributes `json:"attributes,omitzero"`
	Zone                string                           `json:"zone,omitzero" validate:"allow="`
	Node                string                           `json:"node,omitzero" validate:"optional,allow="`
}

type SnapshotSpec struct {
	CreateFinalSnapshot *bool                     `json:"createFinalSnapshot,omitzero" validate:"default=true"`
	RetentionDuration   SnapshotRetentionDuration `json:"retentionDuration,omitzero"`
	Schedule            SnapshotSchedule          `json:"schedule,omitzero"`
}

// VolumeSetTags aliases base.Tags for the reflection base-type registry.
type VolumeSetTags = base.Tags

type VolumeSet struct {
	Id           string           `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name        `json:"name,omitzero"`
	Kind         base.Kind        `json:"kind,omitzero"`
	Version      *float64         `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string           `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         VolumeSetTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string           `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string           `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links       `json:"links,omitzero" validate:"readOnly"`
	Spec         VolumeSetSpec    `json:"spec" validate:"required"`
	Status       *VolumeSetStatus `json:"status,omitzero"`
	Gvc          any              `json:"gvc,omitzero" validate:"allow="`
}

type VolumeSetConfigLocationVolumeZoneMap map[string]VolumeZoneMap

type VolumeSetConfig struct {
	LocationVolumeZoneMap VolumeSetConfigLocationVolumeZoneMap `json:"locationVolumeZoneMap,omitzero"`
}

type VolumeSetSpecCustomEncryptionRegions map[string]CustomEncryptionRegion

type VolumeSetSpecCustomEncryption struct {
	Regions VolumeSetSpecCustomEncryptionRegions `json:"regions,omitzero"`
}

type VolumeSetSpecAutoscalingPredictive struct {
	Enabled                *bool    `json:"enabled,omitzero" validate:"default=false"`
	LookbackHours          *float64 `json:"lookbackHours,omitzero" validate:"min=1,max=168,default=24"`
	ProjectionHours        *float64 `json:"projectionHours,omitzero" validate:"min=1,max=72,default=6"`
	MinDataPoints          *float64 `json:"minDataPoints,omitzero" validate:"min=2,max=100,integer,default=10"`
	MinGrowthRateGBPerHour *float64 `json:"minGrowthRateGBPerHour,omitzero" validate:"min=0,default=0.01"`
	ScalingFactor          *float64 `json:"scalingFactor,omitzero" validate:"min=1.1"`
}

type VolumeSetSpecAutoscaling struct {
	MaxCapacity       *float64                            `json:"maxCapacity,omitzero" validate:"min=10,integer"`
	MinFreePercentage *float64                            `json:"minFreePercentage,omitzero" validate:"max=100,min=1"`
	ScalingFactor     *float64                            `json:"scalingFactor,omitzero" validate:"min=1.1"`
	Predictive        *VolumeSetSpecAutoscalingPredictive `json:"predictive,omitzero"`
}

// VolumeSetSpecMountOptionsResources ports MountResources (volumeSet.ts:266); the
// cpu/memory ratio+range check is the mountResourcesCustom Refine.
type VolumeSetSpecMountOptionsResources struct {
	MaxCpu    VsCpu    `json:"maxCpu,omitzero" validate:"default=2000m"`
	MinCpu    VsCpu    `json:"minCpu,omitzero" validate:"default=500m"`
	MinMemory VsMemory `json:"minMemory,omitzero" validate:"default=1Gi"`
	MaxMemory VsMemory `json:"maxMemory,omitzero" validate:"default=2Gi"`
}

type VolumeSetSpecMountOptions struct {
	Resources *VolumeSetSpecMountOptionsResources `json:"resources,omitzero"`
}

// VolumeSetSpec ports `VolumeSetSpec` (volumeSet.ts:318). Required + the
// volumeSetSpecCustom cross-field .custom() are the `required` tag on the parent
// field + the VolumeSetSpec Refine.
type VolumeSetSpec struct {
	InitialCapacity    float64                        `json:"initialCapacity" validate:"min=10,integer,required"`
	PerformanceClass   PerformanceClassName           `json:"performanceClass,omitzero"`
	StorageClassSuffix StorageClassSuffix             `json:"storageClassSuffix,omitzero"`
	FileSystemType     FileSystemType                 `json:"fileSystemType,omitzero" validate:"default=ext4"`
	CustomEncryption   *VolumeSetSpecCustomEncryption `json:"customEncryption,omitzero"`
	Snapshots          *SnapshotSpec                  `json:"snapshots,omitzero"`
	Autoscaling        *VolumeSetSpecAutoscaling      `json:"autoscaling,omitzero"`
	MountOptions       *VolumeSetSpecMountOptions     `json:"mountOptions,omitzero"`
}

// VolumeSetStatus ports `VolumeSetStatus` (volumeSet.ts:166); the duplicate-location
// check is the volumeSetStatusDuplicateLocations Refine.
type VolumeSetStatus struct {
	ParentId       string                    `json:"parentId,omitzero"`
	UsedByWorkload StatusWorkloadLink        `json:"usedByWorkload,omitzero"`
	WorkloadLinks  []StatusWorkloadLink      `json:"workloadLinks,omitzero"`
	BindingId      string                    `json:"bindingId,omitzero"`
	Locations      []VolumeSetStatusLocation `json:"locations,omitzero" validate:"default=[]"`
}

type VolumeSetStatusLocation struct {
	Name               string                   `json:"name" validate:"required"`
	Volumes            []PersistentVolumeStatus `json:"volumes,omitzero" validate:"default=[]"`
	DesiredVolumeCount *float64                 `json:"desiredVolumeCount,omitzero" validate:"integer,default=0,min=0"`
	ClusterId          string                   `json:"clusterId,omitzero"`
}

type VolumeSnapshotTags map[string]string

type VolumeSnapshot struct {
	Name    string               `json:"name" validate:"required"`
	Id      string               `json:"id,omitzero" validate:"default=,allow="`
	Created string               `json:"created" validate:"date,required"`
	Expires string               `json:"expires,omitzero" validate:"date"`
	Size    *float64             `json:"size,omitzero" validate:"default=0,min=0"`
	Tags    []VolumeSnapshotTags `json:"tags,omitzero"`
}

type VolumeZoneMap map[string]string
