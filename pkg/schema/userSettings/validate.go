package userSettings

import (
	"github.com/controlplane-com/libs-go/pkg/schema/org"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects UserSettings, the schema the merged settings document validates against.
func (UserSettings) ValidationRule() *validate.Rule { return validate.Reflect(UserSettings{}) }

func (UserSettingsGlobal) ValidationRule() *validate.Rule { return validate.Object(nil).Unknown(true) }

func (UserSettingsOrg) ValidationRule() *validate.Rule { return validate.Object(nil).Unknown(true) }

// ValidationRule keys orgs by the org name rule (org.ts checkOrgName), so a key no org could have is refused.
func (UserSettingsOrgs) ValidationRule() *validate.Rule {
	return validate.Object(nil).Pattern(org.OrgName("").ValidationRule(), UserSettingsOrg(nil).ValidationRule())
}
