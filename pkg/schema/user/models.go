/* auto-generated */

package user

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type InviteRequest struct {
	Emails []string `json:"emails,omitempty"`
}

type InviteResponseErrors map[string]base.ApiError

type InviteResponseInvitations map[string]string

type InviteResponse struct {
	Errors      InviteResponseErrors      `json:"errors,omitempty"`
	Invitations InviteResponseInvitations `json:"invitations,omitempty"`
}

type UserKind string

const (
	UserKindUser UserKind = "user"
)

type UserTags map[string]any

type User struct {
	Id           string     `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Kind         UserKind   `json:"kind,omitempty"`
	Version      *float32   `json:"version,omitempty"`
	Tags         UserTags   `json:"tags,omitempty"`
	Created      string     `json:"created,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Links        base.Links `json:"links,omitempty"`
	Idp          string     `json:"idp,omitempty"`
	Email        string     `json:"email,omitempty"`
}
