/* hand-maintained — canonical source of truth for the `userSettings` document schema.
 * Ported from nodelibs/schema/src/userSettings.ts, replacing the joi-to-go emission.
 * `validate:` tags carry what Reflect expresses; the open scopes and the org-name key
 * pattern live in validate.go. See go-libs/schema/ipSet/models.go for the tag DSL.
 */

package userSettings

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type UserSettingsKind string

const (
	UserSettingsKindUserSettings UserSettingsKind = "userSettings"
)

func (UserSettingsKind) EnumValues() []string { return []string{string(UserSettingsKindUserSettings)} }

// UserSettingsGlobal and UserSettingsOrg are open objects: their content is the client's.
type UserSettingsGlobal map[string]any

type UserSettingsOrg map[string]any

// UserSettingsOrgs is keyed by org name.
type UserSettingsOrgs map[string]UserSettingsOrg

type UserSettings struct {
	Kind         UserSettingsKind   `json:"kind,omitzero"`
	Version      *float64           `json:"version,omitzero" validate:"integer,min=0"`
	Created      string             `json:"created,omitzero" validate:"date"`
	LastModified string             `json:"lastModified,omitzero" validate:"date"`
	Links        base.Links         `json:"links,omitzero"`
	Global       UserSettingsGlobal `json:"global,omitzero"`
	Orgs         UserSettingsOrgs   `json:"orgs,omitzero"`
}
