/* auto-generated */

package policy

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/query"

type Binding struct {
	Permissions    []string `json:"permissions"`
	PrincipalLinks []string `json:"principalLinks"`
}

type PolicyTags map[string]any

type PolicyTarget string

const (
	PolicyTargetAll PolicyTarget = "all"
)

type PolicyOrigin string

const (
	PolicyOriginDefault PolicyOrigin = "default"
	PolicyOriginBuiltin PolicyOrigin = "builtin"
)

type Policy struct {
	Id           string       `json:"id,omitempty"`
	Name         base.Name    `json:"name,omitempty"`
	Kind         base.Kind    `json:"kind,omitempty"`
	Version      float32      `json:"version"`
	Description  string       `json:"description,omitempty"`
	Tags         PolicyTags   `json:"tags,omitempty"`
	Created      string       `json:"created,omitempty"`
	LastModified string       `json:"lastModified,omitempty"`
	Links        base.Links   `json:"links,omitempty"`
	TargetKind   base.Kind    `json:"targetKind,omitempty"`
	TargetLinks  []string     `json:"targetLinks,omitempty"`
	TargetQuery  query.Query  `json:"targetQuery,omitempty"`
	Target       PolicyTarget `json:"target,omitempty"`
	Origin       PolicyOrigin `json:"origin,omitempty"`
	Bindings     []Binding    `json:"bindings,omitempty"`
}
