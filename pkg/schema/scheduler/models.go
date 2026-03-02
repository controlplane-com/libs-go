package scheduler

import (
	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
	"github.com/controlplane-com/libs-go/pkg/schema/identity"
	"github.com/controlplane-com/libs-go/pkg/schema/org"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

type Hints struct {
	CertificateRequestors []string `json:"certificateRequestors,omitempty"`
}

type VolumeSetFlags struct {
	PendingVolumeCreationGracePeriodSeconds float32 `json:"pendingVolumeCreationGracePeriodSeconds,omitempty"`
}

type Keda struct {
	Identity *identity.Identity `json:"identity,omitempty"`
}

type ScheduleCommandEventType string

const (
	ScheduleCommandEventTypeCreate ScheduleCommandEventType = "create"
	ScheduleCommandEventTypeDelete ScheduleCommandEventType = "delete"
)

type ScheduleCommandDeleteReason string

const (
	ScheduleCommandDeleteReasonLocationDeactivated ScheduleCommandDeleteReason = "locationDeactivated"
	ScheduleCommandDeleteReasonGvcDeleted          ScheduleCommandDeleteReason = "gvcDeleted"
	ScheduleCommandDeleteReasonWorkloadDeleted     ScheduleCommandDeleteReason = "workloadDeleted"
	ScheduleCommandDeleteReasonScheduledElsewhere  ScheduleCommandDeleteReason = "scheduledElsewhere"
	ScheduleCommandDeleteReasonSuspended           ScheduleCommandDeleteReason = "suspended"
)

type ScheduleCommand struct {
	Id             string                      `json:"id"`
	EventType      ScheduleCommandEventType    `json:"eventType"`
	DeleteReason   ScheduleCommandDeleteReason `json:"deleteReason,omitempty"`
	Timestamp      string                      `json:"timestamp"`
	Gvc            *gvc.Gvc                    `json:"gvc"`
	Org            *org.Org                    `json:"org,omitempty"`
	GvcConfig      *gvc.GvcConfig              `json:"gvcConfig,omitempty"`
	Workload       *workload.Workload          `json:"workload,omitempty"`
	WorkloadConfig *workload.WorkloadConfig    `json:"workloadConfig,omitempty"`
	Identity       *identity.Identity          `json:"identity,omitempty"`
	Hints          *Hints                      `json:"hints,omitempty"`
	VolumeSetFlags *VolumeSetFlags             `json:"volumeSetFlags,omitempty"`
	Keda           *Keda                       `json:"keda,omitempty"`
}

type ScheduleServiceCommandDeleteReason string

const (
	ScheduleServiceCommandDeleteReasonLocationDeactivated ScheduleServiceCommandDeleteReason = "locationDeactivated"
	ScheduleServiceCommandDeleteReasonServiceDeleted      ScheduleServiceCommandDeleteReason = "serviceDeleted"
	ScheduleServiceCommandDeleteReasonScheduledElsewhere  ScheduleServiceCommandDeleteReason = "scheduledElsewhere"
)

type ScheduleServiceCommand struct {
	Id           string                             `json:"id"`
	EventType    ScheduleCommandEventType           `json:"eventType"`
	DeleteReason ScheduleServiceCommandDeleteReason `json:"deleteReason,omitempty"`
	Timestamp    string                             `json:"timestamp"`
	Service      any                                `json:"service"`
}
