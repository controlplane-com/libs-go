package sandbox

import (
	"encoding/json"
	"testing"
)

func TestRedactSandboxCredential(t *testing.T) {
	body := `{
		"kind":"workload","name":"wk",
		"tags":{"cpln/sandbox":"true","cpln/dev-ide-password":"hunter2"},
		"spec":{"sandbox":{"type":"worker","token":"supersecrettoken1"},
			"containers":[{"name":"dev","env":[
				{"name":"DEVBOXD_TOKEN","value":"supersecrettoken1"},
				{"name":"CODE_SERVER_PASSWORD","value":"pw"},
				{"name":"APP_PORT","value":"8080"},
				{"name":"USER_VAR","value":"keep"}]}]}}`
	var in map[string]any
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		t.Fatal(err)
	}
	out := RedactSandboxCredential(in)

	section := out["spec"].(map[string]any)["sandbox"].(map[string]any)
	if _, ok := section["token"]; ok {
		t.Error("token not stripped")
	}
	if out["tags"].(map[string]any)["cpln/dev-ide-password"] != "<redacted>" {
		t.Error("ide-password tag not redacted")
	}
	envs := out["spec"].(map[string]any)["containers"].([]any)[0].(map[string]any)["env"].([]any)
	got := map[string]string{}
	for _, e := range envs {
		m := e.(map[string]any)
		got[m["name"].(string)] = m["value"].(string)
	}
	if got["DEVBOXD_TOKEN"] != "<redacted>" || got["CODE_SERVER_PASSWORD"] != "<redacted>" {
		t.Errorf("credential env not redacted: %v", got)
	}
	if got["APP_PORT"] != "8080" || got["USER_VAR"] != "keep" {
		t.Errorf("non-secret env was altered: %v", got)
	}
	// the caller's object is untouched
	if in["tags"].(map[string]any)["cpln/dev-ide-password"] != "hunter2" {
		t.Error("input was mutated")
	}
	// a non-sandbox workload passes through unchanged
	plain := map[string]any{"spec": map[string]any{"type": "serverless"}}
	if RedactSandboxCredential(plain)["spec"].(map[string]any)["type"] != "serverless" {
		t.Error("non-sandbox workload altered")
	}
}
