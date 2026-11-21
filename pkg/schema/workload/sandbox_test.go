package workload

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/sandbox"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// The sandbox subtype (sandbox.ts applySandboxSubtype, master 7f8e46b5d — dev/worker),
// mirroring sandbox.test.ts. The validation context is org "o" / gvc "g" (check()).
func sandboxWorkload(specExtra, containerExtra string) string {
	container := `{"name":"dev","image":"ubuntu:24.04"`
	if containerExtra != "" {
		container += "," + containerExtra
	}
	container += "}"
	spec := `"type":"stateful","containers":[` + container + `],"sandbox":{}`
	if specExtra != "" {
		spec += "," + specExtra
	}
	return `{"kind":"workload","name":"example","spec":{` + spec + `}}`
}

// a valid worker token (schema min 16)
const sandboxToken = "tok-0123456789abcdef"

func sandboxCheck(t *testing.T, body string) (*validate.Error, string) {
	t.Helper()
	return check[Workload](t, Workload{}.ValidationRule(), body)
}

// devContainer is the rendered dev container, for assertions on what the overlay wrote.
func devContainer(t *testing.T, out string) map[string]any {
	t.Helper()
	var w struct {
		Tags map[string]any `json:"tags"`
		Spec struct {
			Containers []map[string]any `json:"containers"`
		} `json:"spec"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	return w.Spec.Containers[0]
}

func TestSandboxOverlay(t *testing.T) {
	verr, out := sandboxCheck(t, sandboxWorkload("", ""))
	assertErr(t, verr, "")
	dev := devContainer(t, out)

	for _, want := range []string{
		`{"protocol":"http","number":8888}`,
		`{"name":"APP_PORT","value":"8080"}`,
		`"uri":"cpln://volumeset/example-state"`,
		`"path":"/home/devuser"`,
		`"cpln/sandbox":"true"`,
		`"cpln/kata":"true"`,
		`"cpln/scaleToZero":"true"`,
		`"cpln/dev-type":"devbox"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the overlay did not write %s:\n%s", want, out)
		}
	}
	// a dev box carries no credential and no privileged tag
	for _, absent := range []string{"CODE_SERVER_PASSWORD", "DEVBOXD_TOKEN", "cpln/kataPrivileged"} {
		if strings.Contains(out, absent) {
			t.Errorf("the dev overlay wrote %s:\n%s", absent, out)
		}
	}
	probe, _ := dev["readinessProbe"].(map[string]any)
	httpGet, _ := probe["httpGet"].(map[string]any)
	if httpGet["port"] != 8888.0 || httpGet["path"] != "/healthz" {
		t.Errorf("readinessProbe = %v", probe)
	}
	var w struct {
		Spec struct {
			Sandbox map[string]any `json:"sandbox"`
		} `json:"spec"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	vol, _ := w.Spec.Sandbox["volume"].(map[string]any)
	if w.Spec.Sandbox["appPort"] != 8080.0 || w.Spec.Sandbox["ide"] != "browser" ||
		w.Spec.Sandbox["type"] != "dev" || vol["initialCapacity"] != 20.0 {
		t.Errorf("section defaults = %v", w.Spec.Sandbox)
	}
}

func TestSandboxWorkerOverlay(t *testing.T) {
	verr, out := sandboxCheck(t, sandboxWorkload(`"sandbox":{"type":"worker","token":"`+sandboxToken+`"}`, ""))
	assertErr(t, verr, "")
	dev := devContainer(t, out)

	for _, want := range []string{
		`{"protocol":"http","number":49983}`,
		`{"name":"APP_PORT","value":"8080"}`,
		`{"name":"DEVBOXD_TOKEN","value":"` + sandboxToken + `"}`,
		`"cpln/kataPrivileged":"true"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the worker overlay did not write %s:\n%s", want, out)
		}
	}
	// a worker has no IDE and is not stamped cpln/dev-type
	if strings.Contains(out, "cpln/dev-type") {
		t.Errorf("a worker was stamped cpln/dev-type:\n%s", out)
	}
	var w struct {
		Spec struct {
			Sandbox map[string]any `json:"sandbox"`
		} `json:"spec"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	if _, hasIde := w.Spec.Sandbox["ide"]; hasIde {
		t.Errorf("worker kept an ide: %v", w.Spec.Sandbox)
	}
	probe, _ := dev["readinessProbe"].(map[string]any)
	httpGet, _ := probe["httpGet"].(map[string]any)
	if httpGet["port"] != 49983.0 || httpGet["path"] != "/_health" {
		t.Errorf("worker readinessProbe = %v", probe)
	}
}

// The overlay must be a fixpoint: re-validating a stored spec changes nothing, or every
// read-modify-write of a sandbox would rewrite the row.
func TestSandboxOverlayIsAFixpoint(t *testing.T) {
	for _, body := range []string{
		sandboxWorkload("", ""),
		sandboxWorkload("", `"env":[{"name":"ZED","value":"z"},{"name":"AAA","value":"a"}]`),
		sandboxWorkload(`"sandbox":{"type":"worker","token":"`+sandboxToken+`"}`, ""),
		sandboxWorkload(`"sandbox":{"scaleToZeroDelay":120},"defaultOptions":{"autoscaling":{"minScale":1,"maxScale":1}}`, ""),
	} {
		verr, first := sandboxCheck(t, body)
		assertErr(t, verr, "")
		verr, second := sandboxCheck(t, first)
		assertErr(t, verr, "")
		if first != second {
			t.Errorf("not a fixpoint:\n first: %s\nsecond: %s", first, second)
		}
		names := envNames(t, first)
		for i := 1; i < len(names); i++ {
			if names[i-1] > names[i] {
				t.Errorf("env is not at the EnvCollection fixpoint: %v", names)
			}
		}
	}
}

func envNames(t *testing.T, out string) []string {
	t.Helper()
	var names []string
	env, _ := devContainer(t, out)["env"].([]any)
	for _, e := range env {
		m, _ := e.(map[string]any)
		s, _ := m["name"].(string)
		names = append(names, s)
	}
	return names
}

func editSpec(t *testing.T, out string, edit func(spec map[string]any)) string {
	t.Helper()
	var w map[string]any
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	spec, _ := w["spec"].(map[string]any)
	edit(spec)
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func editSection(t *testing.T, out string, edit func(section map[string]any)) string {
	t.Helper()
	return editSpec(t, out, func(spec map[string]any) {
		section, _ := spec["sandbox"].(map[string]any)
		edit(section)
	})
}

func TestSandboxOverlayRejections(t *testing.T) {
	for _, c := range []struct {
		name, body, want string
	}{
		{"a non-stateful workload cannot carry a sandbox section",
			sandboxWorkload(`"type":"standard"`, ""), "requires spec.type 'stateful'"},
		{"a reserved port on any container is rejected",
			sandboxWorkload("", `"ports":[{"protocol":"http","number":8443}]`), "reserved by the sandbox runtime"},
		{"a reserved env name in a non-dev container is rejected",
			`{"kind":"workload","name":"example","spec":{"type":"stateful","sandbox":{},"containers":[` +
				`{"name":"dev","image":"ubuntu:24.04"},` +
				`{"name":"side","image":"nginx:1","env":[{"name":"APP_PORT","value":"9"}]}]}}`,
			"reserved by the sandbox runtime"},
		{"the dev container cannot use the singular port field",
			sandboxWorkload("", `"port":3000`), "cannot coexist"},
		{"a reserved appPort is rejected",
			sandboxWorkload(`"sandbox":{"appPort":8888}`, ""), "reserved by the sandbox runtime"},
		{"a worker without a token is rejected",
			sandboxWorkload(`"sandbox":{"type":"worker"}`, ""), "token is required"},
		{"a worker on the standard dev image is rejected",
			sandboxWorkload(`"sandbox":{"type":"worker","token":"`+sandboxToken+`"}`, `"image":"`+sandbox.SandboxStandardImage+`"`),
			"runs the dev type only"},
		{"a DEVBOXD_TOKEN env that disagrees with the token is rejected",
			sandboxWorkload(`"sandbox":{"type":"worker","token":"`+sandboxToken+`"}`, `"env":[{"name":"DEVBOXD_TOKEN","value":"other"}]`),
			"derived from spec.sandbox.token"},
		{"an APP_PORT other than the derived one is refused, naming spec.sandbox.appPort",
			sandboxWorkload("", `"env":[{"name":"APP_PORT","value":"9999"}]`),
			"derived from spec.sandbox.appPort"},
		{"a DEVBOXD_TOKEN on a dev sandbox is refused: a dev box carries none",
			sandboxWorkload("", `"env":[{"name":"DEVBOXD_TOKEN","value":"x"}]`),
			"a dev sandbox carries none"},
	} {
		verr, _ := sandboxCheck(t, c.body)
		if verr == nil {
			t.Errorf("%s: accepted", c.name)
			continue
		}
		if c.want != "" && !strings.Contains(verr.Message, c.want) {
			t.Errorf("%s: %q does not contain %q", c.name, verr.Message, c.want)
		}
	}
}

func TestSandboxOverlayOwnsTheReservedEntries(t *testing.T) {
	// a derived entry equal to its section value round-trips as a single canonical entry
	// (a stale one is refused by the overlay; the handler re-derives it before this point)
	verr, out := sandboxCheck(t, sandboxWorkload("", `"env":[{"name":"APP_PORT","value":"8080"}]`))
	assertErr(t, verr, "")
	if !strings.Contains(out, `{"name":"APP_PORT","value":"8080"}`) || strings.Count(out, `"name":"APP_PORT"`) != 1 {
		t.Errorf("consistent APP_PORT was not canonicalized to one entry: %s", out)
	}

	// a cloned spec's foreign mount at the home path is rewritten to this sandbox's
	verr, out = sandboxCheck(t, sandboxWorkload("",
		`"volumes":[{"uri":"cpln://volumeset/other-state","path":"/home/devuser"},{"uri":"cpln://volumeset/extra","path":"/mnt/extra"}]`))
	assertErr(t, verr, "")
	volumes, _ := devContainer(t, out)["volumes"].([]any)
	homes := 0
	for _, v := range volumes {
		m, _ := v.(map[string]any)
		if m["path"] == "/home/devuser" {
			homes++
			if m["uri"] != "cpln://volumeset/example-state" {
				t.Errorf("home volume uri = %v", m["uri"])
			}
		}
	}
	if homes != 1 || len(volumes) != 2 {
		t.Errorf("volumes = %v", volumes)
	}

	// a user readiness probe is preserved
	verr, out = sandboxCheck(t, sandboxWorkload("", `"readinessProbe":{"httpGet":{"path":"/live","port":3000},"periodSeconds":10}`))
	assertErr(t, verr, "")
	probe, _ := devContainer(t, out)["readinessProbe"].(map[string]any)
	httpGet, _ := probe["httpGet"].(map[string]any)
	if httpGet["path"] != "/live" {
		t.Errorf("readinessProbe was overwritten: %v", probe)
	}
}

func TestSandboxMarkerTagTracksTheSection(t *testing.T) {
	// a sandbox is always stamped marker + scaleToZero; the delay pins when autoscaling exists
	verr, out := sandboxCheck(t, sandboxWorkload(`"sandbox":{"scaleToZeroDelay":120},"defaultOptions":{"autoscaling":{"minScale":1,"maxScale":1}}`, ""))
	assertErr(t, verr, "")
	if !strings.Contains(out, `"cpln/scaleToZero":"true"`) || !strings.Contains(out, `"scaleToZeroDelay":120`) {
		t.Errorf("scaleToZeroDelay did not stamp the tag and pin the delay: %s", out)
	}

	// dropping the section drops the marker
	verr, out = sandboxCheck(t, editSpec(t, out, func(spec map[string]any) { delete(spec, "sandbox") }))
	assertErr(t, verr, "")
	if strings.Contains(out, sandbox.SandboxMarkerTag) {
		t.Errorf("the marker tag survived the section it tracks: %s", out)
	}

	// a plain workload is untouched
	verr, out = sandboxCheck(t, `{"kind":"workload","name":"plain","spec":{"type":"serverless","containers":[{"name":"app","image":"nginx:1","port":8080}]}}`)
	assertErr(t, verr, "")
	if strings.Contains(out, sandbox.SandboxMarkerTag) || strings.Contains(out, `"env"`) {
		t.Errorf("a plain workload took the overlay: %s", out)
	}
}
