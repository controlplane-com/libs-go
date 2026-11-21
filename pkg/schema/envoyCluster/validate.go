package envoyCluster

// Validation for `spec.sidecar.envoy.clusters[]`. Ports nodelibs/schema/src/envoyCluster.ts.
//
// The generator emitted models.go only, and it collapsed 44 of the 48 Cluster fields
// to bare `any` because their Joi schema is a `Joi.when(...)` union rather than a
// shape. The consequence was that neither the CPLN restriction guard NOR the
// underlying schema existed: a tenant could set `connect_timeout`, `lb_policy`,
// `transport_socket`, `outlier_detection` or arbitrary `endpoints` on a Control
// Plane-owned `cpln_*` cluster and the platform accepted it.
//
// The guard is re-expressed here as a `.When("name", …)` on each guarded key, which
// is what CplnRestrictionCondition compiles to. Two of them (`load_assignment`'s
// `endpoints`/`policy`) reference `...name` — an ancestor-2 ref the engine cannot
// resolve — so those are hoisted: the whole `load_assignment` key carries the when,
// whose branches are the two complete load_assignment shapes. The ref is then a
// plain sibling of the key it is declared on, and the two branches say exactly what
// the restricted and unrestricted documents look like.

import (
	"reflect"
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// cplnNameRe is CplnRestrictionCondition's discriminator (envoyCommon.ts:9).
var cplnNameRe = regexp.MustCompile(`^cpln_.*`)

// ruleOf is the rule a Go type reflects to (its own ValidationRule when it has one).
func ruleOf[T any]() *validate.Rule { return validate.FromType(reflect.TypeOf((*T)(nil)).Elem()) }

func obj(keys ...any) *validate.Rule {
	order := make([]string, 0, len(keys)/2)
	m := make(map[string]*validate.Rule, len(keys)/2)
	for i := 0; i+1 < len(keys); i += 2 {
		k := keys[i].(string)
		order = append(order, k)
		m[k] = keys[i+1].(*validate.Rule)
	}
	return validate.ObjectOrdered(order, m)
}

// cplnGuard ports CplnRestrictionCondition(nameLocation, unrestricted, restricted?):
// a `cpln_`-prefixed cluster name selects the restricted schema (forbidden unless
// given), anything else the unrestricted one. The base is Joi.any() — as
// `Joi.when()` is — which also keeps the engine's when-merge from writing the
// branch's keys into a shared base rule.
func cplnGuard(unrestricted *validate.Rule, restricted ...*validate.Rule) *validate.Rule {
	then := validate.Any().Forbidden()
	if len(restricted) > 0 {
		then = restricted[0]
	}
	return validate.Any().When("name", validate.When{
		Is:        validate.String().Regex(cplnNameRe),
		Then:      then,
		Otherwise: unrestricted,
	})
}

func uint32Min1() *validate.Rule { return ruleOf[envoyCommon.UInt32]().Min(1) }

// --- envoyCluster.ts sub-schemas (no generated Go struct exists for any of these) ---

func healthCheckConfigRule() *validate.Rule {
	return obj(
		"port_value", validate.Number().Integer(),
		"hostname", validate.String(),
		"address", ruleOf[envoyCommon.Address](),
		"disable_active_health_check", validate.Boolean(),
	)
}

func endpointRule() *validate.Rule {
	return obj(
		"address", ruleOf[envoyCommon.Address](),
		"health_check_config", healthCheckConfigRule(),
		"hostname", validate.String(),
		"additional_addresses", validate.Array(ruleOf[envoyCommon.AdditionalAddress]()),
	)
}

func filterRule() *validate.Rule {
	return obj(
		"name", validate.String().Required(),
		"typed_config", validate.Any(),
		"config_discovery", ruleOf[envoyCommon.ExtensionConfigSource](),
	)
}

func lbEndpointRule() *validate.Rule {
	return obj(
		"endpoint", endpointRule(),
		"health_status", ruleOf[envoyCommon.HealthStatus](),
		"metadata", ruleOf[Metadata](),
		"load_balancing_weight", uint32Min1(),
	)
}

func lbEndpointRestrictedRule() *validate.Rule {
	return obj("endpoint", obj("address", obj("socket_address", ruleOf[envoyCommon.SocketAddress]())))
}

func localityLbEndpointsRule() *validate.Rule {
	return obj(
		"locality", ruleOf[envoyCommon.Locality](),
		"lb_endpoints", validate.Array(lbEndpointRule()),
		"load_balancing_weight", uint32Min1(),
		"priority", ruleOf[envoyCommon.UInt32](),
	)
}

func localityLbEndpointsRestrictedRule() *validate.Rule {
	return obj("lb_endpoints", validate.Array(lbEndpointRestrictedRule()).Max(1))
}

func policyRule() *validate.Rule {
	return obj(
		"drop_overloads", validate.Array(obj(
			"category", validate.String().Required(),
			"drop_percentage", ruleOf[envoyCommon.FractionalPercent](),
		)),
		"overprovisioning_factor", ruleOf[envoyCommon.UInt32](),
		"endpoint_stale_after", envoyCommon.DurationRule(),
		"weighted_priority_health", validate.Boolean(),
	)
}

func transportSocketMatchRule() *validate.Rule {
	return obj(
		"name", validate.String().Required(),
		"match", ruleOf[envoyCommon.Struct](),
		"transport_socket", ruleOf[envoyCommon.TransportSocket](),
	)
}

func edsClusterConfigRule() *validate.Rule {
	return obj(
		"eds_config", ruleOf[envoyCommon.ConfigSource](),
		"service_name", validate.String(),
	)
}

func circuitBreakerThresholdsRule() *validate.Rule {
	return obj(
		"priority", ruleOf[envoyCommon.RoutingPriority](),
		"max_connections", ruleOf[envoyCommon.UInt32](),
		"max_pending_requests", ruleOf[envoyCommon.UInt32](),
		"max_requests", ruleOf[envoyCommon.UInt32](),
		"retry_budget", obj(
			"budget_percent", ruleOf[envoyCommon.Percent](),
			"min_retry_concurrency", ruleOf[envoyCommon.UInt32](),
		),
		"track_remaining", validate.Boolean(),
		"max_connection_pools", ruleOf[envoyCommon.UInt32](),
	)
}

func circuitBreakersRule() *validate.Rule {
	return obj(
		"thresholds", validate.Array(circuitBreakerThresholdsRule()),
		"per_host_thresholds", validate.Array(circuitBreakerThresholdsRule()),
	)
}

func outlierDetectionRule() *validate.Rule {
	u := ruleOf[envoyCommon.UInt32]
	d := envoyCommon.DurationRule
	return obj(
		"consecutive_5xx", u(),
		"interval", d(),
		"base_ejection_time", d(),
		"max_ejection_percent", u(),
		"enforcing_consecutive5xx", u(),
		"enforcing_success_rate", u(),
		"success_rate_minimum_hosts", u(),
		"success_rate_request_volume", u(),
		"success_rate_stdev_factor", u(),
		"consecutive_gateway_failure", u(),
		"enforcing_consecutive_gateway_failure", u(),
		"split_external_local_origin_errors", validate.Boolean(),
		"consecutive_local_origin_failure", u(),
		"enforcing_consecutive_local_origin_failure", u(),
		"enforcing_local_origin_success_rate", u(),
		"failure_percentage_threshold", u(),
		"enforcing_failure_percentage", u(),
		"enforcing_failure_percentage_local_origin", u(),
		"failure_percentage_minimum_hosts", u(),
		"failure_percentage_request_volume", u(),
		"max_ejection_time", d(),
		"max_ejection_time_jitter", d(),
		"successful_active_health_check_uneject_host", validate.Boolean(),
	)
}

func lbSubsetConfigRule() *validate.Rule {
	return obj(
		"fallback_policy", validate.String().Valid("NO_FALLBACK", "ANY_ENDPOINT", "DEFAULT_SUBSET"),
		"default_subset", ruleOf[envoyCommon.Struct](),
		"locality_weight_aware", validate.Boolean(),
		"scale_locality_weight", validate.Boolean(),
		"panic_mode_any", validate.Boolean(),
		"list_as_any", validate.Boolean(),
		"metadata_fallback_policy", validate.String().Valid("METADATA_NO_FALLBACK", "FALLBACK_LIST"),
	)
}

func ringHashLbConfigRule() *validate.Rule {
	return obj(
		"minimum_ring_size", ruleOf[envoyCommon.UInt64](),
		"hash_function", validate.String().Valid("XX_HASH", "MURMUR_HASH_2"),
		"maximum_ring_size", ruleOf[envoyCommon.UInt64](),
	)
}

func leastRequestLbConfigRule() *validate.Rule {
	return obj(
		"choice_count", ruleOf[envoyCommon.UInt64](),
		"active_request_bias", ruleOf[envoyCommon.RuntimeDouble](),
		"slow_start_config", obj(
			"slow_start_window", envoyCommon.DurationRule(),
			"aggression", ruleOf[envoyCommon.RuntimeDouble](),
			"min_weight_percent", ruleOf[envoyCommon.Percent](),
		),
	)
}

func commonLbConfigRule() *validate.Rule {
	return obj(
		"healthy_panic_threshold", ruleOf[envoyCommon.Percent](),
		"zone_aware_lb_config", obj(
			"routing_enabled", ruleOf[envoyCommon.Percent](),
			"min_cluster_size", ruleOf[envoyCommon.UInt64](),
			"fail_traffic_on_panic", validate.Boolean(),
		),
		"update_merge_window", envoyCommon.DurationRule(),
		"ignore_new_hosts_until_first_hc", validate.Boolean(),
		"close_connections_on_host_set_change", validate.Boolean(),
		"consistent_hashing_lb_config", obj(
			"use_hostname_for_hashing", validate.Boolean(),
			"hash_balance_factor", ruleOf[envoyCommon.UInt32](),
		),
		"override_host_status", ruleOf[envoyCommon.HealthStatusSet](),
	).Nand("zone_aware_lb_config", "locality_weighted_lb_config")
}

// loadAssignmentRule is the load_assignment document for one branch of the guard.
func loadAssignmentRule(endpoints, policy *validate.Rule) *validate.Rule {
	return obj(
		"cluster_name", validate.String().Required(),
		"endpoints", endpoints,
		"policy", policy,
	)
}

// clusterFieldRules is the guarded schema for each Cluster key that
// CplnRestrictionCondition wraps, keyed by json name.
func clusterFieldRules() map[string]*validate.Rule {
	u := ruleOf[envoyCommon.UInt32]
	d := envoyCommon.DurationRule
	b := validate.Boolean
	return map[string]*validate.Rule{
		"transport_socket_matches": cplnGuard(validate.Array(transportSocketMatchRule())),
		"alt_stat_name":            cplnGuard(validate.String()),
		"cluster_type": cplnGuard(obj(
			"name", validate.String().Required(),
			"typed_config", validate.Any(),
		)),
		"eds_cluster_config":                cplnGuard(edsClusterConfigRule()),
		"connect_timeout":                   cplnGuard(d()),
		"per_connection_buffer_limit_bytes": cplnGuard(u()),
		"lb_policy": cplnGuard(validate.String().Valid("ROUND_ROBIN", "LEAST_REQUEST", "RING_HASH",
			"RANDOM", "MAGLEV", "CLUSTER_PROVIDED", "LOAD_BALANCING_POLICY_CONFIG")),
		"load_assignment": validate.Any().When("name", validate.When{
			Is:        validate.String().Regex(cplnNameRe),
			Then:      loadAssignmentRule(validate.Array(localityLbEndpointsRestrictedRule()).Max(1), validate.Any().Forbidden()),
			Otherwise: loadAssignmentRule(validate.Array(localityLbEndpointsRule()), policyRule()),
		}),
		"health_checks":                    cplnGuard(validate.Array(ruleOf[envoyCommon.HealthCheck]())),
		"max_requests_per_connection":      cplnGuard(u()),
		"circuit_breakers":                 cplnGuard(circuitBreakersRule()),
		"upstream_http_protocol_options":   cplnGuard(ruleOf[envoyCommon.UpstreamHttpProtocolOptions]()),
		"common_http_protocol_options":     cplnGuard(ruleOf[envoyCommon.HttpProtocolOptions]()),
		"http_protocol_options":            cplnGuard(ruleOf[envoyCommon.Http1ProtocolOptions]()),
		"http2_protocol_options":           cplnGuard(ruleOf[envoyCommon.Http2ProtocolOptions]()),
		"typed_extension_protocol_options": cplnGuard(validate.Array(validate.Object(nil).Pattern(validate.String(), validate.Any()))),
		"dns_refresh_rate":                 cplnGuard(d()),
		"dns_failure_refresh_rate":         cplnGuard(ruleOf[envoyCommon.TimeInterval]()),
		"respect_dns_ttl":                  cplnGuard(b()),
		"dns_lookup_family":                cplnGuard(validate.String().Valid("AUTO", "V4_ONLY", "V6_ONLY", "V4_PREFERRED", "ALL")),
		"dns_resolvers":                    cplnGuard(validate.Array(ruleOf[envoyCommon.Address]())),
		"use_tcp_for_dns_lookups":          cplnGuard(b()),
		"dns_resolution_config":            cplnGuard(ruleOf[envoyCommon.DnsResolutionConfig]()),
		"typed_dns_resolver_config":        cplnGuard(ruleOf[envoyCommon.TypedExtensionConfig]()),
		"wait_for_warm_on_init":            cplnGuard(b()),
		"outlier_detection":                cplnGuard(outlierDetectionRule()),
		"cleanup_interval":                 cplnGuard(d()),
		"upstream_bind_config":             cplnGuard(ruleOf[envoyCommon.BindConfig]()),
		"lb_subset_config":                 cplnGuard(lbSubsetConfigRule()),
		"ring_hash_lb_config":              cplnGuard(ringHashLbConfigRule()),
		"maglev_lb_config":                 cplnGuard(obj("table_size", ruleOf[envoyCommon.UInt64]())),
		"least_request_lb_config":          cplnGuard(leastRequestLbConfigRule()),
		"common_lb_config":                 cplnGuard(commonLbConfigRule()),
		// the one guarded key whose restricted branch is the SAME schema, not forbidden
		"transport_socket":   cplnGuard(ruleOf[envoyCommon.TransportSocket](), ruleOf[envoyCommon.TransportSocket]()),
		"metadata":           cplnGuard(ruleOf[Metadata]()),
		"protocol_selection": cplnGuard(validate.String().Valid("USE_CONFIGURED_PROTOCOL", "USE_DOWNSTREAM_PROTOCOL")),
		"upstream_connection_options": cplnGuard(obj(
			"tcp_keepalive", ruleOf[envoyCommon.TcpKeepalive](),
			"set_local_interface_name_on_upstream_connections", b(),
		)),
		"close_connections_on_host_health_failure": cplnGuard(b()),
		"ignore_health_on_host_removal":            cplnGuard(b()),
		"filters":                                  cplnGuard(validate.Array(filterRule())),
		"load_balancing_policy": cplnGuard(obj("policies", validate.Array(
			obj("typed_extension_config", ruleOf[envoyCommon.TypedExtensionConfig]())))),
		"track_timeout_budgets": cplnGuard(b()),
		"upstream_config":       cplnGuard(ruleOf[envoyCommon.TypedExtensionConfig]()),
		"track_cluster_stats": cplnGuard(obj(
			"timeout_budgets", b(),
			"request_response_sizes", b(),
			"per_endpoint_stats", b(),
		)),
		"preconnect_policy": cplnGuard(obj(
			"per_upstream_preconnect_ratio", validate.Number(),
			"predictive_preconnect_ratio", validate.Number(),
		)),
		"connection_pool_per_downstream_connection": cplnGuard(b()),
	}
}

// ValidationRule ports envoyCluster.ts's `Cluster`: the reflected struct (which
// carries `name`'s required and `type`'s enum) with every CplnRestrictionCondition
// key replaced by its guard, plus `.xor('type','cluster_type')`.
func (Cluster) ValidationRule() *validate.Rule {
	r := validate.Reflect(Cluster{})
	for k, sub := range clusterFieldRules() {
		rule := sub
		r = r.Fork(k, func(*validate.Rule) *validate.Rule { return rule })
	}
	return r.Xor("type", "cluster_type")
}
