package sandbox

import "encoding/json"

// RedactSandboxCredential returns a copy of a workload map with the worker token, the
// retired password and the credential env values removed (sandbox.ts
// redactSandboxCredential). A non-sandbox workload comes back unchanged. The audit and
// unauthorized-read paths call it so a worker's DEVBOXD_TOKEN never lands in a record or
// a response.
func RedactSandboxCredential(workload map[string]any) map[string]any {
	spec, _ := workload["spec"].(map[string]any)
	if spec == nil {
		return workload
	}
	if _, ok := spec["sandbox"].(map[string]any); !ok {
		return workload
	}
	out := deepCopyMap(workload)
	outSpec, _ := out["spec"].(map[string]any)
	if section, _ := outSpec["sandbox"].(map[string]any); section != nil {
		delete(section, "password")
		delete(section, "token")
	}
	if tags, _ := out["tags"].(map[string]any); tags != nil {
		if _, ok := tags[SandboxIdePasswordTag]; ok {
			tags[SandboxIdePasswordTag] = "<redacted>"
		}
	}
	containers, _ := outSpec["containers"].([]any)
	for _, c := range containers {
		cm, _ := c.(map[string]any)
		envs, _ := cm["env"].([]any)
		for _, e := range envs {
			em, _ := e.(map[string]any)
			name, _ := em["name"].(string)
			if !redactedEnvName(name) {
				continue
			}
			if _, present := em["value"]; present {
				em["value"] = "<redacted>"
			}
		}
	}
	return out
}

// redactedEnvName reports whether an env value must be hidden: the reserved names whose
// values are credentials (SandboxCredentialEnvNames).
func redactedEnvName(name string) bool {
	for _, n := range SandboxCredentialEnvNames {
		if n == name {
			return true
		}
	}
	return false
}

// deepCopyMap round-trips through JSON, matching the TS JSON.parse(JSON.stringify) copy
// so the redaction never mutates the caller's object.
func deepCopyMap(m map[string]any) map[string]any {
	b, err := json.Marshal(m)
	if err != nil {
		return m
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return m
	}
	return out
}
