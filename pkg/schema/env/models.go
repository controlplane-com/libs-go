/* auto-generated */

package env

type EnvCollection []EnvVar

type EnvFromCollection []EnvFromSource

type EnvFromSourceSecretRef struct {
	Name string `json:"name"`
}

type EnvFromSource struct {
	SecretRef EnvFromSourceSecretRef `json:"secretRef"`
}

type EnvName string

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}
