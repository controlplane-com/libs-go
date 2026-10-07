/* auto-generated */

package userSettings

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type UserSettingsKind string

const (
	UserSettingsKindUserSettings UserSettingsKind = "userSettings"
)

type UserSettingsOrgs map[string]UserSettingsOrg

type UserSettings struct {
	Kind         UserSettingsKind   `json:"kind,omitempty"`
	Version      *float32           `json:"version,omitempty"`
	Created      string             `json:"created,omitempty"`
	LastModified string             `json:"lastModified,omitempty"`
	Links        base.Links         `json:"links,omitempty"`
	Global       UserSettingsGlobal `json:"global,omitempty"`
	Orgs         UserSettingsOrgs   `json:"orgs,omitempty"`
}

type UserSettingsGlobal map[string]any

type UserSettingsOrg map[string]any
