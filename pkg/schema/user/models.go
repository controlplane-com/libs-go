/* hand-maintained — canonical source of truth for the `user` resource schema.
 * Was joi-to-go-generated from nodelibs/schema/src/user.ts. User is a standalone
 * Joi.object (a base-field subset, NOT Base.keys). `validate:` tags carry the rules.
 * See user.ts.
 * InviteRequest/InviteResponse keep their hand-written pkg/handlers rules.
 */

package user

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type InviteRequest struct {
	Emails []string `json:"emails,omitzero"`
}

type InviteResponseErrors map[string]base.ApiError

type InviteResponseInvitations map[string]string

type InviteResponse struct {
	Errors      InviteResponseErrors      `json:"errors,omitzero"`
	Invitations InviteResponseInvitations `json:"invitations,omitzero"`
}

type UserKind string

const (
	UserKindUser UserKind = "user"
)

func (UserKind) EnumValues() []string { return []string{string(UserKindUser)} }

// UserTags aliases base.Tags for the reflection base-type registry.
type UserTags = base.Tags

type User struct {
	Id           string     `json:"id,omitzero"`
	Name         string     `json:"name,omitzero"`
	Kind         UserKind   `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive"`
	Tags         UserTags   `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date"`
	LastModified string     `json:"lastModified,omitzero" validate:"date"`
	Links        base.Links `json:"links,omitzero"`
	Idp          string     `json:"idp,omitzero"`
	Email        string     `json:"email,omitzero" validate:"email,trim"`
}
