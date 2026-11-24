/* hand-maintained — canonical source of truth for the `cloudaccount` resource
 * schema. Formerly emitted by joi-to-go from nodelibs/schema/src/cloudaccount.ts;
 * now the CANONICAL definition carrying `validate:` struct tags that pkg/validate.FromType
 * reflects into the *validate.Rule tree (replacing pkg/handlers/schema_cloudaccount.go).
 *
 * `data` is a Joi.alternatives().conditional('provider', …): the provider scalar
 * selects the credential body. A field tag cannot carry a cross-field conditional,
 * so `Data` is Any here and the conditional is expressed by the object-level
 * .Refine layered in CloudAccount.ValidationRule (validate.go). See
 * go-libs/schema/ipSet/models.go for the tag DSL grammar.
 */

package cloudaccount

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// AwsRoleArn is an AWS IAM role ARN — the strongly-typed replacement for the
// `pattern=cloudAccountAwsRoleArn` tag (Joi.string().regex(awsRoleArnRe)).
type AwsRoleArn string

// Suffix of an AWS bridge role, cpln-<org>-<suffix>; validates the instructions' ?suffix=.
type Suffix string

// AwsPolicyArn is an IAM policy ARN, customer or AWS managed.
type AwsPolicyArn string

// GcpProjectId is a GCP project id — the strongly-typed replacement for the
// `pattern=gcpProjectId` tag (Joi.string().regex(gcpProjectIDRe)).
type GcpProjectId string

// SecretLink is an org-local link constrained to a "secret" resource — the
// strongly-typed replacement for the `localLink=secret` tag on the azure/ngs
// credential body (Joi LocalLink.custom(toKind('secret')), cloudaccount.ts).
// It expands a `//secret/foo` link to org-absolute and kind-checks it.
type SecretLink string

type AwsConfig struct {
	RoleArn           AwsRoleArn   `json:"roleArn" validate:"required"`
	BoundaryPolicyArn AwsPolicyArn `json:"boundaryPolicyArn,omitzero"`
}

type AzureConfig struct {
	SecretLink SecretLink `json:"secretLink" validate:"required"`
}

// CloudAccountTags aliases base.Tags so the reflection base-type registry resolves
// it to the shared TagsRule.
type CloudAccountTags = base.Tags

type CloudAccountStatus struct {
	Usable      *bool  `json:"usable,omitzero"`
	LastChecked string `json:"lastChecked,omitzero" validate:"date,iso"`
	LastError   string `json:"lastError,omitzero"`
}

type CloudAccount struct {
	Id           string              `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name           `json:"name,omitzero"`
	Kind         base.Kind           `json:"kind,omitzero"`
	Version      *float64            `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string              `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         CloudAccountTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string              `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string              `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links          `json:"links,omitzero" validate:"readOnly"`
	Provider     base.CloudProvider  `json:"provider,omitzero"`
	Data         any                 `json:"data,omitzero" validate:"unmodifiable" ts:"GcpConfig | AwsConfig | AzureConfig | NgsConfig"`
	Status       *CloudAccountStatus `json:"status,omitzero"`
}

type GcpConfig struct {
	ProjectId GcpProjectId `json:"projectId" validate:"required,min=6,max=30"`
}

type InstructionsData map[string]any

type Instructions struct {
	Message string           `json:"message,omitzero"`
	Data    InstructionsData `json:"data,omitzero"`
}

type NgsConfig struct {
	SecretLink SecretLink `json:"secretLink" validate:"required"`
}
