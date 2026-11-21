/* auto-generated */

package repositorySync

type RepositorySyncSourceDirectory struct {
	Recurse *bool `json:"recurse,omitzero"`
}

type RepositorySyncSource struct {
	RepoURL           string                         `json:"repoURL"`
	TargetRevision    string                         `json:"targetRevision,omitzero"`
	Path              string                         `json:"path"`
	AuthSecretLink    string                         `json:"authSecretLink,omitzero"`
	WebhookSecretLink string                         `json:"webhookSecretLink"`
	Directory         *RepositorySyncSourceDirectory `json:"directory,omitzero"`
}

type RepositorySync struct {
	Source *RepositorySyncSource `json:"source,omitzero"`
}
