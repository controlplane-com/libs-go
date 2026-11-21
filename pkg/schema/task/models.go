/* hand-maintained — canonical source of truth for the `task` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/task.ts.
 * NOT Base-derived (plain object). Registered as the `task` kind.
 */

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
	Id           string       `json:"id,omitzero"`
	Kind         TaskKind     `json:"kind,omitzero" validate:"valid=task"`
	Description  string       `json:"description,omitzero" validate:"trim"`
	Created      string       `json:"created,omitzero" validate:"date"`
	LastModified string       `json:"lastModified,omitzero" validate:"date"`
	TargetEmail  string       `json:"targetEmail,omitzero" validate:"email"`
	CreatorLink  string       `json:"creatorLink,omitzero"`
	Type         string       `json:"type,omitzero"`
	Context      TaskContext  `json:"context,omitzero"`
	Status       TaskStatus   `json:"status,omitzero" validate:"valid=pending|complete|canceled"`
	Response     TaskResponse `json:"response,omitzero"`
	Links        base.Links   `json:"links,omitzero" validate:"optional"`
}
