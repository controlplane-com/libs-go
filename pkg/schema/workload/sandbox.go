/* The sandbox subtype's write-time overlay (sandbox.ts applySandboxSubtype). It lives
 * here, not in the sandbox package, because it rewrites the WORKLOAD — ports, env,
 * volumes, probe, firewall and tags — and the sandbox package must stay a leaf the
 * workload can import. */

package workload

import (
	"sort"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/env"
	"github.com/controlplane-com/libs-go/pkg/schema/port"
	"github.com/controlplane-com/libs-go/pkg/schema/sandbox"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"github.com/controlplane-com/libs-go/pkg/schema/volumeSpec"
)

// applySandboxSubtype normalizes spec.sandbox into the workload and maintains the
// marker tag in both directions. Pure and idempotent: reserved entries must be absent
// or exactly canonical, so a stored spec round-trips unchanged. orgName/gvcName are
// vestigial (the retired password-secret/identity model used them) and unused now.
func applySandboxSubtype(w any, spec *WorkloadSpec, orgName, gvcName string, h *validate.Helper) error {
	section := spec.Sandbox
	if section == nil {
		if tags, _ := validate.Get(w, "tags").(map[string]any); tags != nil {
			delete(tags, sandbox.SandboxMarkerTag)
		}
		return nil
	}

	sandboxType := string(section.Type)
	if sandboxType == "" {
		sandboxType = "dev"
	}
	daemon := sandbox.SandboxRunsDaemon(sandboxType) // worker: devboxd listens, the token gates it

	if spec.Type != "stateful" {
		return h.Errorf("workload.sandbox", "spec.sandbox type '%s' requires spec.type 'stateful' (type is '%s'): it keeps a home volume", sandboxType, spec.Type)
	}

	if daemon {
		section.Ide = "" // a worker has no IDE
	}

	containers := spec.Containers
	if len(containers) == 0 {
		return nil
	}
	dev := &containers[0]
	name := validate.Str(w, "name")

	// the standard dev image has no devboxd: a worker on it never becomes ready
	if daemon && string(dev.Image) == sandbox.SandboxStandardImage {
		return h.Errorf("workload.sandbox", "spec.containers[0].image: the standard sandbox image runs the dev type only; a %s sandbox needs an image that runs devboxd on port %d - set spec.containers[0].image to one (for example %s)", sandboxType, sandbox.SandboxDevboxdPort, sandbox.SandboxWorkerStandardImage)
	}

	// reserved ports are pod-wide; only the injected entry (gateway for dev, devboxd for
	// worker) is allowed, on the dev container
	canonical := sandbox.SandboxGatewayPort
	if daemon {
		canonical = sandbox.SandboxDevboxdPort
	}
	for i := range containers {
		c := &containers[i]
		for _, p := range c.Ports {
			if !reservedSandboxPort(float64(p.Number)) {
				continue
			}
			if i == 0 && int(p.Number) == canonical && p.Protocol == ContainerSpecPortsProtocolHttp {
				continue
			}
			return reservedSandboxEntry(h, "port", jsNum(float64(p.Number)), "(container '"+string(c.Name)+"')")
		}
		if c.Port != nil && reservedSandboxPort(float64(*c.Port)) {
			return reservedSandboxEntry(h, "port", jsNum(float64(*c.Port)), "(container '"+string(c.Name)+"')")
		}
	}
	if dev.Port != nil {
		return h.Errorf("workload.sandbox", "the sandbox dev container must declare ports via ports[]; 'port' cannot coexist with the injected %d port", canonical)
	}
	if !hasPortNumber(dev.Ports, canonical) {
		injected := ContainerSpecPorts{Protocol: ContainerSpecPortsProtocolHttp, Number: port.Port(canonical)}
		dev.Ports = append([]ContainerSpecPorts{injected}, dev.Ports...)
	}

	// derived env: APP_PORT from appPort; DEVBOXD_TOKEN from token and
	// DEVBOXD_MAX_SESSIONS from scaling (worker); a dev box carries no credential
	section.Password = ""
	if daemon {
		if strings.TrimSpace(section.Token) == "" {
			return h.Errorf("workload.sandbox", "spec.sandbox.token is required for a %s sandbox: it secures the devboxd API. Omit it on create and the server generates one.", sandboxType)
		}
	} else {
		section.Token = ""
	}
	var credential *env.EnvVar
	if daemon {
		credential = &env.EnvVar{Name: "DEVBOXD_TOKEN", Value: env.EnvValue(section.Token)}
	}
	appPort := 8080.0
	if section.AppPort != nil {
		appPort = float64(*section.AppPort)
	}
	bootEnv := []env.EnvVar{{Name: "APP_PORT", Value: env.EnvValue(jsNum(appPort))}}
	if credential != nil {
		bootEnv = append(bootEnv, *credential)
		maxSessions := float64(sandbox.SandboxScalingMaxSessionsDefault)
		if section.Scaling != nil && section.Scaling.MaxSessions != nil {
			maxSessions = *section.Scaling.MaxSessions
		}
		bootEnv = append(bootEnv, env.EnvVar{Name: "DEVBOXD_MAX_SESSIONS", Value: env.EnvValue(jsNum(maxSessions))})
	}

	for i := range containers {
		for _, e := range containers[i].Env {
			if !reservedSandboxEnv(string(e.Name)) {
				continue
			}
			if i != 0 {
				return h.Errorf("workload.sandbox", "spec.containers[%d].env: '%s' is reserved by the sandbox runtime (dev container only)", i, e.Name)
			}
			if e.Name == "CODE_SERVER_PASSWORD" {
				continue // retired: a legacy row's entry is dropped below
			}
			// a stored canonical entry round-trips; a different value belongs in the section field
			derived := findBootEnv(bootEnv, string(e.Name))
			if derived == nil {
				return h.Errorf("workload.sandbox", "spec.containers[0].env: '%s' is reserved by the sandbox runtime (a %s sandbox carries none)", e.Name, sandboxType)
			}
			if e.Value != derived.Value {
				field := sandbox.SandboxDerivedEnvFields[string(e.Name)]
				return h.Errorf("workload.sandbox", "spec.containers[0].env: %s is derived from %s; set %s instead", e.Name, field, field)
			}
		}
	}
	// canonicalize: derived entries replace stale ones; sorted to the EnvCollection fixpoint
	merged := make(env.EnvCollection, 0, len(dev.Env)+len(bootEnv))
	for _, e := range dev.Env {
		if !reservedSandboxEnv(string(e.Name)) {
			merged = append(merged, e)
		}
	}
	merged = append(merged, bootEnv...)
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].Name < merged[j].Name })
	dev.Env = merged

	// every sandbox mounts its home volumeset at the reserved path
	if name != "" {
		homeUri := volumeSpec.VolumeSpecUri("cpln://volumeset/" + sandbox.SandboxHomeVolumesetName(name))
		for _, v := range dev.Volumes {
			if v.Uri == homeUri && string(v.Path) != sandbox.SandboxReservedMountPath {
				return h.Errorf("workload.sandbox", "%s is the sandbox home volume and must be mounted at %s", homeUri, sandbox.SandboxReservedMountPath)
			}
		}
		// the overlay owns the reserved path outright: stale entries (e.g. a cloned spec
		// carrying the source sandbox's volume uri) are rewritten, not rejected
		kept := make(VolumeList, 0, len(dev.Volumes)+1)
		for _, v := range dev.Volumes {
			p := string(v.Path)
			if p == sandbox.SandboxReservedMountPath || strings.HasPrefix(p, sandbox.SandboxReservedMountPath+"/") {
				continue
			}
			kept = append(kept, v)
		}
		dev.Volumes = append(kept, volumeSpec.VolumeSpec{
			Uri:            homeUri,
			RecoveryPolicy: "retain",
			Path:           volumeSpec.VolumeSpecPath(sandbox.SandboxReservedMountPath),
		})
	}

	// default the readiness probe (gateway for dev, devboxd for worker); the owner may replace it
	if dev.ReadinessProbe == nil {
		probePath := "/healthz"
		probePort := port.Port(sandbox.SandboxGatewayPort)
		if daemon {
			probePath = "/_health"
			probePort = port.Port(sandbox.SandboxDevboxdPort)
		}
		dev.ReadinessProbe = &ContainerSpecReadinessProbe{
			HttpGet: &ContainerSpecReadinessProbeHttpGet{
				Path: probePath, Port: &probePort, Scheme: ContainerSpecReadinessProbeHttpGetSchemeHttp,
			},
			InitialDelaySeconds: f64(1),
			PeriodSeconds:       f64(5),
			TimeoutSeconds:      f64(1),
			SuccessThreshold:    f64(1),
			FailureThreshold:    f64(20),
		}
	}

	// open by default: the sign-in (dev) or the token (worker) gates access, not the
	// firewall; a caller's firewallConfig wins
	if spec.FirewallConfig == nil {
		spec.FirewallConfig = &WorkloadSpecFirewallConfig{
			External: &WorkloadSpecFirewallConfigExternal{
				InboundAllowCIDR:      CIDRList{"0.0.0.0/0"},
				InboundBlockedCIDR:    CIDRList{},
				OutboundAllowHostname: []DnsName{},
				OutboundAllowCIDR:     CIDRList{"0.0.0.0/0"},
				OutboundBlockedCIDR:   CIDRList{},
			},
		}
	}

	// cpln/scaleToZero gates the kata scale-to-zero machinery; the autoscaling field
	// alone does nothing. kataPrivileged: devboxd needs it for nested containers; it
	// breaks cpln workload exec, so a dev box is not stamped.
	tags, _ := validate.Get(w, "tags").(map[string]any)
	if tags == nil {
		tags = map[string]any{}
		validate.Set(w, "tags", base.Tags(tags))
	}
	tags[sandbox.SandboxMarkerTag] = "true"
	tags["cpln/kata"] = "true"
	tags["cpln/scaleToZero"] = "true"
	delete(tags, sandbox.SandboxIdePasswordTag)
	if daemon {
		tags["cpln/kataPrivileged"] = "true"
		delete(tags, sandbox.SandboxDevTypeTag)
	} else {
		tags[sandbox.SandboxDevTypeTag] = sandbox.SandboxDevTypeValue
	}
	// one replica, one home volume; rewriting numbers on a validated object stays at the
	// schema fixpoint
	if spec.DefaultOptions != nil && spec.DefaultOptions.Autoscaling != nil {
		spec.DefaultOptions.Autoscaling.MinScale = f64(1)
		spec.DefaultOptions.Autoscaling.MaxScale = f64(1)
		delay := float64(sandbox.SandboxScaleToZeroDelayDefault)
		if section.ScaleToZeroDelay != nil {
			delay = *section.ScaleToZeroDelay
		}
		spec.DefaultOptions.Autoscaling.ScaleToZeroDelay = &delay
	}
	if len(tags) > 50 {
		return h.Errorf("workload.sandbox", "You cannot have more than 50 tags per object (the sandbox runtime adds cpln/sandbox, cpln/kata, cpln/scaleToZero and, per type, cpln/kataPrivileged or cpln/dev-type cpln/dev-ide-password)")
	}
	return nil
}

func f64(v float64) *float64 { return &v }

// jsNum renders a number as JS String(n) does, which is what the boot env and the
// reserved-entry messages interpolate.
func jsNum(v float64) string { return base.JSString(v) }

func hasPortNumber(ports []ContainerSpecPorts, n int) bool {
	for _, p := range ports {
		if int(p.Number) == n {
			return true
		}
	}
	return false
}

func reservedSandboxPort(n float64) bool {
	for _, p := range sandbox.SandboxReservedPorts {
		if int(n) == p {
			return true
		}
	}
	return false
}

// findBootEnv returns the derived boot entry of the given name, or nil (sandbox.ts
// bootEnv.find). A dev box carries only APP_PORT; a worker adds DEVBOXD_TOKEN and
// DEVBOXD_MAX_SESSIONS.
func findBootEnv(bootEnv []env.EnvVar, name string) *env.EnvVar {
	for i := range bootEnv {
		if string(bootEnv[i].Name) == name {
			return &bootEnv[i]
		}
	}
	return nil
}

func reservedSandboxEnv(name string) bool {
	for _, n := range sandbox.SandboxReservedEnvNames {
		if n == name {
			return true
		}
	}
	return false
}

func reservedSandboxEntry(h *validate.Helper, kind, name, detail string) error {
	return h.Errorf("workload.sandbox", "%s '%s' is reserved by the sandbox runtime %s", kind, name, detail)
}
