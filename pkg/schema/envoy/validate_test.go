package envoy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

const jwtAuthnType = "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication"

func jwtFilter(provider string) map[string]any {
	return map[string]any{"name": "envoy.filters.http.jwt_authn", "typed_config": map[string]any{
		"@type":     jwtAuthnType,
		"providers": map[string]any{provider: map[string]any{"issuer": "i"}},
		"rules":     []any{map[string]any{"requires": map[string]any{"provider_name": provider}}},
	}}
}

func cplnCluster(name string) map[string]any {
	return map[string]any{"name": name, "type": "STATIC",
		"load_assignment": map[string]any{"cluster_name": name}}
}

func rule() *validate.Rule { return EnvoyFilters{}.ValidationRule() }

func TestUiFilterNamingValidator(t *testing.T) {
	cases := []validate.Case{
		{Name: "no cpln name at all", Input: map[string]any{
			"clusters": []any{map[string]any{"name": "a", "type": "STATIC"}}}},
		{Name: "cpln provider with no cluster", Input: map[string]any{"http": []any{jwtFilter("cpln_a")}},
			WantErr: "There does not exist a cluster filter using the name: cpln_a for the 'name' and 'load_assignment.cluster_name'"},
		{Name: "cpln cluster with no jwt filter", Input: map[string]any{
			"clusters": []any{cplnCluster("cpln_a")}},
			WantErr: "There does not exist a jwt authn filter using cpln_a as a provider"},
		{Name: "cluster name cpln_ but load_assignment name differs", Input: map[string]any{
			"clusters": []any{map[string]any{"name": "cpln_a", "type": "STATIC"}}},
			WantErr: "There does not exist a cluster filter using the name: cpln_a"},
		{Name: "provider without a matching rule", Input: map[string]any{
			"clusters": []any{cplnCluster("cpln_a")},
			"http": []any{map[string]any{"name": "envoy.filters.http.jwt_authn", "typed_config": map[string]any{
				"@type":     jwtAuthnType,
				"providers": map[string]any{"cpln_a": map[string]any{"issuer": "i"}}}}}},
			WantErr: "There does not exist a jwt authn filter using cpln_a as a provider"},
		{Name: "rule without a matching provider", Input: map[string]any{
			"clusters": []any{cplnCluster("cpln_a")},
			"http": []any{map[string]any{"name": "envoy.filters.http.jwt_authn", "typed_config": map[string]any{
				"@type": jwtAuthnType,
				"rules": []any{map[string]any{"requires": map[string]any{"provider_name": "cpln_a"}}}}}}},
			WantErr: "There does not exist a jwt authn filter using cpln_a as a provider"},
		{Name: "complete pair", Input: map[string]any{
			"clusters": []any{cplnCluster("cpln_a")}, "http": []any{jwtFilter("cpln_a")}}},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(rule()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestUiFilterNamingValidatorTypedWalk drives the same rule through the TYPED
// evaluator — the path a gvc/workload write actually takes, where `clusters` is a
// []envoyCluster.Cluster rather than a []any of maps.
func TestUiFilterNamingValidatorTypedWalk(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantErr string
	}{
		{name: "cpln provider with no cluster",
			body:    `{"http":[` + mustJSON(jwtFilter("cpln_a")) + `]}`,
			wantErr: "There does not exist a cluster filter using the name: cpln_a"},
		{name: "cpln cluster with no jwt filter",
			body:    `{"clusters":[` + mustJSON(cplnCluster("cpln_a")) + `]}`,
			wantErr: "There does not exist a jwt authn filter using cpln_a as a provider"},
		{name: "complete pair",
			body: `{"clusters":[` + mustJSON(cplnCluster("cpln_a")) + `],"http":[` + mustJSON(jwtFilter("cpln_a")) + `]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dst EnvoyFilters
			d, err := validate.DecodeInto(&dst, validate.Bytes(tc.body))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			verr := d.Validate(validate.For[EnvoyFilters]().Rule(), &validate.Context{}, validate.Prefs{})
			if tc.wantErr == "" {
				if verr != nil {
					t.Fatalf("unexpected error: %s", verr.Message)
				}
				return
			}
			if verr == nil {
				t.Fatalf("expected error containing %q, got success", tc.wantErr)
			}
			if got := verr.Message; !strings.Contains(got, tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, got)
			}
		})
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// TWO `cpln_` providers in one filter is the input that used to split the verdict:
// JS picks the FIRST INSERTED key, and a Go map has no insertion order, so the port
// picked the first SORTED one — and the name it picks decides which cluster/jwt pair
// the guard demands, so this both STORED a config TS refuses and REFUSED one TS
// stores. The engine now carries the body's key order (validate/order.go), so the
// answer follows the wire.
//
// The wants are the LIVE TS service's answers for the same bodies, captured
// 2026-09-16 against the oracle on POST /org/:org/gvc and POST .../workload (all four
// rows, both wire orders x both resolving clusters, on both twins).
func TestMultipleCplnProvidersFollowWireOrder(t *testing.T) {
	// The later filter has the last word on the name and no `rules` to override it,
	// so whichever provider the BODY listed first decides which cluster/jwt pair the
	// guard demands.
	trailing := func(first, second string) string {
		return `{"name":"envoy.filters.http.jwt_authn","typed_config":{"@type":"` + jwtAuthnType + `",` +
			`"providers":{"` + first + `":{"issuer":"https://` + first + `"},"` + second + `":{"issuer":"https://` + second + `"}}}}`
	}
	doc := func(cluster, first, second string) string {
		return `{"clusters":[` + mustJSON(cplnCluster(cluster)) + `],` +
			`"http":[` + mustJSON(jwtFilter(cluster)) + `,` + trailing(first, second) + `]}`
	}

	for _, tc := range []struct {
		name, body, wantErr string
	}{
		{name: "inserted-first resolves, sorted-first does not",
			body: doc("cpln_z", "cpln_z", "cpln_a")},
		{name: "sorted-first resolves but the body wrote it second",
			body:    doc("cpln_a", "cpln_z", "cpln_a"),
			wantErr: "There does not exist a cluster filter using the name: cpln_z"},
		{name: "sorted-first resolves and the body wrote it first",
			body: doc("cpln_a", "cpln_a", "cpln_z")},
		{name: "inserted-first is the sorted-last name",
			body:    doc("cpln_z", "cpln_a", "cpln_z"),
			wantErr: "There does not exist a cluster filter using the name: cpln_a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkErr(t, tc.wantErr, decodeAndValidate(t, tc.body))
			checkErr(t, tc.wantErr, decodeAndValidateGeneric(t, tc.body))
		})
	}
}

// decodeAndValidate runs the body through the TYPED evaluator, the path a gvc or
// workload write takes.
func decodeAndValidate(t *testing.T, body string) *validate.Error {
	t.Helper()
	var dst EnvoyFilters
	d, err := validate.DecodeInto(&dst, validate.Bytes(body))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return d.Validate(validate.For[EnvoyFilters]().Rule(), &validate.Context{}, validate.Prefs{})
}

// decodeAndValidateGeneric runs the SAME rule over the SAME bytes through the GENERIC
// evaluator, by decoding the document into an `any` field so the whole subtree stays a
// map. The two evaluators read one order table at one path, so a body that is accepted
// by one has to be accepted by the other — a split between them would be its own bug.
func decodeAndValidateGeneric(t *testing.T, body string) *validate.Error {
	t.Helper()
	var dst struct {
		Envoy any `json:"envoy"`
	}
	d, err := validate.DecodeInto(&dst, validate.Bytes(`{"envoy":`+body+`}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return d.Validate(validate.Object(map[string]*validate.Rule{"envoy": rule()}), &validate.Context{}, validate.Prefs{})
}

func checkErr(t *testing.T, wantErr string, verr *validate.Error) {
	t.Helper()
	if wantErr == "" {
		if verr != nil {
			t.Fatalf("expected the body to be accepted, got %q", verr.Message)
		}
		return
	}
	if verr == nil {
		t.Fatalf("expected error containing %q, got success", wantErr)
	}
	if !strings.Contains(verr.Message, wantErr) {
		t.Fatalf("expected error containing %q, got %q", wantErr, verr.Message)
	}
}

// A body whose order is GONE — a Go map handed straight to the generic evaluator, as
// the PATCH merge tree is — falls back to the order a jsonb column would hand back
// (length, then bytewise), which is the order TS reads the stored row in. Sorting
// lexicographically instead split the PATCH verdict for `cpln_z` vs `cpln_aa`.
func TestOrderlessValueFallsBackToStoredOrder(t *testing.T) {
	trailing := map[string]any{"name": "envoy.filters.http.jwt_authn", "typed_config": map[string]any{
		"@type": jwtAuthnType,
		"providers": map[string]any{
			"cpln_z":  map[string]any{"issuer": "https://cpln_z"},
			"cpln_aa": map[string]any{"issuer": "https://cpln_aa"},
		},
	}}
	doc := map[string]any{
		"clusters": []any{cplnCluster("cpln_aa")},
		"http":     []any{jwtFilter("cpln_aa"), trailing},
	}
	// jsonb hands back cpln_z (6 bytes) before cpln_aa (7), so the guard demands
	// cpln_z and the stored document no longer validates — which is what TS answers
	// for the same PATCH.
	_, err := validate.Validate(rule(), doc, nil, validate.Prefs{})
	if err == nil {
		t.Fatal("expected the shorter key to win the fallback order")
	}
	if !strings.Contains(err.Message, "There does not exist a cluster filter using the name: cpln_z") {
		t.Fatalf("expected cpln_z to drive the guard, got %q", err.Message)
	}
}
