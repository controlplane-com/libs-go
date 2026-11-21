package workload

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// check decodes body into T and validates it through the TYPED walk (the workload
// rules carry RefineT/CustomTyped container customs).
func check[T any](t *testing.T, rule *validate.Rule, body string) (*validate.Error, string) {
	t.Helper()
	var v T
	d, err := validate.DecodeInto(&v, validate.Bytes(body))
	if err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	verr := d.Validate(rule, &validate.Context{Org: "o", Gvc: "g", Method: "POST"}, validate.Prefs{})
	out, _ := json.Marshal(v)
	return verr, string(out)
}

func spec(t *testing.T, body string) (*validate.Error, string) {
	t.Helper()
	return check[WorkloadSpec](t, WorkloadSpec{}.ValidationRule(), body)
}

func status(t *testing.T, body string) (*validate.Error, string) {
	t.Helper()
	return check[WorkloadStatus](t, WorkloadStatus{}.ValidationRule(), body)
}

func assertErr(t *testing.T, got *validate.Error, want string) {
	t.Helper()
	switch {
	case want == "":
		if got != nil {
			t.Fatalf("unexpected error: %s (path=%v)", got.Message, got.Path)
		}
	case got == nil:
		t.Fatalf("expected an error containing %q, got success", want)
	case !strings.Contains(got.Message, want):
		t.Fatalf("got %q want it to contain %q", got.Message, want)
	}
}

const ctr = `{"name":"c1","image":"nginx","cpu":"50m","memory":"128Mi"}`

// cidrHolder gives the CIDRList element rule an object to hang off: the typed walk
// needs a struct destination, and the label under test is the array index.
type cidrHolder struct {
	Cidrs CIDRList `json:"cidrs,omitzero"`
}

// TestStatefulConcurrencyMetric pins workloadOptions.ts:43-48: `concurrency` is a
// supported stateful metric AND, being metrics[0], the DEFAULT one. Dropping it both
// refused requests TS accepts and silently changed the stored metric.
func TestStatefulConcurrencyMetric(t *testing.T) {
	verr, _ := spec(t, `{"type":"stateful","containers":[`+ctr+`],"defaultOptions":{"autoscaling":{"metric":"concurrency","target":50}}}`)
	assertErr(t, verr, "")

	verr, out := spec(t, `{"type":"stateful","containers":[`+ctr+`],"defaultOptions":{}}`)
	assertErr(t, verr, "")
	if !strings.Contains(out, `"metric":"concurrency"`) {
		t.Fatalf("stateful should default to concurrency: %s", out)
	}
}

// TestScaleToZeroKataEscape pins workloadOptions.ts:352-366 — minScale 0 is allowed
// via KEDA *or* kata (metric concurrency/rps) — and the exact refusal wording.
func TestScaleToZeroKataEscape(t *testing.T) {
	as := func(typ, metric string) string {
		target := `"target":50,`
		if metric == "keda" {
			target = "" // ResourceTarget forbids a target when the metric is keda
		}
		return `{"type":"` + typ + `","containers":[` + ctr + `],"defaultOptions":{"capacityAI":false,"autoscaling":{"metric":"` + metric + `",` + target + `"minScale":0,"maxScale":5}}}`
	}
	assertErr(t, first(spec(t, as("standard", "rps"))), "")
	assertErr(t, first(spec(t, as("standard", "keda"))), "")
	assertErr(t, first(spec(t, as("stateful", "concurrency"))), "")
	assertErr(t, first(spec(t, as("standard", "cpu"))),
		"workloads of type 'standard' can only scale to zero via KEDA (metric 'keda') or kata (metric 'concurrency'/'rps'); please enable KEDA or set a kata scale-to-zero metric")
}

func first(e *validate.Error, _ string) *validate.Error { return e }

// TestKataMultiMetric pins workload.ts:1081-1098: concurrency/rps are scaled by the
// kata-agent, which is single-metric only, so they cannot appear in a multi set.
func TestKataMultiMetric(t *testing.T) {
	wl := func(tags, options string) string {
		return `{"kind":"workload","name":"w1","tags":` + tags +
			`,"spec":{"type":"standard","containers":[` + ctr + `],` + options + `}}`
	}
	rule := Workload{}.ValidationRule()
	multi := `"defaultOptions":{"autoscaling":{"multi":[{"metric":"rps","target":50}]}}`
	local := `"localOptions":[{"location":"//location/l","autoscaling":{"multi":[{"metric":"rps","target":50}]}}]`

	verr, _ := check[Workload](t, rule, wl(`{"cpln/kata":"true"}`, multi))
	assertErr(t, verr, "autoscaling metric 'rps' cannot be used in a multi-metric set on a kata (cpln/kata tag) workload")
	verr, _ = check[Workload](t, rule, wl(`{"cpln/kata":"true"}`, local))
	assertErr(t, verr, "autoscaling metric 'rps' cannot be used in a multi-metric set on a kata (cpln/kata tag) workload")
	// plural form, and both offenders named
	verr, _ = check[Workload](t, rule, wl(`{"cpln/kata":"true"}`,
		`"defaultOptions":{"autoscaling":{"multi":[{"metric":"rps","target":50},{"metric":"cpu","target":50}]}}`))
	assertErr(t, verr, "autoscaling metric 'rps' cannot be used")
	// cpu-only multi is fine on kata; rps multi is fine off kata
	assertErr(t, first(check[Workload](t, rule, wl(`{"cpln/kata":"true"}`,
		`"defaultOptions":{"autoscaling":{"multi":[{"metric":"cpu","target":50}]}}`))), "")
	assertErr(t, first(check[Workload](t, rule, wl(`{}`, multi))), "")
}

// TestGracePeriodBlamesRolloutOptions: TS attaches the custom to RolloutOptions, so
// Joi reports spec.rolloutOptions — the field the caller has to fix — not the root.
func TestGracePeriodBlamesRolloutOptions(t *testing.T) {
	body := `{"kind":"workload","name":"w1","spec":{"type":"standard","containers":[` + ctr + `],"rolloutOptions":{"terminationGracePeriodSeconds":1000}}}`
	verr, _ := check[Workload](t, Workload{}.ValidationRule(), body)
	assertErr(t, verr, "terminationGracePeriodSeconds must be at most 900.")
	if got := strings.Join(verr.Path, "."); got != "spec.rolloutOptions" {
		t.Fatalf("path %q, want spec.rolloutOptions", got)
	}
	if !strings.HasPrefix(verr.Message, `"spec.rolloutOptions" failed custom validation`) {
		t.Fatalf("message %q", verr.Message)
	}
}

// TestGpuXorArms pins workload.ts:289-300: TS remaps ONLY object.xor, so the
// "neither peer" arm keeps Joi's own object.missing message and type.
func TestGpuXorArms(t *testing.T) {
	rule := ContainerSpecGpu{}.ValidationRule()
	verr, _ := check[ContainerSpecGpu](t, rule, `{}`)
	assertErr(t, verr, `"value" must contain at least one of [nvidia, custom]`)
	if verr.Type != "object.missing" {
		t.Fatalf("type %q, want object.missing", verr.Type)
	}
	verr, _ = check[ContainerSpecGpu](t, rule, `{"nvidia":{"model":"t4"},"custom":{"resource":"amd.com/gpu"}}`)
	assertErr(t, verr, "The parameter spec.gpu must contain one of nvidia or custom and not both")
	if verr.Type != "object.xor" {
		t.Fatalf("type %q, want object.xor", verr.Type)
	}
	assertErr(t, first(check[ContainerSpecGpu](t, rule, `{"nvidia":{"model":"t4"}}`)), "")
}

// TestGpuResourceVsRuntimeClass pins the two DIFFERENT regexes of workload.ts:274-283.
// Sharing one type rejected every `_`/`-` extended-resource name TS accepts.
func TestGpuResourceVsRuntimeClass(t *testing.T) {
	rule := ContainerSpecGpu{}.ValidationRule()
	assertErr(t, first(check[ContainerSpecGpu](t, rule, `{"custom":{"resource":"my_vendor.com/gpu-a"}}`)), "")
	assertErr(t, first(check[ContainerSpecGpu](t, rule, `{"custom":{"resource":"amd.com/gpu","runtimeClass":"bad_class"}}`)),
		"fails to match the required pattern")
	assertErr(t, first(check[ContainerSpecGpu](t, rule, `{"custom":{"resource":"amd.com/gpu","runtimeClass":"nvidia"}}`)), "")
}

// TestCIDRLeadingZeros: Joi's ip() matches a pattern, so `010.1.1.1` is a valid
// dotted quad to it; net.ParseIP has rejected leading zeros since Go 1.17, which
// turned addresses TS stores into a 400.
func TestCIDRLeadingZeros(t *testing.T) {
	rule := validate.Object(map[string]*validate.Rule{"cidrs": CIDRList{}.ValidationRule()})
	for _, c := range []struct{ name, in, wantErr string }{
		{"leading zeros", `{"cidrs":["010.1.1.1"]}`, ""},
		{"leading zeros cidr", `{"cidrs":["010.001.1.0/24"]}`, ""},
		{"plain v4", `{"cidrs":["1.2.3.4"]}`, ""},
		{"v6", `{"cidrs":["2001:db8::1"]}`, ""},
		{"out of range", `{"cidrs":["999.1.1.1"]}`, "must be a valid ip address"},
		{"four digits", `{"cidrs":["0100.1.1.1"]}`, "must be a valid ip address"},
		{"three octets", `{"cidrs":["1.2.3"]}`, "must be a valid ip address"},
	} {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, first(check[cidrHolder](t, rule, c.in)), c.wantErr)
		})
	}
	// Joi's messages lead with the LABEL, not the offending value.
	verr, _ := check[cidrHolder](t, rule, `{"cidrs":["999.1.1.1"]}`)
	if !strings.HasPrefix(verr.Message, `"cidrs[0]"`) {
		t.Fatalf("message should lead with the label, got %q", verr.Message)
	}
	if verr.Type != "string.ipVersion" {
		t.Fatalf("type %q, want string.ipVersion", verr.Type)
	}
}

// TestServerlessEmptyPortsArray: `_.filter(c => c.port ?? c.ports)` — an empty array
// is truthy in JS, so `ports: []` counts as exposing a port. A TS quirk, but the
// accept/reject boundary clients were built against.
func TestServerlessEmptyPortsArray(t *testing.T) {
	assertErr(t, first(spec(t, `{"type":"serverless","containers":[{"name":"c1","image":"nginx","cpu":"50m","memory":"128Mi","ports":[]}]}`)), "")
	assertErr(t, first(spec(t, `{"type":"serverless","containers":[`+ctr+`]}`)),
		"workloads of type 'serverless' must have at least one container port exposed.")
}

// TestVolumesUniqueBlamesIndex: Joi's array.unique reports the DUPLICATE ELEMENT.
func TestVolumesUniqueBlamesIndex(t *testing.T) {
	body := `{"type":"standard","containers":[{"name":"c1","image":"nginx","cpu":"50m","memory":"128Mi",` +
		`"volumes":[{"uri":"cpln://volumeset/vs1","path":"/a"},{"uri":"cpln://volumeset/vs1","path":"/a"}]}]}`
	verr, _ := spec(t, body)
	assertErr(t, verr, `"containers[0].volumes[1]" contains a duplicate value`)
	if verr.Ctx["pos"] != 1 {
		t.Fatalf("context.pos = %v, want 1", verr.Ctx["pos"])
	}
}

// TestStatusSubtree pins workload.ts:546-591. The whole status subtree was
// unvalidated, so a root writer could persist a status downstream cannot trust.
func TestStatusSubtree(t *testing.T) {
	for _, c := range []struct{ name, body, wantErr string }{
		{"healthCheck.active required", `{"healthCheck":{"success":true}}`, `"healthCheck.active" is required`},
		{"code too high", `{"healthCheck":{"active":true,"code":999}}`, "less than or equal to 599"},
		{"code too low", `{"healthCheck":{"active":true,"code":100}}`, "greater than or equal to 200"},
		{"failures negative", `{"healthCheck":{"active":true,"failures":-1}}`, "greater than or equal to 0"},
		{"successes fractional", `{"healthCheck":{"active":true,"successes":1.5}}`, "must be an integer"},
		{"lastChecked not a date", `{"healthCheck":{"active":true,"lastChecked":"not-a-date"}}`, "must be a valid date"},
		{"resolvedAt not a date", `{"resolvedImages":{"resolvedAt":"zzz"}}`, "must be a valid date"},
		{"nextRetryAt not a date", `{"resolvedImages":{"nextRetryAt":"zzz"}}`, "must be a valid date"},
		{"resolvedForVersion fractional", `{"resolvedImages":{"resolvedForVersion":1.5}}`, "must be an integer"},
		{"manifest fields required", `{"resolvedImages":{"images":[{"manifests":[{"image":"a"}]}]}}`, `"resolvedImages.images[0].manifests[0].mediaType" is required`},
		{"loadBalancer url", `{"loadBalancer":[{"url":"not a url"}]}`, "must be a valid uri"},
		// Joi.object().pattern(string, string().allow('')): a registry really does
		// return an empty `variant`, and the generic map rule refused it.
		{"platform empty value", `{"resolvedImages":{"images":[{"manifests":[{"image":"a","mediaType":"m","digest":"d","platform":{"os":"linux","variant":""}}]}]}}`, ""},
		{"valid status", `{"healthCheck":{"active":true,"code":200,"failures":0,"successes":3,"lastChecked":"2026-01-01T00:00:00Z"},"loadBalancer":[{"url":"https://x.y/z"}]}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, first(status(t, c.body)), c.wantErr)
		})
	}
}

// TestExcludedContainerNames pins `Name.disallow(...ExcludedContainerNames)`
// (workload.ts:303). Spelling it as a custom lost Joi's any.invalid envelope: the
// message named the field "name" instead of spec.containers[0].name, and the
// `invalids` list a client renders its own message from was absent entirely.
func TestExcludedContainerNames(t *testing.T) {
	for _, c := range []struct{ name, wantErr string }{
		{"istio-proxy", "contains an invalid value"},
		{"queue-proxy", "contains an invalid value"},
		{"cpln-dbaas-config", "contains an invalid value"},
		{"app", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			err, _ := spec(t, `{"type":"standard","containers":[{"name":"`+c.name+`","image":"nginx"}]}`)
			assertErr(t, err, c.wantErr)
			if c.wantErr == "" {
				return
			}
			d := err.Details()[0]
			if d.Type != "any.invalid" {
				t.Fatalf("type = %q, want any.invalid", d.Type)
			}
			if d.Context["label"] != "containers[0].name" {
				t.Fatalf("label = %v, want containers[0].name", d.Context["label"])
			}
			inv, _ := d.Context["invalids"].([]any)
			if len(inv) != len(wlExcludedContainerNames) {
				t.Fatalf("invalids = %v, want the full disallow list", d.Context["invalids"])
			}
			if inv[0] != "istio-proxy" {
				t.Fatalf("invalids order changed: %v", inv)
			}
		})
	}
}

// TestKataScaleToZero pins the shapes the actuator's kata scale-to-zero relies on
// (workloadOptions.ts:352-366): minScale 0 is legal on a 'keda'-capable workload type
// via KEDA *or* a concurrency/rps metric. Rejecting these is a Go-only 400 on
// workloads TypeScript accepts, i.e. a production outage for the kata path.
func TestKataScaleToZero(t *testing.T) {
	as := func(typ, opts string) string {
		return `{"type":"` + typ + `","containers":[{"name":"c1","image":"nginx","ports":[{"number":8080,"protocol":"http"}]}],"defaultOptions":{"autoscaling":` + opts + `}}`
	}
	for _, c := range []struct{ name, typ, opts, wantErr string }{
		{"stateful concurrency", "stateful", `{"metric":"concurrency","target":100}`, ""},
		{"stateful min0 concurrency", "stateful", `{"metric":"concurrency","target":100,"minScale":0,"maxScale":5}`, ""},
		{"stateful min0 rps", "stateful", `{"metric":"rps","target":100,"minScale":0,"maxScale":5}`, ""},
		{"standard min0 rps", "standard", `{"metric":"rps","target":100,"minScale":0,"maxScale":5}`, ""},
		{"standard min0 cpu", "standard", `{"metric":"cpu","target":50,"minScale":0,"maxScale":5}`, "can only scale to zero via KEDA"},
		{"stateful min0 latency", "stateful", `{"metric":"latency","target":50,"minScale":0,"maxScale":5}`, "can only scale to zero via KEDA"},
		{"serverless min0", "serverless", `{"metric":"concurrency","target":100,"minScale":0,"maxScale":5}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			err, _ := spec(t, as(c.typ, c.opts))
			assertErr(t, err, c.wantErr)
		})
	}
}

// TestProbePortDefaults pins the probe loop of ContainerSpec.custom (workload.ts:425,
// master 160c232a6): every probe kind defaults its port from the container, grpc
// preferring a grpc port, and only the types that deploy probes as-is require one.
func TestProbePortDefaults(t *testing.T) {
	rule := Workload{}.ValidationRule()
	wl := func(specType, containers string) string {
		job := ""
		if specType == "cron" {
			job = `"job":{"schedule":"* * * * *"},`
		}
		return `{"name":"wl1","version":1,"spec":{"type":"` + specType + `",` + job + `"containers":[` + containers + `]}}`
	}
	ctr := func(extra string) string {
		return `{"name":"main","image":"image","cpu":"50m","memory":"128Mi",` + extra + `}`
	}
	for _, kind := range []string{"httpGet", "tcpSocket", "grpc"} {
		probe := func(body string) string { return `{"` + kind + `":` + body + `}` }

		verr, out := check[Workload](t, rule, wl("standard", ctr(
			`"port":8080,"readinessProbe":`+probe(`{"port":8081}`)+`,"livenessProbe":`+probe(`{"port":8082}`))))
		assertErr(t, verr, "")
		wantPorts(t, kind, out, 8081, 8082)

		verr, out = check[Workload](t, rule, wl("standard", ctr(
			`"ports":[{"number":8080,"protocol":"http"}],"readinessProbe":`+probe(`{}`)+`,"livenessProbe":`+probe(`{}`))))
		assertErr(t, verr, "")
		wantPorts(t, kind, out, 8080, 8080)

		verr, out = check[Workload](t, rule, wl("standard", ctr(`"port":8080,"readinessProbe":`+probe(`{}`))))
		assertErr(t, verr, "")
		wantPorts(t, kind, out, 8080, 0)

		verr, _ = check[Workload](t, rule, wl("standard", ctr(`"livenessProbe":`+probe(`{}`))))
		assertErr(t, verr, "Probes using "+kind+" must have a port defined.")

		// cron strips probes and serverless drops them from a portless container, so
		// neither requires a port
		verr, out = check[Workload](t, rule, wl("cron", ctr(`"readinessProbe":`+probe(`{}`))))
		assertErr(t, verr, "")
		if strings.Contains(out, "readinessProbe") {
			t.Errorf("cron kept the probe it strips: %s", out)
		}
		verr, _ = check[Workload](t, rule, wl("serverless",
			ctr(`"port":8080`)+`,{"name":"side","image":"image","cpu":"50m","memory":"128Mi","livenessProbe":`+probe(`{}`)+`}`))
		assertErr(t, verr, "")
	}

	verr, out := check[Workload](t, rule, wl("standard", ctr(
		`"ports":[{"number":8080,"protocol":"http"},{"number":8090,"protocol":"grpc"}],"readinessProbe":{"grpc":{}}`)))
	assertErr(t, verr, "")
	wantPorts(t, "grpc", out, 8090, 0)
}

// wantPorts reads the probe ports back off the rendered workload; 0 means "expect none".
func wantPorts(t *testing.T, kind, out string, readiness, liveness int) {
	t.Helper()
	var w struct {
		Spec struct {
			Containers []map[string]any `json:"containers"`
		} `json:"spec"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct {
		probe string
		want  int
	}{{"readinessProbe", readiness}, {"livenessProbe", liveness}} {
		section, _ := w.Spec.Containers[0][p.probe].(map[string]any)
		probe, _ := section[kind].(map[string]any)
		got, ok := probe["port"].(float64)
		if p.want == 0 {
			if ok {
				t.Errorf("%s.%s: port %v, want none", p.probe, kind, got)
			}
			continue
		}
		if !ok || int(got) != p.want {
			t.Errorf("%s.%s: port %v (present=%v), want %d", p.probe, kind, got, ok, p.want)
		}
	}
}

// TestVolumeRootPathRejected pins volumeSpec.ts:85-97 + workload.ts:1054: a volume
// mounted on "/" trims to "" and is rejected as "volumes[].path is required", never
// mounted over the container root (the Go path custom used to remap "" back to "/").
func TestVolumeRootPathRejected(t *testing.T) {
	body := `{"type":"standard","containers":[{"name":"c1","image":"nginx","cpu":"50m","memory":"128Mi",` +
		`"volumes":[{"uri":"cpln://volumeset/vs1","path":"/"}]}]}`
	verr, _ := spec(t, body)
	assertErr(t, verr, "container 'c1': volumes[].path is required")
	// A real mount path is still accepted.
	ok := `{"type":"standard","containers":[{"name":"c1","image":"nginx","cpu":"50m","memory":"128Mi",` +
		`"volumes":[{"uri":"cpln://volumeset/vs1","path":"/data"}]}]}`
	okErr, _ := spec(t, ok)
	assertErr(t, okErr, "")
}

// rolloutOnReferenceChange is an optional boolean that round-trips (explicit true
// and false) on non-vm workloads and is stripped for type vm (eeeb7c73f).
func TestRolloutOnReferenceChange(t *testing.T) {
	for _, v := range []string{"true", "false"} {
		verr, out := spec(t, `{"type":"standard","containers":[`+ctr+`],"rolloutOptions":{"rolloutOnReferenceChange":`+v+`}}`)
		assertErr(t, verr, "")
		if !strings.Contains(out, `"rolloutOnReferenceChange":`+v) {
			t.Fatalf("standard dropped rolloutOnReferenceChange=%s: %s", v, out)
		}
	}
	_, out := spec(t, `{"type":"vm","containers":[`+ctr+`],"vm":{},"rolloutOptions":{"rolloutOnReferenceChange":true,"scalingPolicy":"Parallel"}}`)
	if strings.Contains(out, "rolloutOnReferenceChange") {
		t.Fatalf("vm did not strip rolloutOnReferenceChange: %s", out)
	}
}
