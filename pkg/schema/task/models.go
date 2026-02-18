/* auto-generated */

package task

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type TaskKind string

const (
	TaskKindTask TaskKind = "task"
)

type TaskContext map[string]any

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusComplete TaskStatus = "complete"
	TaskStatusCanceled TaskStatus = "canceled"
)

type TaskResponse map[string]any

type Task struct {
	Id           string       `json:"id,omitempty"`
	Kind         TaskKind     `json:"kind,omitempty"`
	Description  string       `json:"description,omitempty"`
	Created      string       `json:"created,omitempty"`
	LastModified string       `json:"lastModified,omitempty"`
	TargetEmail  string       `json:"targetEmail,omitempty"`
	CreatorLink  string       `json:"creatorLink,omitempty"`
	Type         string       `json:"type,omitempty"`
	Context      TaskContext  `json:"context,omitempty"`
	Status       TaskStatus   `json:"status,omitempty"`
	Response     TaskResponse `json:"response,omitempty"`
	Links        []base.Link  `json:"links,omitempty"`
}
