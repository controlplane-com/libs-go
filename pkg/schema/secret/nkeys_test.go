package secret

import (
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestNatsAccountErrorPrefix pins the nkeys failure text: TS interpolates the thrown JS
// error, whose toString() carries the class name ("NKeysError: nkeys: invalid checksum").
func TestNatsAccountErrorPrefix(t *testing.T) {
	r := NatsAccount{}.ValidationRule()
	obj := map[string]any{
		"accountId":  "A" + strings.Repeat("B", 55),
		"privateKey": "SA" + strings.Repeat("B", 54),
	}
	_, err := r.ValidateAt([]string{"data"}, obj, nil, validate.Prefs{})
	if err == nil {
		t.Fatal("a bogus nkeys account accepted")
	}
	if !strings.Contains(err.Error(), "NKeysError: ") {
		t.Errorf("message = %q, want an NKeysError: prefix", err.Error())
	}
}
