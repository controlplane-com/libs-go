package vm

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// check decodes body into T and validates it through the TYPED walk (these rules
// carry RefineT customs, which the generic evaluator refuses).
func check[T any](t *testing.T, rule *validate.Rule, body string) (string, string) {
	t.Helper()
	var v T
	d, err := validate.DecodeInto(&v, validate.Bytes(body))
	if err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	verr := d.Validate(rule, &validate.Context{Org: "o", Gvc: "g", Method: "POST"}, validate.Prefs{})
	out, _ := json.Marshal(v)
	if verr != nil {
		return verr.Message, string(out)
	}
	return "", string(out)
}

func assertErr(t *testing.T, got, want string) {
	t.Helper()
	switch {
	case want == "" && got != "":
		t.Fatalf("unexpected error: %s", got)
	case want != "" && !strings.Contains(got, want):
		t.Fatalf("got %q want it to contain %q", got, want)
	}
}

// TestBootSourceXor pins VmBootSource's `.xor('oci','http')` (vm.ts:41). Without it
// a boot disk with both sources — or with none — was accepted and stored.
func TestBootSourceXor(t *testing.T) {
	rule := VmSpecBootDiskSource{}.ValidationRule()
	for _, c := range []struct{ name, body, wantErr string }{
		{"both", `{"oci":{"image":"quay.io/x/y:1"},"http":{"url":"https://a/b"}}`, "contains a conflict between exclusive peers [oci, http]"},
		{"neither", `{}`, "must contain at least one of [oci, http]"},
		{"oci only", `{"oci":{"image":"quay.io/x/y:1"}}`, ""},
		{"http only", `{"http":{"url":"https://a/b"}}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecBootDiskSource](t, rule, c.body)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestCloudInitXor pins `.xor('userData','userDataBase64','userDataSecret')` (vm.ts:137).
func TestCloudInitXor(t *testing.T) {
	rule := VmSpecCloudInit{}.ValidationRule()
	for _, c := range []struct{ name, body, wantErr string }{
		{"two forms", `{"userData":"#cloud-config","userDataBase64":"aGk="}`, "contains a conflict between exclusive peers [userData, userDataBase64, userDataSecret]"},
		{"none", `{}`, "must contain at least one of [userData, userDataBase64, userDataSecret]"},
		{"inline only", `{"userData":"#cloud-config"}`, ""},
		{"secret only", `{"userDataSecret":"//secret/s1"}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecCloudInit](t, rule, c.body)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestSecretLink pins `LocalLink.custom(toKind('secret'))` on every secret reference
// under spec.vm (vm.ts:124,132,139) — they were plain unvalidated strings.
func TestSecretLink(t *testing.T) {
	rule := VmSpecCloudInit{}.ValidationRule()
	got, _ := check[VmSpecCloudInit](t, rule, `{"userDataSecret":"not-a-link"}`)
	assertErr(t, got, "fails to match the required pattern")
	got, _ = check[VmSpecCloudInit](t, rule, `{"userData":"#c","sshPublicKeySecrets":["nope"]}`)
	assertErr(t, got, "fails to match the required pattern")

	acRule := validate.Reflect(VmSpecAccessCredentials{})
	got, _ = check[VmSpecAccessCredentials](t, acRule, `{"sshPublicKeySecret":"nope","users":["root"]}`)
	assertErr(t, got, "fails to match the required pattern")
	got, _ = check[VmSpecAccessCredentials](t, acRule, `{"sshPublicKeySecret":"//secret/s1","users":["root"]}`)
	assertErr(t, got, "")
}

// TestVmUser pins `max(32)` + `/^[a-z_][a-z0-9_-]*$/` on accessCredentials[].users[]
// (vm.ts:141-147) — arbitrary text was being injected into the guest.
func TestVmUser(t *testing.T) {
	rule := validate.Reflect(VmSpecAccessCredentials{})
	for _, c := range []struct{ name, user, wantErr string }{
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "less than or equal to 32"},
		{"uppercase", "Root", "fails to match the required pattern"},
		{"space", "bad user", "fails to match the required pattern"},
		{"leading digit", "1user", "fails to match the required pattern"},
		{"ok", "ubuntu", ""},
		{"ok underscore", "_dev-user_1", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecAccessCredentials](t, rule, `{"sshPublicKeySecret":"//secret/s1","users":["`+c.user+`"]}`)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestFirmware pins the secureBoot throw (vm.ts:90-98) and the uuidv4-only GUID
// (vm.ts:79-82). Secure Boot needs persistent EFI NVRAM the clusters do not provide,
// so accepting it stored a guest that cannot boot after a reboot.
func TestFirmware(t *testing.T) {
	rule := VmSpecFirmware{}.ValidationRule()
	for _, c := range []struct{ name, body, wantErr string }{
		{"secureBoot true", `{"secureBoot":true}`, "firmware.secureBoot is not yet supported"},
		{"secureBoot false", `{"secureBoot":false}`, ""},
		{"uuid v1", `{"uuid":"b54adc00-67f9-11d9-9669-0800200c9a66"}`, "must be a valid GUID"},
		{"uuid v4", `{"uuid":"f47ac10b-58cc-4372-a567-0e02b2c3d479"}`, ""},
		{"uuid garbage", `{"uuid":"nope"}`, "must be a valid GUID"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecFirmware](t, rule, c.body)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestVmSpecDefaults pins vm.ts:166-181's deep defaults: without them a minimal
// type=vm workload stored none of the structural defaults TS stores, so the two
// services disagreed on a spec they had both just accepted.
func TestVmSpecDefaults(t *testing.T) {
	_, out := check[VmSpec](t, VmSpec{}.ValidationRule(),
		`{"bootDisk":{"source":{"oci":{"image":"quay.io/x/y:1"}},"persist":{"volumeSet":"cpln://volumeset/vs1"}}}`)
	for _, want := range []string{
		`"firmware":{"bootloader":"efi","secureBoot":false}`,
		`"clock":{"timezone":"UTC"}`,
		`"networks":[{"name":"default"}]`,
		`"guestOS":"linux"`,
		`"runStrategy":"Always"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing default %s in %s", want, out)
		}
	}
	// An explicit network must survive untouched.
	if _, out := check[VmSpec](t, VmSpec{}.ValidationRule(),
		`{"bootDisk":{"source":{"oci":{"image":"quay.io/x/y:1"}}},"networks":[{"name":"net1"}]}`); !strings.Contains(out, `"networks":[{"name":"net1"}]`) {
		t.Fatalf("explicit network overwritten: %s", out)
	}
}

// TestBootSourceHttpUrl pins `.uri({scheme:['http','https']})` (vm.ts:24-26). The
// prefix check this replaced took any string starting with http(s):// — `https://`
// with no host, `http://a b` — and reported a message Joi never emits.
func TestBootSourceHttpUrl(t *testing.T) {
	rule := VmSpecBootDiskSource{}.ValidationRule()
	for _, c := range []struct{ name, url, wantErr string }{
		{"ftp", "ftp://e/x.img", "must be a valid uri with a scheme matching the http|https pattern"},
		{"not a uri", "nope", "must be a valid uri with a scheme matching the http|https pattern"},
		{"uppercase scheme", "HTTP://e/x.img", "must be a valid uri with a scheme matching the http|https pattern"},
		{"empty authority", "https://", "must be a valid uri with a scheme matching the http|https pattern"},
		{"space in host", "http://a b", "must be a valid uri with a scheme matching the http|https pattern"},
		// Joi special-cases these two back to the plain string.uri even with a scheme set.
		{"bare http:/", "http:/", "must be a valid uri"},
		{"https ok", "https://example.com/x.qcow2", ""},
		{"http ok with query", "http://example.com/x.img?v=1#f", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecBootDiskSource](t, rule, `{"http":{"url":`+strconv.Quote(c.url)+`}}`)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestCloudInitBase64 pins `Joi.string().base64()` (vm.ts:119): Joi requires padding,
// so an unpadded payload is rejected — and the message names the field the way every
// other string rule does, not the hardcoded "userDataBase64".
func TestCloudInitBase64(t *testing.T) {
	rule := VmSpecCloudInit{}.ValidationRule()
	for _, c := range []struct{ name, v, wantErr string }{
		{"garbage", "!!!!", "must be a valid base64 string"},
		{"unpadded", "YWJjZA", "must be a valid base64 string"},
		{"padded", "YWJjZA==", ""},
		{"exact block", "YWJj", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[VmSpecCloudInit](t, rule, `{"userDataBase64":`+strconv.Quote(c.v)+`}`)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestVolumeSetUriMessage pins VolumeSetUri's `.message()` override (vm.ts:5-7) — the
// pattern message Joi would otherwise emit leaks the regex at the caller.
func TestVolumeSetUriMessage(t *testing.T) {
	rule := validate.Reflect(VmSpecBootDiskPersist{})
	got, _ := check[VmSpecBootDiskPersist](t, rule, `{"volumeSet":"cpln://vs/vs1"}`)
	if got != "volumeSet must be in the form cpln://volumeset/<name>" {
		t.Fatalf("got %q", got)
	}
	if got, _ := check[VmSpecBootDiskPersist](t, rule, `{"volumeSet":"cpln://volumeset/vs1"}`); got != "" {
		t.Fatalf("unexpected error: %s", got)
	}
}
