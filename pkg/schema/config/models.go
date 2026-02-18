/* auto-generated */

package config

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ConfigKind string

const (
	ConfigKindConfig ConfigKind = "config"
)

type Config[D any] struct {
	Kind         ConfigKind `json:"kind,omitempty"`
	Id           string     `json:"id,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Version      float32    `json:"version"`
	Data         D          `json:"data,omitempty"`
	Links        base.Links `json:"links,omitempty"`
}
