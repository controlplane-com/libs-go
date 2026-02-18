/* auto-generated */

package auditctx

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AuditContextTags map[string]any

type AuditContextOrigin string

const (
	AuditContextOriginDefault AuditContextOrigin = "default"
	AuditContextOriginBuiltin AuditContextOrigin = "builtin"
)

type AuditContext struct {
	Id           string             `json:"id,omitempty"`
	Name         base.Name          `json:"name,omitempty"`
	Kind         base.Kind          `json:"kind,omitempty"`
	Version      float32            `json:"version"`
	Description  string             `json:"description,omitempty"`
	Tags         AuditContextTags   `json:"tags,omitempty"`
	Created      string             `json:"created,omitempty"`
	LastModified string             `json:"lastModified,omitempty"`
	Links        base.Links         `json:"links,omitempty"`
	Status       AuditContextStatus `json:"status,omitempty"`
	Origin       AuditContextOrigin `json:"origin,omitempty"`
}

type AuditContextStatus map[string]any
