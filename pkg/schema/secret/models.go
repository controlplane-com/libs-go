/* hand-maintained — canonical source of truth for the `secret` resource schema.
 *
 * `validate:` struct tags drive pkg/validate reflection. See
 * nodelibs/schema/src/secret.ts and the tag DSL on go-libs/schema/ipSet/models.go.
 *
 * The top-level Secret reflects Base.keys + type + data; the `data` value is a
 * type-selected union (secret.ts:155 alternatives().conditional('type', …)) applied
 * by the secretDataRefine object-rule wired in validate.go, which dispatches to the
 * per-type sub-rule ValidationRule methods (Opaque, AwsKey, EcrPull,
 * UsernamePassword, AzureConnector, Tls, KeyPair, Dictionary, NatsAccount).
 */

package secret

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AwsKey struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	RoleArn    string `json:"roleArn,omitzero"`
	ExternalId string `json:"externalId,omitzero"`
}

type AzureConnector struct {
	Url  string `json:"url"`
	Code string `json:"code"`
}

type Dictionary map[string]string

type EcrPull struct {
	AccessKey  string   `json:"accessKey"`
	SecretKey  string   `json:"secretKey"`
	RoleArn    string   `json:"roleArn,omitzero"`
	ExternalId string   `json:"externalId,omitzero"`
	Repos      []string `json:"repos,omitzero"`
}

type KeyPair struct {
	SecretKey  string `json:"secretKey"`
	PublicKey  string `json:"publicKey,omitzero"`
	Passphrase string `json:"passphrase,omitzero"`
}

type NatsAccount struct {
	AccountId  string `json:"accountId"`
	PrivateKey string `json:"privateKey"`
}

type OpaqueEncoding string

const (
	OpaqueEncodingPlain  OpaqueEncoding = "plain"
	OpaqueEncodingBase64 OpaqueEncoding = "base64"
)

type Opaque struct {
	// Payload is validated by OpaqueRule as alternatives().conditional('encoding')
	// (secret.ts:6): base64 string when encoding=base64, else any string.
	Payload  any            `json:"payload,omitzero" ts:"string"`
	Encoding OpaqueEncoding `json:"encoding,omitzero"`
}

// SecretTags aliases base.Tags so the reflection base-type registry resolves it
// to the shared TagsRule.
type SecretTags = base.Tags

type SecretType string

const (
	SecretTypeOpaque         SecretType = "opaque"
	SecretTypeTls            SecretType = "tls"
	SecretTypeGcp            SecretType = "gcp"
	SecretTypeAws            SecretType = "aws"
	SecretTypeEcr            SecretType = "ecr"
	SecretTypeUserpass       SecretType = "userpass"
	SecretTypeKeypair        SecretType = "keypair"
	SecretTypeAzureSdk       SecretType = "azure-sdk"
	SecretTypeAzureConnector SecretType = "azure-connector"
	SecretTypeDocker         SecretType = "docker"
	SecretTypeDictionary     SecretType = "dictionary"
	SecretTypeNatsAccount    SecretType = "nats-account"
)

// EnumValues ports `type: Joi.string().valid(...)` (secret.ts:171).
func (SecretType) EnumValues() []string {
	return []string{
		string(SecretTypeOpaque), string(SecretTypeTls), string(SecretTypeGcp),
		string(SecretTypeAws), string(SecretTypeEcr), string(SecretTypeUserpass),
		string(SecretTypeKeypair), string(SecretTypeAzureSdk), string(SecretTypeAzureConnector),
		string(SecretTypeDocker), string(SecretTypeDictionary), string(SecretTypeNatsAccount),
	}
}

type Secret struct {
	Id           string     `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name  `json:"name,omitzero"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string     `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         SecretTags `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string     `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links `json:"links,omitzero" validate:"readOnly"`
	Type         SecretType `json:"type,omitzero"`
	// Data is validated by type via a RegisterRefine on Secret (secret.ts:155
	// alternatives().conditional('type', …)) — see schema_reflect_iam.go.
	Data SecretData `json:"data,omitzero"`
}

type SecretData any //ts:AwsKey | AzureConnector | Dictionary | EcrPull | KeyPair | NatsAccount | Opaque | Tls | UsernamePassword | string

type Tls struct {
	Key   string `json:"key,omitzero"`
	Cert  string `json:"cert"`
	Chain string `json:"chain,omitzero"`
}

type UsernamePasswordEncoding string

const (
	UsernamePasswordEncodingPlain  UsernamePasswordEncoding = "plain"
	UsernamePasswordEncodingBase64 UsernamePasswordEncoding = "base64"
)

type UsernamePassword struct {
	Username string                   `json:"username"`
	Password string                   `json:"password"`
	Encoding UsernamePasswordEncoding `json:"encoding,omitzero"`
}
