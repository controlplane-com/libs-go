/* auto-generated */

package env

type EnvCollection []EnvVar

type EnvName string

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}
