package scheduler

import (
	"github.com/controlplane-com/libs-go/pkg/schema/gvc"
	"github.com/controlplane-com/libs-go/pkg/schema/identity"
	"github.com/controlplane-com/libs-go/pkg/schema/org"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

type Hints struct {
	CertificateRequestors []string `json:"certificateRequestors,omitzero"`
}

type Keda struct {
	Identity *identity.Identity `json:"identity,omitzero"`
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
	DeleteReason   ScheduleCommandDeleteReason `json:"deleteReason,omitzero"`
	Timestamp      string                      `json:"timestamp"`
	Gvc            *gvc.Gvc                    `json:"gvc"`
	Org            *org.Org                    `json:"org,omitzero"`
	GvcConfig      *gvc.GvcConfig              `json:"gvcConfig,omitzero"`
	Workload       *workload.Workload          `json:"workload,omitzero"`
	WorkloadConfig *workload.WorkloadConfig    `json:"workloadConfig,omitzero"`
	Identity       *identity.Identity          `json:"identity,omitzero"`
	Hints          *Hints                      `json:"hints,omitzero"`
	Keda           *Keda                       `json:"keda,omitzero"`
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
	DeleteReason ScheduleServiceCommandDeleteReason `json:"deleteReason,omitzero"`
	Timestamp    string                             `json:"timestamp"`
	Service      any                                `json:"service"`
}
