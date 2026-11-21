/* hand-maintained — canonical source of truth (nodelibs/schema/src/cronjob.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go.
 * Object customs (dynamic date defaults on JobExecutionCondition; delete-null
 * completionTime on JobExecutionStatus) are RegisterRefine'd in pkg/handlers. */

package cronjob

import "github.com/controlplane-com/libs-go/pkg/schema/containerstatus"

type JobExecutionCondition struct {
	Status             string `json:"status,omitzero" validate:"allow=,default=Unknown"`
	Type               string `json:"type" validate:"allow=,required"`
	LastDetectionTime  string `json:"lastDetectionTime,omitzero" validate:"date"`
	LastTransitionTime string `json:"lastTransitionTime,omitzero" validate:"date"`
	Message            string `json:"message,omitzero"`
	Reason             string `json:"reason,omitzero"`
}

type JobExecutionStatusStatus string

const (
	JobExecutionStatusStatusSuccessful JobExecutionStatusStatus = "successful"
	JobExecutionStatusStatusFailed     JobExecutionStatusStatus = "failed"
	JobExecutionStatusStatusActive     JobExecutionStatusStatus = "active"
	JobExecutionStatusStatusPending    JobExecutionStatusStatus = "pending"
	JobExecutionStatusStatusInvalid    JobExecutionStatusStatus = "invalid"
	JobExecutionStatusStatusRemoved    JobExecutionStatusStatus = "removed"
	JobExecutionStatusStatusEmpty      JobExecutionStatusStatus = ""
)

// EnumValues resolves JobExecutionStatusStatus to Joi.string().valid(...) (incl. ”).
func (JobExecutionStatusStatus) EnumValues() []string {
	return []string{"successful", "failed", "active", "pending", "invalid", "removed", ""}
}

type JobExecutionStatusContainers map[string]containerstatus.ContainerStatus

type JobExecutionStatus struct {
	WorkloadVersion float64                      `json:"workloadVersion" validate:"required"`
	Status          JobExecutionStatusStatus     `json:"status,omitzero" validate:"allow=,default=pending"`
	StartTime       string                       `json:"startTime,omitzero" validate:"date"`
	CompletionTime  string                       `json:"completionTime,omitzero" validate:"date,allow=null"`
	Conditions      []JobExecutionCondition      `json:"conditions,omitzero"`
	Name            string                       `json:"name" validate:"required"`
	Replica         string                       `json:"replica,omitzero"`
	Containers      JobExecutionStatusContainers `json:"containers,omitzero"`
	Message         string                       `json:"message,omitzero"`
}
