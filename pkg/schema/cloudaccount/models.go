/* auto-generated */

package cloudaccount

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AwsConfig struct {
	RoleArn string `json:"roleArn"`
}

type AzureConfig struct {
	SecretLink string `json:"secretLink"`
}

type CloudAccountTags map[string]any

type CloudAccountStatus struct {
	Usable      bool   `json:"usable,omitempty"`
	LastChecked string `json:"lastChecked,omitempty"`
	LastError   string `json:"lastError,omitempty"`
}

type CloudAccount struct {
	Id           string             `json:"id,omitempty"`
	Name         base.Name          `json:"name,omitempty"`
	Kind         base.Kind          `json:"kind,omitempty"`
	Version      *float32           `json:"version,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tags         CloudAccountTags   `json:"tags,omitempty"`
	Created      string             `json:"created,omitempty"`
	LastModified string             `json:"lastModified,omitempty"`
	Links        base.Links         `json:"links,omitempty"`
	Provider     base.CloudProvider `json:"provider,omitempty"`
	Data         any/* TODO: [object Object]*/ `json:"data,omitempty"`
	Status       CloudAccountStatus `json:"status,omitempty"`
}

type GcpConfig struct {
	ProjectId string `json:"projectId"`
}

type InstructionsData map[string]any

type Instructions struct {
	Message string           `json:"message,omitempty"`
	Data    InstructionsData `json:"data,omitempty"`
}

type NgsConfig struct {
	SecretLink string `json:"secretLink"`
}
