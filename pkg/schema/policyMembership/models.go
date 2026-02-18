/* auto-generated */

package policyMembership

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type PolicyMembership struct {
	Kind     base.Kind       `json:"kind,omitempty"`
	Policies []PolicySummary `json:"policies,omitempty"`
	Links    base.Links      `json:"links,omitempty"`
}

type PolicySummary struct {
	Link        string    `json:"link,omitempty"`
	Description string    `json:"description,omitempty"`
	TargetKind  base.Kind `json:"targetKind,omitempty"`
}
