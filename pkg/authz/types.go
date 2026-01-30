package authz

// ProfileName represents a named authorization profile
type ProfileName string

const (
	ProfileAnyValidUser ProfileName = "any-valid-user"
	ProfileRootUser     ProfileName = "root-user"
	ProfileAccountUser  ProfileName = "account-user"
	ProfileDataService  ProfileName = "data-service"
	ProfileMetering     ProfileName = "metering"
)

// User represents authenticated user information
type User struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Root  bool   `json:"root,omitempty"`
	Org   string `json:"org,omitempty"`
	Extra any    `json:"extra,omitempty"`
}

// AuthenticateRequest represents an authentication request
type AuthenticateRequest struct {
	Token   string `json:"token"`
	Profile string `json:"profile"`
	Scope   string `json:"scope,omitempty"` // e.g., org name
}

// AuthenticateResponse represents an authentication response
type AuthenticateResponse struct {
	User          *User  `json:"user,omitempty"`
	Authenticated bool   `json:"authenticated"`
	Error         string `json:"error,omitempty"`
}

// AuthorizeRequest represents an authorization request
type AuthorizeRequest struct {
	Token       string   `json:"token"`
	Profile     string   `json:"profile"`
	Permissions []string `json:"permissions,omitempty"`
	Scope       string   `json:"scope,omitempty"` // e.g., org name
}

// AuthorizeResponse represents an authorization response
type AuthorizeResponse struct {
	User       *User  `json:"user,omitempty"`
	Authorized bool   `json:"authorized"`
	Error      string `json:"error,omitempty"`
}
