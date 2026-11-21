/* hand-maintained — canonical source of truth for the `command` schema + every
 * typed command spec/status. `validate:` tags drive pkg/validate.FromType
 * reflection (replaces the schema_command.go builder). See
 * nodelibs/schema/src/command.ts.
 *
 * The registered `command` kind uses the generic spec/status (Object.unknown); the
 * typed specs/statuses mirror the Joi source for parity (executor / future use).
 * conflictQuery embeds query.Query (→ QueryRule via the base-type registry);
 * containerOverrides embeds workload.ContainerOverride; volume/snapshot/volumeSet
 * fields embed the volumeSet-owned types; the Expand/Shrink timeoutSeconds default
 * and the Number-keyed / allow('') pattern maps are Refines / RegisterType (see
 * schema_reflect_datanet.go). */

package command

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/query"
import "github.com/controlplane-com/libs-go/pkg/schema/volumeSet"
import "github.com/controlplane-com/libs-go/pkg/schema/workload"
import "github.com/controlplane-com/libs-go/pkg/schema/cronjob"
import "github.com/controlplane-com/libs-go/pkg/schema/containerstatus"

type Cluster struct {
	ClusterId string `json:"clusterId,omitzero"`
	Since     string `json:"since,omitzero" validate:"date"`
}

type Clusters map[string]Cluster

// CommandTags aliases base.Tags for the reflection base-type registry.
type CommandTags = base.Tags

type CommandSpec map[string]any

type CommandStatus map[string]any

type Command struct {
	Id             string                `json:"id,omitzero" validate:"uuid"`
	OwnerId        string                `json:"ownerId,omitzero" validate:"uuid"`
	Kind           base.Kind             `json:"kind,omitzero"`
	Version        *float64              `json:"version,omitzero" validate:"positive,integer"`
	Created        string                `json:"created,omitzero" validate:"date"`
	LastModified   string                `json:"lastModified,omitzero" validate:"date"`
	Links          base.Links            `json:"links,omitzero"`
	Tags           CommandTags           `json:"tags,omitzero" validate:"optional"`
	Type           string                `json:"type" validate:"required"`
	LifecycleStage CommandLifecycleStage `json:"lifecycleStage,omitzero" validate:"default=pending"`
	Spec           CommandSpec           `json:"spec,omitzero"`
	Status         CommandStatus         `json:"status,omitzero"`
	ConflictQuery  *query.Query          `json:"conflictQuery,omitzero" validate:"optional"`
}

type CommandLifecycleStage string //ts:export

const (
	CommandLifecycleStagePending               CommandLifecycleStage = "pending"
	CommandLifecycleStageRunning               CommandLifecycleStage = "running"
	CommandLifecycleStageCancellationRequested CommandLifecycleStage = "cancellation-requested"
	CommandLifecycleStageCancelled             CommandLifecycleStage = "cancelled"
	CommandLifecycleStageCompleted             CommandLifecycleStage = "completed"
	CommandLifecycleStageFailed                CommandLifecycleStage = "failed"
)

func (CommandLifecycleStage) EnumValues() []string {
	return []string{"pending", "running", "cancellation-requested", "cancelled", "completed", "failed"}
}

type CreateVolumeSnapshotSpecSnapshotTags map[string]string

type CreateVolumeSnapshotSpec struct {
	Location               string                                 `json:"location" validate:"required"`
	VolumeIndex            float64                                `json:"volumeIndex" validate:"required"`
	SnapshotName           base.Name                              `json:"snapshotName" validate:"required"`
	SnapshotExpirationDate string                                 `json:"snapshotExpirationDate,omitzero" validate:"date"`
	SnapshotTags           []CreateVolumeSnapshotSpecSnapshotTags `json:"snapshotTags,omitzero"`
}

type CreateVolumeSnapshotStatusStage string

const (
	CreateVolumeSnapshotStatusStageCreateSnapshot  CreateVolumeSnapshotStatusStage = "create-snapshot"
	CreateVolumeSnapshotStatusStageUpdateVolumeSet CreateVolumeSnapshotStatusStage = "update-volume-set"
	CreateVolumeSnapshotStatusStageCleanupK8S      CreateVolumeSnapshotStatusStage = "cleanup-k8s"
	CreateVolumeSnapshotStatusStageRevert          CreateVolumeSnapshotStatusStage = "revert"
)

type CreateVolumeSnapshotStatusClusterIdByLocation map[string]string

type CreateVolumeSnapshotStatus struct {
	Stage               CreateVolumeSnapshotStatusStage               `json:"stage,omitzero" validate:"required,default=create-snapshot,valid=create-snapshot|update-volume-set|cleanup-k8s|revert"`
	Messages            []string                                      `json:"messages,omitzero"`
	ClusterId           string                                        `json:"clusterId,omitzero"`
	ClusterIdByLocation CreateVolumeSnapshotStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	NewSnapshotId       string                                        `json:"newSnapshotId,omitzero"`
	NewSnapshotSize     *float64                                      `json:"newSnapshotSize,omitzero"`
	CreationStartTime   string                                        `json:"creationStartTime,omitzero" validate:"date"`
}

type DeleteCloudDevicesStatusClusterIdByLocation map[string]string

type DeleteCloudDevicesStatusStage string

const (
	DeleteCloudDevicesStatusStageDeleteSnapshots        DeleteCloudDevicesStatusStage = "delete-snapshots"
	DeleteCloudDevicesStatusStageDeleteVolume           DeleteCloudDevicesStatusStage = "delete-volume"
	DeleteCloudDevicesStatusStageFinalizeVolumeDeletion DeleteCloudDevicesStatusStage = "finalize-volume-deletion"
	DeleteCloudDevicesStatusStageUpdateVolumeSet        DeleteCloudDevicesStatusStage = "update-volume-set"
)

type DeleteCloudDevicesStatusPvcRef struct {
	Namespace string `json:"namespace" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type DeleteCloudDevicesStatusSnapshotDeletionStatus map[string]SnapshotDeletionStatus

type DeleteCloudDevicesStatus struct {
	ClusterId              string                                         `json:"clusterId,omitzero"`
	ClusterIdByLocation    DeleteCloudDevicesStatusClusterIdByLocation    `json:"clusterIdByLocation,omitzero"`
	Volume                 volumeSet.PersistentVolumeStatus               `json:"volume" validate:"required"`
	Stage                  DeleteCloudDevicesStatusStage                  `json:"stage,omitzero" validate:"required,default=delete-snapshots,valid=delete-snapshots|delete-volume|finalize-volume-deletion|update-volume-set"`
	Messages               []string                                       `json:"messages,omitzero"`
	PvcRef                 *DeleteCloudDevicesStatusPvcRef                `json:"pvcRef,omitzero"`
	SnapshotDeletionStatus DeleteCloudDevicesStatusSnapshotDeletionStatus `json:"snapshotDeletionStatus" validate:"required,default={}"`
}

type DeleteOrphanedVolumeSnapshotSpec struct {
	Location    string  `json:"location" validate:"required"`
	SnapshotId  string  `json:"snapshotId" validate:"required"`
	VolumeIndex float64 `json:"volumeIndex" validate:"required"`
}

type DeleteOrphanedVolumeSnapshotStatusStage string

const (
	DeleteOrphanedVolumeSnapshotStatusStageDeleteSnapshot        DeleteOrphanedVolumeSnapshotStatusStage = "delete-snapshot"
	DeleteOrphanedVolumeSnapshotStatusStageAwaitSnapshotDeletion DeleteOrphanedVolumeSnapshotStatusStage = "await-snapshot-deletion"
	DeleteOrphanedVolumeSnapshotStatusStageUpdateVolumeSet       DeleteOrphanedVolumeSnapshotStatusStage = "update-volume-set"
	DeleteOrphanedVolumeSnapshotStatusStageFail                  DeleteOrphanedVolumeSnapshotStatusStage = "fail"
)

type DeleteOrphanedVolumeSnapshotStatusClusterIdByLocation map[string]string

type DeleteOrphanedVolumeSnapshotStatus struct {
	Stage               DeleteOrphanedVolumeSnapshotStatusStage               `json:"stage,omitzero" validate:"required,default=delete-snapshot,valid=delete-snapshot|update-volume-set|fail"`
	ClusterId           string                                                `json:"clusterId,omitzero"`
	ClusterIdByLocation DeleteOrphanedVolumeSnapshotStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	Messages            []string                                              `json:"messages,omitzero"`
}

type DeleteOrphanedVolumeSpec struct {
	Location                     string  `json:"location" validate:"required"`
	StorageDeviceId              string  `json:"storageDeviceId" validate:"required"`
	VolumeIndex                  float64 `json:"volumeIndex" validate:"required"`
	NewlyObservedStorageDeviceId string  `json:"newlyObservedStorageDeviceId,omitzero"`
}

type DeleteOrphanedVolumeStatusStage string

const (
	DeleteOrphanedVolumeStatusStageDeleteStorageResources DeleteOrphanedVolumeStatusStage = "delete-storage-resources"
	DeleteOrphanedVolumeStatusStageCleanupK8S             DeleteOrphanedVolumeStatusStage = "cleanup-k8s"
	DeleteOrphanedVolumeStatusStageUpdateVolumeSet        DeleteOrphanedVolumeStatusStage = "update-volume-set"
	DeleteOrphanedVolumeStatusStageFail                   DeleteOrphanedVolumeStatusStage = "fail"
)

type DeleteOrphanedVolumeStatusClusterIdByLocation map[string]string

type DeleteOrphanedVolumeStatus struct {
	Stage               DeleteOrphanedVolumeStatusStage               `json:"stage,omitzero" validate:"required,default=delete-storage-resources,valid=delete-storage-resources|cleanup-k8s|update-volume-set|fail"`
	ClusterId           string                                        `json:"clusterId,omitzero"`
	ClusterIdByLocation DeleteOrphanedVolumeStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	Messages            []string                                      `json:"messages,omitzero"`
}

type DeleteVolumeSetLocationStatusStage string

const (
	DeleteVolumeSetLocationStatusStageDeleteVolumes         DeleteVolumeSetLocationStatusStage = "delete-volumes"
	DeleteVolumeSetLocationStatusStageDeleteOrphanedVolumes DeleteVolumeSetLocationStatusStage = "delete-orphaned-volumes"
	DeleteVolumeSetLocationStatusStageCleanupFilesystem     DeleteVolumeSetLocationStatusStage = "cleanup-filesystem"
	DeleteVolumeSetLocationStatusStageComplete              DeleteVolumeSetLocationStatusStage = "complete"
)

// DeleteVolumeSetLocationStatusVolumes = pattern(Number, DeleteCloudDevicesStatus) —
// the NUMBER key rule is a RegisterType (reflection hardcodes map keys to String).
type DeleteVolumeSetLocationStatusVolumes map[string]DeleteCloudDevicesStatus

type DeleteVolumeSetLocationStatus struct {
	Stage   DeleteVolumeSetLocationStatusStage   `json:"stage,omitzero" validate:"required,default=delete-volumes,valid=delete-volumes|delete-orphaned-volumes|cleanup-filesystem|complete"`
	Volumes DeleteVolumeSetLocationStatusVolumes `json:"volumes,omitzero"`
}

type DeleteVolumeSetSpec struct {
	VolumeSet volumeSet.VolumeSet `json:"volumeSet" validate:"required"`
	// command.ts:405 `Joi.array().items(Joi.string().required())`. The rule is built in
	// validate.go: the engine now has Joi's required-item semantics, so this reports
	// array.includesRequiredUnknowns rather than array.min's different message.
	Locations []string `json:"locations,omitzero"`
}

type DeleteVolumeSetStatus map[string]DeleteVolumeSetLocationStatus

type DeleteVolumeSnapshotSpec struct {
	Location     string  `json:"location" validate:"required"`
	VolumeIndex  float64 `json:"volumeIndex" validate:"required"`
	SnapshotName string  `json:"snapshotName" validate:"required"`
}

type DeleteVolumeSnapshotStatusClusterIdByLocation map[string]string

type DeleteVolumeSnapshotStatusStage string

const (
	DeleteVolumeSnapshotStatusStageDeleteSnapshot        DeleteVolumeSnapshotStatusStage = "delete-snapshot"
	DeleteVolumeSnapshotStatusStageAwaitSnapshotDeletion DeleteVolumeSnapshotStatusStage = "await-snapshot-deletion"
	DeleteVolumeSnapshotStatusStageUpdateVolumeSet       DeleteVolumeSnapshotStatusStage = "update-volume-set"
)

type DeleteVolumeSnapshotStatus struct {
	ClusterId           string                                        `json:"clusterId,omitzero"`
	ClusterIdByLocation DeleteVolumeSnapshotStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	Snapshot            volumeSet.VolumeSnapshot                      `json:"snapshot" validate:"required"`
	Stage               DeleteVolumeSnapshotStatusStage               `json:"stage,omitzero" validate:"required,default=delete-snapshot,valid=delete-snapshot|update-volume-set"`
	Messages            []string                                      `json:"messages,omitzero"`
	SnapshotId          string                                        `json:"snapshotId,omitzero"`
}

type DeleteVolumeSpec struct {
	Location    string  `json:"location" validate:"required"`
	VolumeIndex float64 `json:"volumeIndex" validate:"required"`
}

type DeleteVolumeStatusStage string

const (
	DeleteVolumeStatusStageUpdateVolumeSet         DeleteVolumeStatusStage = "update-volume-set"
	DeleteVolumeStatusStageDeleteStorageResources  DeleteVolumeStatusStage = "delete-storage-resources"
	DeleteVolumeStatusStageShutdownReplica         DeleteVolumeStatusStage = "shutdown-replica"
	DeleteVolumeStatusStageAwaitReplicaTermination DeleteVolumeStatusStage = "await-replica-termination"
	DeleteVolumeStatusStageFail                    DeleteVolumeStatusStage = "fail"
	DeleteVolumeStatusStageCleanupK8S              DeleteVolumeStatusStage = "cleanup-k8s"
)

type DeleteVolumeStatusClusterIdByLocation map[string]string

type DeleteVolumeStatus struct {
	Stage                   DeleteVolumeStatusStage               `json:"stage,omitzero" validate:"required,default=update-volume-set,valid=update-volume-set|delete-storage-resources|shutdown-replica|await-replica-termination|fail|cleanup-k8s"`
	ClusterId               string                                `json:"clusterId,omitzero"`
	ClusterIdByLocation     DeleteVolumeStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	Messages                []string                              `json:"messages,omitzero"`
	InUseByWorkloadId       string                                `json:"inUseByWorkloadId,omitzero"`
	StorageDeviceIdToRemove string                                `json:"storageDeviceIdToRemove,omitzero"`
}

// ExpandVolumeSpec — timeoutSeconds default (600 when not >0) is a Refine.
type ExpandVolumeSpec struct {
	Location           string   `json:"location" validate:"required"`
	VolumeIndex        float64  `json:"volumeIndex" validate:"required"`
	NewStorageCapacity float64  `json:"newStorageCapacity" validate:"required"`
	TimeoutSeconds     *float64 `json:"timeoutSeconds,omitzero"`
}

type ExpandVolumeStatusClusterIdByLocation map[string]string

type ExpandVolumeStatusStage string

const (
	ExpandVolumeStatusStageExpandVolume            ExpandVolumeStatusStage = "expand-volume"
	ExpandVolumeStatusStageDeleteStatefulSet       ExpandVolumeStatusStage = "delete-stateful-set"
	ExpandVolumeStatusStageAwaitOnlineResize       ExpandVolumeStatusStage = "await-online-resize"
	ExpandVolumeStatusStageAwaitReplicaTermination ExpandVolumeStatusStage = "await-replica-termination"
	ExpandVolumeStatusStageAwaitExpansionCompleted ExpandVolumeStatusStage = "await-expansion-completed"
	ExpandVolumeStatusStageUpdateVolumeSet         ExpandVolumeStatusStage = "update-volume-set"
	ExpandVolumeStatusStageRecreateReplica         ExpandVolumeStatusStage = "recreate-replica"
	ExpandVolumeStatusStageCleanupK8S              ExpandVolumeStatusStage = "cleanup-k8s"
	ExpandVolumeStatusStageRevert                  ExpandVolumeStatusStage = "revert"
)

type ExpandVolumeStatus struct {
	ClusterId             string                                `json:"clusterId,omitzero"`
	ClusterIdByLocation   ExpandVolumeStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	Messages              []string                              `json:"messages,omitzero"`
	Stage                 ExpandVolumeStatusStage               `json:"stage,omitzero" validate:"required,default=expand-volume,valid=expand-volume|delete-stateful-set|await-online-resize|await-replica-termination|await-expansion-completed|update-volume-set|recreate-replica|cleanup-k8s|revert"`
	ReplicaRestartedAt    string                                `json:"replicaRestartedAt,omitzero" validate:"date"`
	OnlineResizeStartedAt string                                `json:"onlineResizeStartedAt,omitzero" validate:"date"`
	LockNames             []string                              `json:"lockNames,omitzero"`
	StorageDeviceId       string                                `json:"storageDeviceId,omitzero"`
}

type ReplaceVolumeSpec struct {
	Location    string  `json:"location" validate:"required"`
	VolumeIndex float64 `json:"volumeIndex" validate:"required"`
}

type ReplaceVolumeStatusStage string

const (
	ReplaceVolumeStatusStageCreateVolume               ReplaceVolumeStatusStage = "create-volume"
	ReplaceVolumeStatusStageCleanupAfterVolumeCreation ReplaceVolumeStatusStage = "cleanup-after-volume-creation"
	ReplaceVolumeStatusStageUpdateVolumeSet            ReplaceVolumeStatusStage = "update-volume-set"
	ReplaceVolumeStatusStageConfigureStorageResources  ReplaceVolumeStatusStage = "configure-storage-resources"
	ReplaceVolumeStatusStageRemoveFinalizer            ReplaceVolumeStatusStage = "remove-finalizer"
	ReplaceVolumeStatusStageShutdownReplica            ReplaceVolumeStatusStage = "shutdown-replica"
	ReplaceVolumeStatusStageAwaitReplicaTermination    ReplaceVolumeStatusStage = "await-replica-termination"
	ReplaceVolumeStatusStageCleanupK8S                 ReplaceVolumeStatusStage = "cleanup-k8s"
	ReplaceVolumeStatusStageFail                       ReplaceVolumeStatusStage = "fail"
	ReplaceVolumeStatusStageRevert                     ReplaceVolumeStatusStage = "revert"
	ReplaceVolumeStatusStageCleanupOldStorageDevice    ReplaceVolumeStatusStage = "cleanup-old-storage-device"
	ReplaceVolumeStatusStageRestartReplica             ReplaceVolumeStatusStage = "restart-replica"
)

type ReplaceVolumeStatusClusterIdByLocation map[string]string

// ReplaceVolumeStatusNewVolumeAttributes = pattern(String, String().allow(”)) —
// the map value's allow(”) is a RegisterType.
type ReplaceVolumeStatusNewVolumeAttributes map[string]string

type ReplaceVolumeStatus struct {
	Stage                   ReplaceVolumeStatusStage               `json:"stage,omitzero" validate:"required,default=create-volume,valid=create-volume|cleanup-after-volume-creation|update-volume-set|configure-storage-resources|remove-finalizer|shutdown-replica|await-replica-termination|cleanup-k8s|fail|revert|cleanup-old-storage-device|restart-replica"`
	Messages                []string                               `json:"messages,omitzero"`
	ClusterId               string                                 `json:"clusterId,omitzero"`
	ClusterIdByLocation     ReplaceVolumeStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	InUseByWorkloadId       string                                 `json:"inUseByWorkloadId,omitzero"`
	StorageDeviceIdToRemove string                                 `json:"storageDeviceIdToRemove,omitzero"`
	NewStorageDeviceId      string                                 `json:"newStorageDeviceId,omitzero"`
	NewVolumeAttributes     ReplaceVolumeStatusNewVolumeAttributes `json:"newVolumeAttributes,omitzero"`
	NewResourceName         string                                 `json:"newResourceName,omitzero"`
	NextVolumeSize          *float64                               `json:"nextVolumeSize,omitzero"`
	TempStorageClassName    string                                 `json:"tempStorageClassName,omitzero"`
	LockNames               []string                               `json:"lockNames,omitzero"`
}

type RestoreVolumeSpec struct {
	VolumeIndex  float64 `json:"volumeIndex" validate:"required"`
	Location     string  `json:"location" validate:"required"`
	SnapshotName string  `json:"snapshotName" validate:"required"`
	Zone         string  `json:"zone,omitzero" validate:"allow="`
}

type RunCronWorkloadSpec struct {
	Location              string                       `json:"location" validate:"required"`
	ScheduleName          string                       `json:"scheduleName,omitzero"`
	ContainerOverrides    []workload.ContainerOverride `json:"containerOverrides,omitzero"`
	ActiveDeadlineSeconds *float64                     `json:"activeDeadlineSeconds,omitzero" validate:"integer,min=1"`
}

type RunCronWorkloadStatusClusterIdByLocation map[string]string

type RunCronWorkloadStatusRetryReason string

const (
	RunCronWorkloadStatusRetryReasonAdmissionRejection RunCronWorkloadStatusRetryReason = "admission-rejection"
	RunCronWorkloadStatusRetryReasonNeverStarted       RunCronWorkloadStatusRetryReason = "never-started"
)

type RunCronWorkloadStatusPendingTerminalStage string

const (
	RunCronWorkloadStatusPendingTerminalStageCompleted RunCronWorkloadStatusPendingTerminalStage = "completed"
	RunCronWorkloadStatusPendingTerminalStageFailed    RunCronWorkloadStatusPendingTerminalStage = "failed"
)

type RunCronWorkloadStatusContainers map[string]containerstatus.ContainerStatus

type RunCronWorkloadStatusDetails string

const (
	RunCronWorkloadStatusDetailsInvalid RunCronWorkloadStatusDetails = "invalid"
	RunCronWorkloadStatusDetailsRemoved RunCronWorkloadStatusDetails = "removed"
)

type RunCronWorkloadStatus struct {
	Replica                     string                                    `json:"replica,omitzero"`
	ClusterIdByLocation         RunCronWorkloadStatusClusterIdByLocation  `json:"clusterIdByLocation,omitzero"`
	MinimumWorkloadVersion      *float64                                  `json:"minimumWorkloadVersion,omitzero"`
	Messages                    []string                                  `json:"messages,omitzero"`
	AdmissionRetries            *float64                                  `json:"admissionRetries,omitzero" validate:"integer"`
	FirstAdmissionRejectionTime string                                    `json:"firstAdmissionRejectionTime,omitzero"`
	LastRejectedReplica         string                                    `json:"lastRejectedReplica,omitzero"`
	AdmissionRetryRequested     *bool                                     `json:"admissionRetryRequested,omitzero"`
	RetryReason                 RunCronWorkloadStatusRetryReason          `json:"retryReason,omitzero" validate:"valid=admission-rejection|never-started"`
	InfraRetries                *float64                                  `json:"infraRetries,omitzero" validate:"integer"`
	Disruptions                 *float64                                  `json:"disruptions,omitzero" validate:"integer"`
	LastDisruptedReplica        string                                    `json:"lastDisruptedReplica,omitzero"`
	PendingTerminalStage        RunCronWorkloadStatusPendingTerminalStage `json:"pendingTerminalStage,omitzero" validate:"valid=completed|failed"`
	JobName                     string                                    `json:"jobName,omitzero"`
	WorkloadVersion             *float64                                  `json:"workloadVersion,omitzero" validate:"integer"`
	StartTime                   string                                    `json:"startTime,omitzero" validate:"date"`
	CompletionTime              string                                    `json:"completionTime,omitzero" validate:"date"`
	Conditions                  []cronjob.JobExecutionCondition           `json:"conditions,omitzero"`
	Containers                  RunCronWorkloadStatusContainers           `json:"containers,omitzero"`
	FirstImagePullBackOffTime   string                                    `json:"firstImagePullBackOffTime,omitzero" validate:"date"`
	Details                     RunCronWorkloadStatusDetails              `json:"details,omitzero" validate:"valid=invalid|removed"`
}

// ShrinkVolumeSpec — timeoutSeconds default (600 when not >0) is a Refine.
type ShrinkVolumeSpec struct {
	Location           string   `json:"location" validate:"required"`
	VolumeIndex        float64  `json:"volumeIndex" validate:"required"`
	NewStorageCapacity float64  `json:"newStorageCapacity" validate:"required"`
	TimeoutSeconds     *float64 `json:"timeoutSeconds,omitzero"`
}

// CopyVolumeSpec ports command.ts CopyVolumeSpec (master 8ea018b06..770c7f6af). The
// Joi custom (timeoutSeconds default, the same-location check, the derived `locations`)
// is CopyVolumeSpec.ValidationRule.
type CopyVolumeSpec struct {
	Location                             string   `json:"location" validate:"required"`
	TargetLocation                       string   `json:"targetLocation" validate:"required"`
	VolumeIndex                          float64  `json:"volumeIndex" validate:"required"`
	AcknowledgeDestinationVolumeDeletion *bool    `json:"acknowledgeDestinationVolumeDeletion,omitzero" validate:"default=false"`
	TimeoutSeconds                       *float64 `json:"timeoutSeconds,omitzero" validate:"optional"`
	TargetVolumeSet                      string   `json:"targetVolumeSet,omitzero" validate:"optional"`
	TargetVolumeSetId                    string   `json:"targetVolumeSetId,omitzero" validate:"optional"`
	TargetGvc                            string   `json:"targetGvc,omitzero" validate:"optional"`
	Locations                            []string `json:"locations,omitzero" validate:"optional"`
}

// CopyVolumeSideStatus is one actuator's half of a copy; each side writes only its own.
type CopyVolumeSideStatus struct {
	Stage                CopyVolumeSideStatusStage `json:"stage,omitzero" validate:"required"`
	Messages             []string                  `json:"messages,omitzero"`
	LockNames            []string                  `json:"lockNames,omitzero"`
	TempStorageClassName string                    `json:"tempStorageClassName,omitzero"`
	Complete             *bool                     `json:"complete,omitzero" validate:"default=false"`
	CleanedUp            *bool                     `json:"cleanedUp,omitzero" validate:"default=false"`
}

type CopyVolumeSideStatusStage string

func (CopyVolumeSideStatusStage) EnumValues() []string {
	return []string{
		"capture-source", "await-receiver", "apply-source-egress", "run-bulk-sync",
		"prepare-destination", "apply-receiver", "await-bulk-sync", "commit-and-prepare-swap",
		"shutdown-target-replica", "await-target-replica-termination",
		"cleanup-k8s", "fail",
	}
}

// CopyVolumeStatus ports command.ts CopyVolumeStatus. Like every command status it is
// not validated on write; it is typed for the consumers that read it.
type CopyVolumeStatus struct {
	ClusterId                   string                `json:"clusterId,omitzero"`
	ClusterIdByLocation         map[string]string     `json:"clusterIdByLocation,omitzero"`
	Messages                    []string              `json:"messages,omitzero"`
	Source                      *CopyVolumeSideStatus `json:"source,omitzero"`
	Target                      *CopyVolumeSideStatus `json:"target,omitzero"`
	ReceiverHost                string                `json:"receiverHost,omitzero"`
	ReceiverSniService          string                `json:"receiverSniService,omitzero"`
	TargetGatewayAddress        string                `json:"targetGatewayAddress,omitzero"`
	NewStorageDeviceId          string                `json:"newStorageDeviceId,omitzero"`
	NewResourceName             string                `json:"newResourceName,omitzero"`
	NewVolumeAttributes         map[string]string     `json:"newVolumeAttributes,omitzero"`
	NewVolumeSize               *float64              `json:"newVolumeSize,omitzero"`
	DestinationDeviceIdToRemove string                `json:"destinationDeviceIdToRemove,omitzero"`
	TargetHasWorkloadPod        *bool                 `json:"targetHasWorkloadPod,omitzero"`
	SourceDeviceId              string                `json:"sourceDeviceId,omitzero"`
	SourceInUse                 *bool                 `json:"sourceInUse,omitzero"`
	BulkSyncTargetPodUid        string                `json:"bulkSyncTargetPodUid,omitzero"`
	BulkSyncAttempt             *float64              `json:"bulkSyncAttempt,omitzero"`
	BulkSyncCompletedAt         string                `json:"bulkSyncCompletedAt,omitzero" validate:"date"`
	BulkSyncProgressPercent     *float64              `json:"bulkSyncProgressPercent,omitzero"`
	BulkSyncTransferredBytes    *float64              `json:"bulkSyncTransferredBytes,omitzero"`
	BulkSyncRate                string                `json:"bulkSyncRate,omitzero"`
	BulkSyncEta                 string                `json:"bulkSyncEta,omitzero"`
}

// TerminalStages are the lifecycle stages a command never leaves (command.ts
// TerminalCommandStages). Defined once: a hand-written copy that omitted "cancelled"
// wedged volume-set deletion on both sides.
var TerminalStages = []string{"failed", "completed", "cancelled"}

type ShrinkVolumeStatusStage string

const (
	ShrinkVolumeStatusStageCreateVolume            ShrinkVolumeStatusStage = "create-volume"
	ShrinkVolumeStatusStageApplyPresyncResources   ShrinkVolumeStatusStage = "apply-presync-resources"
	ShrinkVolumeStatusStageInjectAndAwaitPresync   ShrinkVolumeStatusStage = "inject-and-await-presync"
	ShrinkVolumeStatusStageApplyDeltaSource        ShrinkVolumeStatusStage = "apply-delta-source"
	ShrinkVolumeStatusStageCommitAndPrepareSwap    ShrinkVolumeStatusStage = "commit-and-prepare-swap"
	ShrinkVolumeStatusStageShutdownReplica         ShrinkVolumeStatusStage = "shutdown-replica"
	ShrinkVolumeStatusStageAwaitReplicaTermination ShrinkVolumeStatusStage = "await-replica-termination"
	ShrinkVolumeStatusStageAwaitDataReady          ShrinkVolumeStatusStage = "await-data-ready"
	ShrinkVolumeStatusStageCleanupK8S              ShrinkVolumeStatusStage = "cleanup-k8s"
	ShrinkVolumeStatusStageFail                    ShrinkVolumeStatusStage = "fail"
)

type ShrinkVolumeStatusClusterIdByLocation map[string]string

type ShrinkVolumeStatusOldVolumeAttributes map[string]string

type ShrinkVolumeStatusNewVolumeAttributes map[string]string

type ShrinkVolumeStatus struct {
	Stage                   ShrinkVolumeStatusStage               `json:"stage,omitzero" validate:"required,default=create-volume,valid=create-volume|apply-presync-resources|inject-and-await-presync|apply-delta-source|commit-and-prepare-swap|shutdown-replica|await-replica-termination|await-data-ready|cleanup-k8s|fail"`
	Messages                []string                              `json:"messages,omitzero"`
	ClusterId               string                                `json:"clusterId,omitzero"`
	ClusterIdByLocation     ShrinkVolumeStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	InUseByWorkloadId       string                                `json:"inUseByWorkloadId,omitzero"`
	StorageDeviceIdToRemove string                                `json:"storageDeviceIdToRemove,omitzero"`
	OldVolumeAttributes     ShrinkVolumeStatusOldVolumeAttributes `json:"oldVolumeAttributes,omitzero"`
	OldVolumeSize           *float64                              `json:"oldVolumeSize,omitzero"`
	OldDiskType             volumeSet.DiskType                    `json:"oldDiskType,omitzero"`
	NewStorageDeviceId      string                                `json:"newStorageDeviceId,omitzero"`
	NewVolumeAttributes     ShrinkVolumeStatusNewVolumeAttributes `json:"newVolumeAttributes,omitzero"`
	NewResourceName         string                                `json:"newResourceName,omitzero"`
	NextVolumeSize          *float64                              `json:"nextVolumeSize,omitzero"`
	TempStorageClassName    string                                `json:"tempStorageClassName,omitzero"`
	LockNames               []string                              `json:"lockNames,omitzero"`
	PresyncTargetPodUid     string                                `json:"presyncTargetPodUid,omitzero"`
	PresyncAttempt          *float64                              `json:"presyncAttempt,omitzero"`
	PresyncProgressPercent  *float64                              `json:"presyncProgressPercent,omitzero"`
	FinalSyncCompletedAt    string                                `json:"finalSyncCompletedAt,omitzero" validate:"date"`
	AwaitDataReadySince     string                                `json:"awaitDataReadySince,omitzero" validate:"date"`
}

type SnapshotDeletionStatusStage string

const (
	SnapshotDeletionStatusStagePending             SnapshotDeletionStatusStage = "pending"
	SnapshotDeletionStatusStageK8SResourcesCreated SnapshotDeletionStatusStage = "k8s-resources-created"
	SnapshotDeletionStatusStageDeleted             SnapshotDeletionStatusStage = "deleted"
)

type SnapshotDeletionStatus struct {
	Stage    SnapshotDeletionStatusStage `json:"stage,omitzero" validate:"default=pending,valid=pending|k8s-resources-created|deleted"`
	Messages []string                    `json:"messages,omitzero"`
}

type StopReplicaSpec struct {
	Location string `json:"location" validate:"required"`
	Replica  string `json:"replica,omitzero"`
}

type StopReplicaStatusClusterIdByLocation map[string]string

type StopReplicaStatus struct {
	ClusterId               string                               `json:"clusterId,omitzero"`
	ClusterIdByLocation     StopReplicaStatusClusterIdByLocation `json:"clusterIdByLocation,omitzero"`
	PodCreatedByCommandLink string                               `json:"podCreatedByCommandLink,omitzero"`
	PodId                   string                               `json:"podId" validate:"required"`
}
