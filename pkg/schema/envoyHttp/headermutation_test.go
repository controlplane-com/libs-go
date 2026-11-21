package envoyHttp

import (
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// Mirrors nodelibs/schema/src/envoyHttp.test.ts header_mutation cases (master 894c1be3c).
func TestHeaderMutationFilter(t *testing.T) {
	filter := func(mutations map[string]any) map[string]any {
		return map[string]any{
			"name":     "envoy.filters.http.header_mutation",
			"priority": float64(10),
			"typed_config": map[string]any{
				"@type":     "type.googleapis.com/envoy.extensions.filters.http.header_mutation.v3.HeaderMutation",
				"mutations": mutations,
			},
		}
	}
	validateFilter := func(f map[string]any) error {
		_, err := HttpFilterRule().ValidateAt(
			[]string{"spec", "sidecar", "envoy", "http", validate.IndexPath(0)}, f, nil, validate.Prefs{})
		return err
	}
	rules := func(rs ...map[string]any) map[string]any {
		arr := make([]any, len(rs))
		for i, r := range rs {
			arr[i] = r
		}
		return map[string]any{"request_mutations": arr}
	}

	// valid
	valid := []map[string]any{
		rules(map[string]any{"remove": "x-kintsugi-authctx-shadow"}, map[string]any{"remove": "x-envoy-auth-failure-mode-allowed"}),
		{
			"request_mutations": []any{
				map[string]any{"append": map[string]any{"header": map[string]any{"key": "x-forwarded-by", "value": "sidecar"}, "append_action": "OVERWRITE_IF_EXISTS_OR_ADD"}},
				map[string]any{"remove": "x-debug"},
			},
			"response_mutations": []any{
				map[string]any{"remove": "server"},
				map[string]any{"append": map[string]any{"header": map[string]any{"key": "x-frame-options", "value": "DENY"}}},
			},
		},
	}
	for i, m := range valid {
		if err := validateFilter(filter(m)); err != nil {
			t.Errorf("valid[%d] rejected: %v", i, err)
		}
	}

	// rejected rules (each placed as the sole request_mutations entry)
	bad := map[string]map[string]any{
		"both remove and append":    {"remove": "x-a", "append": map[string]any{"header": map[string]any{"key": "x-b", "value": "v"}}},
		"neither remove nor append": {},
		"pseudo-header remove":      {"remove": ":path"},
		"host remove":               {"remove": "host"},
		"uppercase remove":          {"remove": "X-Foo"},
		"invalid remove":            {"remove": "x foo"},
		"empty remove":              {"remove": ""},
		"pseudo-header append":      {"append": map[string]any{"header": map[string]any{"key": ":authority", "value": "evil"}}},
		"host append":               {"append": map[string]any{"header": map[string]any{"key": "host", "value": "evil"}}},
		"append without header":     {"append": map[string]any{"append_action": "ADD_IF_ABSENT"}},
		"unknown rule field":        {"remove_on_match": map[string]any{"key_matcher": map[string]any{"exact": "x-a"}}},
		"oversized remove":          {"remove": strings.Repeat("x", 16385)},
		"append and append_action":  {"append": map[string]any{"header": map[string]any{"key": "x-a", "value": "v"}, "append": true, "append_action": "ADD_IF_ABSENT"}},
		"CRLF in append value":      {"append": map[string]any{"header": map[string]any{"key": "x-a", "value": "a\r\nx-evil: 1"}}},
		"NUL in append value":       {"append": map[string]any{"header": map[string]any{"key": "x-a", "value": "a\x00b"}}},
	}
	for name, rule := range bad {
		if err := validateFilter(filter(rules(rule))); err == nil {
			t.Errorf("%q accepted, want rejected", name)
		}
	}

	// an unexposed mutation kind
	if err := validateFilter(filter(map[string]any{"query_parameter_mutations": []any{map[string]any{"remove": "a"}}})); err == nil {
		t.Error("query_parameter_mutations accepted, want rejected")
	}

	// a mismatched @type
	f := filter(rules(map[string]any{"remove": "x-a"}))
	f["typed_config"].(map[string]any)["@type"] = "type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors"
	if err := validateFilter(f); err == nil {
		t.Error("mismatched @type accepted, want rejected")
	}
}
