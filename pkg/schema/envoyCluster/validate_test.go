package envoyCluster

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func run(t *testing.T, rule *validate.Rule, cases []validate.Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(rule); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func cluster(kv ...any) map[string]any {
	m := map[string]any{"name": "c", "type": "STATIC"}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func cplnCluster(kv ...any) map[string]any {
	m := cluster(kv...)
	m["name"] = "cpln_x"
	return m
}

func TestClusterBasics(t *testing.T) {
	run(t, ruleOf[Cluster](), []validate.Case{
		{Name: "name required", Input: map[string]any{"type": "STATIC"}, WantErr: `"name" is required`},
		{Name: "type enum", Input: map[string]any{"name": "c", "type": "BOGUS"},
			WantErr: "must be one of [STATIC, STRICT_DNS, LOGICAL_DNS, EDS, ORIGINAL_DST]"},
		{Name: "xor none", Input: map[string]any{"name": "c"},
			WantErr: "must contain at least one of [type, cluster_type]"},
		{Name: "xor both", Input: cluster("cluster_type", map[string]any{"name": "n"}),
			WantErr: "contains a conflict between exclusive peers [type, cluster_type]"},
		{Name: "load_assignment cluster_name required",
			Input:   cluster("load_assignment", map[string]any{}),
			WantErr: `"load_assignment.cluster_name" is required`},
	})
}

// TestCplnGuardForbids is the security-relevant half: on a cpln_-named cluster every
// guarded key is forbidden outright.
func TestCplnGuardForbids(t *testing.T) {
	values := map[string]any{
		"transport_socket_matches":                  []any{map[string]any{"name": "n"}},
		"alt_stat_name":                             "s",
		"cluster_type":                              map[string]any{"name": "n"},
		"eds_cluster_config":                        map[string]any{},
		"connect_timeout":                           "5s",
		"per_connection_buffer_limit_bytes":         5.0,
		"lb_policy":                                 "ROUND_ROBIN",
		"health_checks":                             []any{},
		"max_requests_per_connection":               1.0,
		"circuit_breakers":                          map[string]any{},
		"upstream_http_protocol_options":            map[string]any{},
		"common_http_protocol_options":              map[string]any{},
		"http_protocol_options":                     map[string]any{},
		"http2_protocol_options":                    map[string]any{},
		"typed_extension_protocol_options":          []any{},
		"dns_refresh_rate":                          "5s",
		"dns_failure_refresh_rate":                  map[string]any{"base_interval": "5s"},
		"respect_dns_ttl":                           true,
		"dns_lookup_family":                         "AUTO",
		"dns_resolvers":                             []any{},
		"use_tcp_for_dns_lookups":                   true,
		"dns_resolution_config":                     map[string]any{},
		"typed_dns_resolver_config":                 map[string]any{"name": "n"},
		"wait_for_warm_on_init":                     true,
		"outlier_detection":                         map[string]any{},
		"cleanup_interval":                          "5s",
		"upstream_bind_config":                      map[string]any{},
		"lb_subset_config":                          map[string]any{},
		"ring_hash_lb_config":                       map[string]any{},
		"maglev_lb_config":                          map[string]any{},
		"least_request_lb_config":                   map[string]any{},
		"common_lb_config":                          map[string]any{},
		"metadata":                                  map[string]any{},
		"protocol_selection":                        "USE_CONFIGURED_PROTOCOL",
		"upstream_connection_options":               map[string]any{},
		"close_connections_on_host_health_failure":  true,
		"ignore_health_on_host_removal":             true,
		"filters":                                   []any{map[string]any{"name": "f"}},
		"load_balancing_policy":                     map[string]any{},
		"track_timeout_budgets":                     true,
		"upstream_config":                           map[string]any{"name": "n"},
		"track_cluster_stats":                       map[string]any{},
		"preconnect_policy":                         map[string]any{},
		"connection_pool_per_downstream_connection": true,
	}
	if len(values) != 44 {
		t.Fatalf("expected the 44 guarded fields, listed %d", len(values))
	}
	var cases []validate.Case
	for k, v := range values {
		in := cplnCluster(k, v)
		if k == "cluster_type" { // xor with type
			delete(in, "type")
		}
		cases = append(cases, validate.Case{Name: k, Input: in, WantErr: `"` + k + `" is not allowed`})
		// the SAME value on an unguarded cluster must be accepted
		plain := cluster(k, v)
		if k == "cluster_type" {
			delete(plain, "type")
		}
		cases = append(cases, validate.Case{Name: k + "/unguarded", Input: plain})
	}
	run(t, ruleOf[Cluster](), cases)
}

// TestCplnGuardRestrictedBranches covers the three keys whose restricted branch is a
// narrower schema rather than `forbidden`.
func TestCplnGuardRestrictedBranches(t *testing.T) {
	run(t, ruleOf[Cluster](), []validate.Case{
		{Name: "transport_socket still validated", Input: cplnCluster("transport_socket", map[string]any{}),
			WantErr: `"transport_socket.name" is required`},
		{Name: "transport_socket ok", Input: cplnCluster("transport_socket", map[string]any{"name": "n"})},
		{Name: "endpoints restricted to socket_address",
			Input: cplnCluster("load_assignment", map[string]any{"cluster_name": "cpln_x",
				"endpoints": []any{map[string]any{"locality": map[string]any{"region": "r"}}}}),
			WantErr: `"load_assignment.endpoints[0].locality" is not allowed`},
		{Name: "endpoints capped at one",
			Input: cplnCluster("load_assignment", map[string]any{"cluster_name": "cpln_x",
				"endpoints": []any{map[string]any{}, map[string]any{}}}),
			WantErr: "must contain less than or equal to 1 items"},
		{Name: "endpoints restricted shape ok",
			Input: cplnCluster("load_assignment", map[string]any{"cluster_name": "cpln_x",
				"endpoints": []any{map[string]any{"lb_endpoints": []any{map[string]any{
					"endpoint": map[string]any{"address": map[string]any{
						"socket_address": map[string]any{"address": "1.2.3.4", "port_value": 80.0}}}}}}}})},
		{Name: "policy forbidden",
			Input: cplnCluster("load_assignment", map[string]any{"cluster_name": "cpln_x",
				"policy": map[string]any{}}),
			WantErr: `"load_assignment.policy" is not allowed`},
		{Name: "unguarded endpoints keep the full schema",
			Input: cluster("load_assignment", map[string]any{"cluster_name": "c",
				"endpoints": []any{map[string]any{"locality": map[string]any{"region": "r"}, "priority": 0.0}}})},
		{Name: "unguarded policy allowed",
			Input: cluster("load_assignment", map[string]any{"cluster_name": "c",
				"policy": map[string]any{"overprovisioning_factor": 3.0}})},
	})
}

// TestUnguardedSchemasAreReal proves the `otherwise` branch is the real schema, not
// a pass-through `any` — these are the rules that were lost with the field.
func TestUnguardedSchemasAreReal(t *testing.T) {
	run(t, ruleOf[Cluster](), []validate.Case{
		{Name: "connect_timeout pattern", Input: cluster("connect_timeout", "bogus"),
			WantErr: "fails to match the required pattern"},
		{Name: "lb_policy enum", Input: cluster("lb_policy", "NOPE"), WantErr: "must be one of [ROUND_ROBIN"},
		{Name: "dns_lookup_family enum", Input: cluster("dns_lookup_family", "NOPE"),
			WantErr: "must be one of [AUTO, V4_ONLY, V6_ONLY, V4_PREFERRED, ALL]"},
		{Name: "protocol_selection enum", Input: cluster("protocol_selection", "NOPE"),
			WantErr: "must be one of [USE_CONFIGURED_PROTOCOL, USE_DOWNSTREAM_PROTOCOL]"},
		{Name: "outlier_detection integer", Input: cluster("outlier_detection", map[string]any{"consecutive_5xx": 1.5}),
			WantErr: "must be an integer"},
		{Name: "circuit_breakers priority enum",
			Input:   cluster("circuit_breakers", map[string]any{"thresholds": []any{map[string]any{"priority": "NOPE"}}}),
			WantErr: "must be one of [DEFAULT, HIGH]"},
		{Name: "common_lb_config nand",
			Input: cluster("common_lb_config", map[string]any{
				"zone_aware_lb_config": map[string]any{}, "locality_weighted_lb_config": map[string]any{}}),
			// `locality_weighted_lb_config` is named by the nand but is not a KEY of the
			// object, so Joi rejects it as unknown before the dependency runs — verified
			// against the oracle.
			WantErr: `"common_lb_config.locality_weighted_lb_config" is not allowed`},
		{Name: "filters name required", Input: cluster("filters", []any{map[string]any{"typed_config": map[string]any{}}}),
			WantErr: `"filters[0].name" is required`},
		{Name: "transport_socket_matches name required",
			Input:   cluster("transport_socket_matches", []any{map[string]any{"match": map[string]any{}}}),
			WantErr: `"transport_socket_matches[0].name" is required`},
		{Name: "health_checks xor", Input: cluster("health_checks", []any{map[string]any{"timeout": "5s"}}),
			WantErr: "must contain at least one of [http_health_check"},
		{Name: "lb_subset_config enum", Input: cluster("lb_subset_config", map[string]any{"fallback_policy": "NOPE"}),
			WantErr: "must be one of [NO_FALLBACK, ANY_ENDPOINT, DEFAULT_SUBSET]"},
		{Name: "ring_hash enum", Input: cluster("ring_hash_lb_config", map[string]any{"hash_function": "NOPE"}),
			WantErr: "must be one of [XX_HASH, MURMUR_HASH_2]"},
		{Name: "eds_config xor", Input: cluster("eds_cluster_config", map[string]any{
			"eds_config": map[string]any{"path": "p", "path_config_source": map[string]any{"path": "q"}}}),
			WantErr: "contains a conflict between exclusive peers"},
		{Name: "load_balancing_policy name required", Input: cluster("load_balancing_policy", map[string]any{
			"policies": []any{map[string]any{"typed_extension_config": map[string]any{}}}}),
			WantErr: "is required"},
		{Name: "common_http_protocol_options enum", Input: cluster("common_http_protocol_options",
			map[string]any{"headers_with_underscores_action": "NOPE"}),
			WantErr: "must be one of [ALLOW, REJECT_REQUEST, DROP_HEADER]"},
		{Name: "dns_failure_refresh_rate time interval", Input: cluster("dns_failure_refresh_rate",
			map[string]any{"base_interval": "10s", "max_interval": "1s"}),
			WantErr: "max_interval must be greater than or equal to base_interval"},
		{Name: "endpoint weight min 1", Input: cluster("load_assignment", map[string]any{
			"cluster_name": "c", "endpoints": []any{map[string]any{"load_balancing_weight": 0.0}}}),
			WantErr: "must be greater than or equal to 1"},
		{Name: "lb_endpoint weight min 1", Input: cluster("load_assignment", map[string]any{
			"cluster_name": "c", "endpoints": []any{map[string]any{
				"lb_endpoints": []any{map[string]any{"load_balancing_weight": 0.0}}}}}),
			WantErr: "must be greater than or equal to 1"},
		{Name: "drop_overload category required", Input: cluster("load_assignment", map[string]any{
			"cluster_name": "c", "policy": map[string]any{"drop_overloads": []any{map[string]any{}}}}),
			WantErr: "category"},
		{Name: "rich valid cluster", Input: cluster(
			"excludedWorkloads", []any{"w"},
			"connect_timeout", "5s",
			"metadata", map[string]any{"filter_metadata": map[string]any{"a": map[string]any{"b": 1.0}}},
			"circuit_breakers", map[string]any{"thresholds": []any{map[string]any{
				"priority": "DEFAULT", "max_connections": 1.0}}},
			"load_assignment", map[string]any{"cluster_name": "c", "endpoints": []any{map[string]any{
				"locality":     map[string]any{"region": "r"},
				"lb_endpoints": []any{map[string]any{"health_status": "HEALTHY", "load_balancing_weight": 2.0}},
			}}})},
	})
}

// TestGuardBranchesDoNotLeak pins the hazard the guard is built around: the engine's
// when-merge folds the chosen branch INTO the key's base rule, so a branch carrying
// `keys` must never be merged into a base that already has a keys map — it would be
// written through to the cached rule and the next request would see the other
// branch's schema. Every guard here is based on Joi.any() (no keys) for that reason;
// this test drives ONE cached rule through both branches, repeatedly, in both orders.
func TestGuardBranchesDoNotLeak(t *testing.T) {
	r := ruleOf[Cluster]() // built once, as the engine caches it
	guarded := cplnCluster("load_assignment", map[string]any{"cluster_name": "cpln_x",
		"endpoints": []any{map[string]any{"locality": map[string]any{"region": "r"}}}})
	plain := cluster("load_assignment", map[string]any{"cluster_name": "c",
		"endpoints": []any{map[string]any{"locality": map[string]any{"region": "r"}, "priority": 0.0}}})
	plainPolicy := cluster("load_assignment", map[string]any{"cluster_name": "c",
		"policy": map[string]any{"overprovisioning_factor": 3.0}})
	guardedTimeout := cplnCluster("connect_timeout", "5s")
	plainTimeout := cluster("connect_timeout", "5s")

	for i := 0; i < 3; i++ {
		for _, c := range []validate.Case{
			{Name: "guarded endpoints", Input: guarded, WantErr: `"load_assignment.endpoints[0].locality" is not allowed`},
			{Name: "plain endpoints", Input: plain},
			{Name: "plain policy", Input: plainPolicy},
			{Name: "guarded endpoints again", Input: guarded, WantErr: "is not allowed"},
			{Name: "guarded connect_timeout", Input: guardedTimeout, WantErr: `"connect_timeout" is not allowed`},
			{Name: "plain connect_timeout", Input: plainTimeout},
		} {
			if err := c.Run(r); err != nil {
				t.Fatalf("pass %d, %s: %v", i, c.Name, err)
			}
		}
	}
}
