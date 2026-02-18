/* auto-generated */

package secret

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AwsKey struct {
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	RoleArn    string `json:"roleArn,omitempty"`
	ExternalId string `json:"externalId,omitempty"`
}

type AzureConnector struct {
	Url  string `json:"url"`
	Code string `json:"code"`
}

type Dictionary map[string]string

type EcrPull struct {
	AccessKey  string   `json:"accessKey"`
	SecretKey  string   `json:"secretKey"`
	RoleArn    string   `json:"roleArn,omitempty"`
	ExternalId string   `json:"externalId,omitempty"`
	Repos      []string `json:"repos,omitempty"`
}

type KeyPair struct {
	SecretKey  string `json:"secretKey"`
	PublicKey  string `json:"publicKey,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
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
	Payload  any/* TODO: [object Object]*/ `json:"payload,omitempty"`
	Encoding OpaqueEncoding `json:"encoding,omitempty"`
}

type SecretTags map[string]any

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

type Secret struct {
	Id           string     `json:"id,omitempty"`
	Name         base.Name  `json:"name,omitempty"`
	Kind         base.Kind  `json:"kind,omitempty"`
	Version      float32    `json:"version"`
	Description  string     `json:"description,omitempty"`
	Tags         SecretTags `json:"tags,omitempty"`
	Created      string     `json:"created,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Links        base.Links `json:"links,omitempty"`
	Type         SecretType `json:"type,omitempty"`
	Data         SecretData `json:"data,omitempty"`
}

type SecretData any /* TODO: [object Object]*/

type Tls struct {
	Key   string `json:"key,omitempty"`
	Cert  string `json:"cert"`
	Chain string `json:"chain,omitempty"`
}

type UsernamePasswordEncoding string

const (
	UsernamePasswordEncodingPlain  UsernamePasswordEncoding = "plain"
	UsernamePasswordEncodingBase64 UsernamePasswordEncoding = "base64"
)

type UsernamePassword struct {
	Username string                   `json:"username"`
	Password string                   `json:"password"`
	Encoding UsernamePasswordEncoding `json:"encoding,omitempty"`
}
