package domain

// validate.go attaches validation to the canonical domain types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_domain.go +
// the domain entries in schema_reflect_datanet.go. domain.ts reflects from Domain
// (Reflect(Domain{})); the irreducible Joi bits are layered here:
//
//   - Route (+ RouteMirror/RouteCanary/HeaderOperation) is a PREBUILT rule
//     (RouteRule): its prefix.when('regex', {not:exist, then:default '/'}) is a
//     field-level default that must apply BEFORE the object .xor('prefix','regex'),
//     which a Refine (running AFTER xor) cannot express.
//   - ExternalPort.routes: the uniqueRoute array custom (nominal Routes slice).
//   - ExternalPortCors.maxAge: the corsMaxAge pattern (nominal CorsMaxAge string).
//   - ExternalPortCorsAllowOrigins: xor(exact, regex).
//   - DomainSpec: nand(gvcLink, workloadLink) + the domainSpecCustom cross-field
//     refine (also applies the content-ful ports default).
//   - ExternalPort: the externalPortCustom refine (443 TLS default, tcp-canary
//     rejection, route sort).
//   - Domain: the domainRootRefine (maxRoutes then domainName), in Joi order.
//
// The Port number rule is ported from port.ts (portRule); the pure map/slice value
// helpers reuse go-libs/schema/util.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	routePrefixRe = regexp.MustCompile(`^/[0-9a-zA-Z\-\._~/]*$`)
	// The JS SOURCE of the same pattern. Go's Regexp.String() renders the escapes
	// differently (`\-` vs `-`, `/` vs `\/`), and that text is what the caller sees in
	// the string.pattern.base message and in the describe document — so it is pinned to
	// the way domain.ts:57 spells it.
	routePrefixJS     = `^\/[0-9a-zA-Z-\._~\/]*$`
	routeHostPrefixRe = regexp.MustCompile(`^[0-9a-zA-Z\-\._]*$`)
	routeHostPrefixJS = `^[0-9a-zA-Z-\._]*$`
	corsMaxAgeRe      = regexp.MustCompile(`^[\d\.]+[hms]+$`)
	hostnameRe        = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
)

// ============================ top-level + nested rules ============================

// ValidationRule reflects Domain and layers the two Domain-level .custom()s
// (maxRoutes then domainName) — the schema the "domain" kind validates against.
func (Domain) ValidationRule() *validate.Rule {
	return validate.Reflect(Domain{}).CustomTyped(domainRootRefine)
}

// ValidationRule reflects DomainSpec, layers nand(gvcLink, workloadLink) and the
// domainSpecCustom cross-field refine (which also applies the ports default).
func (DomainSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(DomainSpec{}).
		Nand("gvcLink", "workloadLink"), domainSpecRefine)
}

// ValidationRule reflects ExternalPort and layers the externalPortCustom refine
// (443 TLS default, tcp-canary rejection, route sort).
func (ExternalPort) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(ExternalPort{}), externalPortRefine)
}

// ValidationRule reflects the cors allowOrigins entry and layers xor(exact, regex).
func (ExternalPortCorsAllowOrigins) ValidationRule() *validate.Rule {
	return validate.Reflect(ExternalPortCorsAllowOrigins{}).Xor("exact", "regex")
}

// ============================ nominal leaf / collection types =====================

func (Routes) ValidationRule() *validate.Rule {
	return validate.Array(RouteRule()).CustomTyped(validate.CustomT(uniqueRoutes))
}

func (CorsMaxAge) ValidationRule() *validate.Rule { return validate.String().Regex(corsMaxAgeRe) }

// ValidationRule ports the lowercase cors allow/expose header element type.
func (LowerHeader) ValidationRule() *validate.Rule { return validate.String().Lowercase() }

func (GvcLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("gvc"))
}

func (CertificateGeneratorLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("certificategenerator"))
}

func (WorkloadLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("!workload"))
}

func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

// ============================ prebuilt Route rule =================================

var (
	allowedHeaderWildcards = map[string]bool{
		"%REQUESTED_SERVER_NAME%": true, "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%": true, "%START_TIME%": true,
	}
	headerWildcardRe = regexp.MustCompile(`%([^%]+)%`)
)

// EnvoyHeaderValueRule ports domain.ts:13: only the three allowed %WILDCARD%s.
func EnvoyHeaderValueRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		for _, m := range headerWildcardRe.FindAllString(s, -1) {
			if !allowedHeaderWildcards[m] {
				return nil, h.Errorf("any.custom", "Invalid wildcard in header value: %s. Allowed wildcards are: %%REQUESTED_SERVER_NAME%%, %%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%%, %%START_TIME%%", m)
			}
		}
		return s, nil
	})
}

// ValidationRule makes the canonical EnvoyHeaderValue type self-validating.
func (EnvoyHeaderValue) ValidationRule() *validate.Rule { return EnvoyHeaderValueRule() }

// headerOperationRule ports domain.ts:28.
func headerOperationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"set": validate.Object(nil).Pattern(validate.String(), EnvoyHeaderValueRule()),
	})
}

// RouteMirrorRule ports domain.ts:34.
func RouteMirrorRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"workloadLink": base.LocalLinkRule().Custom(base.ToKind("!workload")).Required(),
		"port":         portRule().Optional(),
		"percent":      validate.Number().Min(0).Max(100).Required(),
	})
}

// RouteCanaryRule ports domain.ts:42.
func RouteCanaryRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"workloadLink": base.LocalLinkRule().Custom(base.ToKind("!workload")).Required(),
		"port":         portRule().Optional(),
		"weight":       validate.Number().Integer().Min(0).Max(100).Required(),
	})
}

// routeCanaryWeight ports the Route .custom (domain.ts:101-122): combined canary
// weight must not exceed 100.
func routeCanaryWeight(r *Route, h *validate.Helper) error {
	total := 0.0
	for _, c := range r.Canaries {
		total += c.Weight
	}
	if len(r.Canaries) > 0 && total > 100 {
		return h.Errorf("any.custom", "the combined weight of all canaries on a route must not exceed 100.")
	}
	return nil
}

// RouteRule ports domain.ts:54. Prebuilt (not reflected) for the prefix `.when`.
func RouteRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"prefix": validate.String().
			// `not: Joi.exist()` — exist() is any().required(), and the condition schema's
			// presence is what decides whether an ABSENT sibling matches. A bare Any()
			// matches an absent regex, so the Not branch stopped firing and the default
			// never applied, leaving xor('prefix','regex') to reject a route with neither.
			When("regex", validate.When{Not: validate.Any().Required(), Then: validate.String().Default("/")}).
			RegexAs(routePrefixRe, routePrefixJS),
		"replacePrefix": validate.String().RegexAs(routePrefixRe, routePrefixJS).Optional(),
		"regex":         base.RegexRule().Optional(),
		"caseInsensitive": validate.Boolean().
			When("regex", validate.When{Is: validate.Any().Required(), Then: validate.Any().Forbidden().Messages(map[string]string{
				"any.unknown": "caseInsensitive applies to prefix only; use the (?i) flag in the regex instead",
			})}).
			Optional(),
		"workloadLink": base.LocalLinkRule().Custom(base.ToKind("!workload")).Required(),
		"port":         portRule().Optional(),
		"hostPrefix":   validate.String().RegexAs(routeHostPrefixRe, routeHostPrefixJS).Optional(),
		"hostRegex":    base.RegexRule().Optional(),
		"headers":      validate.Object(map[string]*validate.Rule{"request": headerOperationRule()}).Optional(),
		"replica":      validate.Number().Integer().Min(0).Optional(),
		"mirror":       validate.Array(RouteMirrorRule().Optional()),
		"canaries":     validate.Array(RouteCanaryRule()).Optional(),
	}).Xor("prefix", "regex").Nand("hostPrefix", "hostRegex").CustomTyped(validate.CustomT(routeCanaryWeight))
}

// ValidationRule makes the canonical Route type self-validating (prebuilt rule).
func (Route) ValidationRule() *validate.Rule { return RouteRule() }

// ============================ Port number rule (port.ts) ==========================

// wlReservedPorts ports the reserved-port list in port.ts.
var wlReservedPorts = []int{8012, 8022, 9090, 9091, 15000, 15001, 15006, 15020, 15021, 15090, 41000}

// portRule ports `Port = Joi.number().integer().max(65535).min(80).custom(reserved)`.
func portRule() *validate.Rule {
	return validate.Number().Integer().Max(65535).Min(80).Custom(func(value any, h *validate.Helper) (any, error) {
		f, _ := value.(float64)
		for _, r := range wlReservedPorts {
			if float64(r) == f {
				return nil, h.Errorf("any.custom", "The port %v is reserved for system use. Please choose a different port. Reserved ports include: %s", int(f), reservedPortsCSV())
			}
		}
		return value, nil
	})
}

func reservedPortsCSV() string {
	parts := make([]string, len(wlReservedPorts))
	for i, r := range wlReservedPorts {
		parts[i] = strconv.Itoa(r)
	}
	return strings.Join(parts, ",")
}

// ============================ array / object customs =============================

// uniqueRoutes ports domain.ts:565-597, in its order: an exact duplicate anywhere wins
// over a shared key, and the shared key reported is the first group (by first
// appearance) to hold two routes, not the first route that completes one.
func uniqueRoutes(routes *Routes, h *validate.Helper) error {
	seen := map[string]bool{}
	for _, r := range *routes {
		key := canonical(r)
		if seen[key] {
			return h.Errorf("any.custom", "Routes must be unique")
		}
		seen[key] = true
	}

	if g, n := firstSharedGroup(*routes, func(r Route) string {
		route := r.Prefix
		if route == "" {
			route = r.Regex
		}
		return route + ";" + routeHost(r)
	}); n > 1 {
		route, host, _ := strings.Cut(g, ";")
		hostPart := ""
		if host != "" {
			hostPart = ";" + host
		}
		return h.Errorf("any.custom", "There are more than one routes for the prefix %s%s", route, hostPart)
	}

	// prefixes that differ only by case collide once one of them matches case-insensitively
	// (master a7f9298a2)
	folded := map[string][]Route{}
	var order []string
	for _, r := range *routes {
		if r.Prefix == "" {
			continue
		}
		k := strings.ToLower(r.Prefix) + ";" + routeHost(r)
		if _, ok := folded[k]; !ok {
			order = append(order, k)
		}
		folded[k] = append(folded[k], r)
	}
	for _, k := range order {
		group := folded[k]
		if len(group) < 2 {
			continue
		}
		for _, r := range group {
			if r.CaseInsensitive != nil && *r.CaseInsensitive {
				return h.Errorf("any.custom", "The case insensitive route %s overlaps with another route that differs only by case", r.Prefix)
			}
		}
	}
	return nil
}

func routeHost(r Route) string {
	if r.HostPrefix != "" {
		return r.HostPrefix
	}
	return r.HostRegex
}

// firstSharedGroup groups routes by key in order of first appearance — lodash groupBy's
// key order — and returns the first group holding more than one route.
func firstSharedGroup(routes Routes, key func(Route) string) (string, int) {
	counts := map[string]int{}
	var order []string
	for _, r := range routes {
		k := key(r)
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	for _, k := range order {
		if counts[k] > 1 {
			return k, counts[k]
		}
	}
	return "", 0
}

// externalPortRefine ports the ExternalPort .custom (domain.ts:226-266): default TLS
// on 443 non-tcp, tcp-canary rejection, and route sorting.
func externalPortRefine(p *ExternalPort, h *validate.Helper) error {
	if util.Float64(p.Number) == 443 && p.Tls == nil && p.Protocol != "tcp" {
		p.Tls = getDefaultExternalPortTLS()
	}
	// No canary/protocol gate: canaries are allowed on tcp ports too, where the
	// weight applies per connection rather than per request (domain.ts, 4a8fcf74b).
	if len(p.Routes) > 0 {
		noRegex := true
		for _, r := range p.Routes {
			if r.Regex != "" {
				noRegex = false
				break
			}
		}
		if noRegex {
			sortDomainRoutes(p.Routes)
		}
	}
	return nil
}

// sortDomainRoutes ports the ExternalPort route sort (domain.ts:242-262).
func sortDomainRoutes(routes []Route) {
	for i := 1; i < len(routes); i++ {
		for j := i; j > 0; j-- {
			if domainRouteLess(&routes[j-1], &routes[j]) {
				routes[j-1], routes[j] = routes[j], routes[j-1]
			} else {
				break
			}
		}
	}
}

// domainRouteLess reports whether route b should sort before route a (i.e. swap).
func domainRouteLess(a, b *Route) bool {
	if len(a.Prefix) != len(b.Prefix) {
		return len(b.Prefix) > len(a.Prefix)
	}
	if len(a.HostPrefix) != len(b.HostPrefix) {
		return len(b.HostPrefix) > len(a.HostPrefix)
	}
	if strings.HasSuffix(a.Prefix, "/") {
		return false
	}
	return strings.HasSuffix(b.Prefix, "/")
}

// domainSpecRefine ports the DomainSpec .custom (domain.ts:317-393). Applies the
// content-ful ports default, dnsMode/cert rules, cipher migration and workloadLink
// route constraints.
func domainSpecRefine(d *DomainSpec, h *validate.Helper) error {
	rawDnsMode := string(d.DnsMode)
	if len(d.Ports) < 1 {
		d.Ports = []ExternalPort{getDefaultExternalPort()}
	}
	ports := d.Ports
	gvcLink := string(d.GvcLink)
	certChallengeType := string(d.CertChallengeType)

	if rawDnsMode == "cname" && gvcLink != "" && certChallengeType == "http01" {
		for i := range ports {
			if ports[i].Tls == nil {
				continue
			}
			if portSecretLink(ports[i].Tls) == "" {
				return h.Errorf("any.custom", "When dnsMode is set to cname the domain must not use subdomain based routing or must have a tls.serverCertificate.secretLink configured for all ports that enables tls.")
			}
		}
	}
	if rawDnsMode == "ns" && certChallengeType == "http01" {
		return h.Errorf("any.custom", "http01 challenge type is not supported for domains with dnsMode ns.")
	}

	// certificateGeneratorLink exclusions (domain.ts:339-347, master 0c0ece779): the
	// generator defines the challenge type and issues the certificate, so neither
	// certChallengeType nor a per-port serverCertificate may be set alongside it.
	if d.CertificateGeneratorLink != "" {
		if certChallengeType != "" {
			return h.Errorf("any.custom", "certChallengeType must not be set when certificateGeneratorLink is configured; the generator defines the challenge type.")
		}
		for i := range ports {
			if portSecretLink(ports[i].Tls) != "" {
				return h.Errorf("any.custom", "certificateGeneratorLink cannot be combined with tls.serverCertificate.secretLink.")
			}
		}
	}

	// cipher migration (domain.ts:342-351).
	for i := range ports {
		tls := ports[i].Tls
		if tls == nil || len(tls.CipherSuites) == 0 {
			continue
		}
		migrated := make([]ExternalPortTlsCipherSuites, 0, len(tls.CipherSuites))
		seen := map[string]bool{}
		for _, c := range tls.CipherSuites {
			nc := oldNewCipherMap[string(c)]
			if !seen[nc] {
				seen[nc] = true
				migrated = append(migrated, ExternalPortTlsCipherSuites(nc))
			}
		}
		tls.CipherSuites = migrated
	}

	workloadLink := string(d.WorkloadLink)
	if workloadLink != "" {
		if len(ports) == 0 {
			return h.Errorf("any.custom", "when workloadLink is configured, at least one port must be configured.")
		}
		for i := range ports {
			routes := ports[i].Routes
			if routes == nil {
				return h.Errorf("any.custom", "port.routes is not iterable")
			}
			if len(routes) != 1 {
				return h.Errorf("any.custom", "when workloadLink is configured, every port must have exactly ONE route.")
			}
			for _, rt := range routes {
				if rl := rt.WorkloadLink; rl != "" && rl != workloadLink {
					return h.Errorf("any.custom", "when workloadLink is configured for the domain, no route can reference another workload.")
				}
			}
		}
		if certChallengeType == "http01" {
			return h.Errorf("any.custom", "when workloadLink is configured, certChallengeType can not be http01.")
		}
	}

	if util.Bool(d.AcceptAllSubdomains) && util.Bool(d.AcceptAllHosts) {
		return h.Errorf("any.custom", "only one of acceptAllSubdomains and acceptAllHosts can be enabled.")
	}
	if !util.Bool(d.AcceptAllSubdomains) && !util.Bool(d.AcceptAllHosts) {
		for i := range ports {
			for _, rt := range ports[i].Routes {
				if rt.HostPrefix != "" || rt.HostRegex != "" {
					return h.Errorf("any.custom", "hostPrefix or hostRegex can only be used if either acceptAllSubdomains or acceptAllHosts is set to true in domain spec.")
				}
			}
		}
	}
	return nil
}

// maxRoutes ports domain.ts:574: per-port route count cap (150, or 200 with the
// cpln/routeLimitOverride tag).
func maxRoutes(v any, h *validate.Helper) error {
	maxR := 150
	if tags, _ := validate.Get(v, "tags").(map[string]any); tags != nil {
		if _, ok := tags["cpln/routeLimitOverride"]; ok {
			maxR = 200
		}
	}
	if spec := validate.Sub[DomainSpec](v, "spec"); spec != nil {
		for i := range spec.Ports {
			if len(spec.Ports[i].Routes) > maxR {
				return h.Errorf("any.custom", "number of routes exceeds maximum of %d", maxR)
			}
		}
	}
	return nil
}

// domainName ports the Domain .custom (domain.ts:462-541): name validation,
// single-GVC routing, .internal rules, dnsMode defaulting, apex-cname.
func domainName(v any, h *validate.Helper) error {
	name := validate.Str(v, "name")
	internal := isInternalDomain(name)

	if internal {
		if !hostnameRe.MatchString(name) || len(name) > 253 {
			return h.Errorf("any.custom", "invalid .internal domain name: %s", name)
		}
	} else if !isValidPublicDomain(name) {
		// domain.ts:462-466 validates a non-.internal name with
		// `Joi.string().domain({tlds:{allow:true}})`. Only "contains a dot" was
		// checked here, so `foo.notatld`, `FOO.COM`, `foo.c` and `1.2.3.4` were all
		// accepted and persisted as domains that can never resolve.
		return h.Errorf("any.custom", "%q must contain a valid domain name", "value")
	}

	spec := validate.Sub[DomainSpec](v, "spec")

	if spec != nil {
		if spec.GvcLink != "" {
			for i := range spec.Ports {
				if len(spec.Ports[i].Routes) > 0 {
					return h.Errorf("any.custom", "Only one of gvcLink or ports.routes may be configured.")
				}
			}
		}
		for i := range spec.Ports {
			routes := spec.Ports[i].Routes
			if routes == nil {
				continue
			}
			var gvcs []string
			for _, rt := range routes {
				gvcs = append(gvcs, rt.WorkloadLink)
				for _, c := range rt.Canaries {
					gvcs = append(gvcs, c.WorkloadLink)
				}
			}
			if msg := sameGvcCheck(gvcs); msg != "" {
				return h.Errorf("any.custom", "%s", msg)
			}
		}
	}

	if spec == nil {
		spec = &DomainSpec{}
		validate.Set(v, "spec", spec)
		spec = validate.Sub[DomainSpec](v, "spec")
	}
	if internal {
		if spec.DnsMode == "" {
			spec.DnsMode = "cname"
		}
		if spec.DnsMode != "cname" {
			return h.Errorf("any.custom", "Only 'cname' dnsMode is supported for .internal domains.")
		}
		if spec.GvcLink != "" {
			return h.Errorf("any.custom", "gvcLink is not supported for .internal domains.")
		}
		if spec.CertChallengeType != "" {
			return h.Errorf("any.custom", "certChallengeType is not supported for .internal domains; provide tls.serverCertificate.secretLink instead.")
		}
		if spec.CertificateGeneratorLink != "" {
			return h.Errorf("any.custom", "certificateGeneratorLink is not supported for .internal domains; provide tls.serverCertificate.secretLink instead.")
		}
		for i := range spec.Ports {
			port := &spec.Ports[i]
			if port.Tls != nil && portSecretLink(port.Tls) == "" && len(port.Routes) > 0 {
				return h.Errorf("any.custom", ".internal domains require tls.serverCertificate.secretLink on every port that enables tls.")
			}
		}
		return nil
	}

	if spec.DnsMode == "" {
		spec.DnsMode = "cname"
		if spec.GvcLink != "" {
			spec.DnsMode = "ns"
		}
	}
	apex, sub, err := ParseHost(name)
	if err != nil {
		return h.Errorf("any.custom", "%s", err.Error())
	}
	if sub == "" && apex == name && spec.DnsMode != "cname" {
		return h.Errorf("any.custom", "cname is the only valid dnsMode for apex domain %s", name)
	}
	return nil
}

// domainRootRefine combines the two Domain-level .custom()s (maxRoutes then
// domainName), in Joi order.
func domainRootRefine(v any, h *validate.Helper) (any, error) {
	if err := maxRoutes(v, h); err != nil {
		return nil, err
	}
	if err := domainName(v, h); err != nil {
		return nil, err
	}
	return v, nil
}

// isValidPublicDomain ports `Joi.string().domain({ tlds: { allow: true } })`.
//
// It delegates to the engine, which carries the ONE port of
// @sideway/address's domain.analyze — NFC normalization, the WHATWG punycode step and
// the bundled IANA list. This was a second hand-rolled copy of that grammar, and the
// copy drifted: it rejected an uppercase TLD (`FOO.COM` is valid to Joi, whose URL step
// lowercases the whole host) and it swapped the IANA list for the Public Suffix List,
// which answers a different question.
func isValidPublicDomain(name string) bool { return validate.IsDomain(name, true) }

// ============================ pure value helpers =================================

// defaultCiphers is the getDefaultExternalPortTLS cipher set (domain.ts:125-134).
var defaultCiphers = []string{
	"ECDHE-ECDSA-AES256-GCM-SHA384",
	"ECDHE-ECDSA-CHACHA20-POLY1305",
	"ECDHE-ECDSA-AES128-GCM-SHA256",
	"ECDHE-RSA-AES256-GCM-SHA384",
	"ECDHE-RSA-CHACHA20-POLY1305",
	"ECDHE-RSA-AES128-GCM-SHA256",
	"AES256-GCM-SHA384",
	"AES128-GCM-SHA256",
}

// oldNewCipherMap ports domain.ts:153 (legacy → current cipher).
var oldNewCipherMap = map[string]string{
	"TLS_RSA_WITH_AES_256_GCM_SHA384":             "ECDHE-RSA-AES256-GCM-SHA384",
	"TLS_RSA_WITH_AES_128_GCM_SHA256":             "ECDHE-RSA-AES128-GCM-SHA256",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256": "ECDHE-RSA-CHACHA20-POLY1305",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":       "ECDHE-RSA-AES256-GCM-SHA384",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":       "ECDHE-RSA-AES128-GCM-SHA256",
	"TLS_CHACHA20_POLY1305_SHA256":                "ECDHE-RSA-CHACHA20-POLY1305",
	"TLS_AES_256_GCM_SHA384":                      "AES256-GCM-SHA384",
	"TLS_AES_128_GCM_SHA256":                      "AES128-GCM-SHA256",
	"ECDHE-ECDSA-AES256-GCM-SHA384":               "ECDHE-ECDSA-AES256-GCM-SHA384",
	"ECDHE-ECDSA-CHACHA20-POLY1305":               "ECDHE-ECDSA-CHACHA20-POLY1305",
	"ECDHE-ECDSA-AES128-GCM-SHA256":               "ECDHE-ECDSA-AES128-GCM-SHA256",
	"ECDHE-RSA-AES256-GCM-SHA384":                 "ECDHE-RSA-AES256-GCM-SHA384",
	"ECDHE-RSA-CHACHA20-POLY1305":                 "ECDHE-RSA-CHACHA20-POLY1305",
	"ECDHE-RSA-AES128-GCM-SHA256":                 "ECDHE-RSA-AES128-GCM-SHA256",
	"AES256-GCM-SHA384":                           "AES256-GCM-SHA384",
	"AES128-GCM-SHA256":                           "AES128-GCM-SHA256",
	"DES-CBC3-SHA":                                "DES-CBC3-SHA",
	"ECDHE-RSA-AES128-SHA":                        "ECDHE-RSA-AES128-SHA",
	"ECDHE-RSA-AES256-SHA":                        "ECDHE-RSA-AES256-SHA",
	"AES128-SHA":                                  "AES128-SHA",
	"AES256-SHA":                                  "AES256-SHA",
}

// getDefaultExternalPortTLS ports domain.ts:269.
func getDefaultExternalPortTLS() *ExternalPortTls {
	suites := make([]ExternalPortTlsCipherSuites, len(defaultCiphers))
	for i, c := range defaultCiphers {
		suites[i] = ExternalPortTlsCipherSuites(c)
	}
	return &ExternalPortTls{MinProtocolVersion: "TLSV1_2", CipherSuites: suites}
}

// getDefaultExternalPort builds the default external port (domain.ts:269-293):
// 443/http2 with an empty route list and TLSv1.2 + the default cipher suites.
func getDefaultExternalPort() ExternalPort {
	return ExternalPort{Number: util.Ptr(443.0), Protocol: "http2", Routes: Routes{}, Tls: getDefaultExternalPortTLS()}
}

// portSecretLink returns tls.serverCertificate.secretLink (empty when absent).
func portSecretLink(tls *ExternalPortTls) string {
	if tls == nil || tls.ServerCertificate == nil {
		return ""
	}
	return string(tls.ServerCertificate.SecretLink)
}

// sameGvcCheck ports domain.ts sameGvc: the route/canary workloadLinks must resolve
// to a single GVC (the gvc segment is getPart(link,3)). Returns "" when OK.
func sameGvcCheck(links []string) string {
	seen := map[string]bool{}
	for _, s := range links {
		seen[gvcSegmentOf(s)] = true
	}
	if len(seen) > 1 {
		return "Domains may only route to Workloads in a single GVC."
	}
	return ""
}

// gvcSegmentOf returns the gvc segment (index 4) of a `/org/{org}/gvc/{gvc}/...` link.
func gvcSegmentOf(link string) string {
	comps := strings.Split(link, "/")
	if len(comps) > 4 {
		return comps[4]
	}
	return ""
}

func isInternalDomain(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".internal")
}

// internalApex is the rightmost two labels of an .internal name (domain.ts:451).
func internalApex(name string) string {
	labels := strings.Split(name, ".")
	if len(labels) <= 2 {
		return name
	}
	return strings.Join(labels[len(labels)-2:], ".")
}

// ParseHost splits a hostname into its registrable apex (eTLD+1) and the subdomain
// labels below it, mirroring tld-extract's {domain, sub}.
//
// tld-extract THROWS rather than falling back when the name's suffix is not in the
// ICANN list (`EXAMPLE.COM` — the lookup is case-sensitive and `COM` is not a key) or
// when the name IS exactly a public suffix (`co.uk`). Every TS call site lets that
// throw out, so it is a 400; the port swallowed it and created the domain. The error
// is the JS message verbatim, because it reaches the client.
func ParseHost(name string) (apex, sub string, err error) {
	apex, sub = parseHost(name)
	if isInternalDomain(name) {
		// .internal is not in the Public Suffix List (schema/domain.ts:436), so
		// tld-extract is never asked about it.
		return apex, sub, nil
	}
	parts := strings.Split(name, ".")
	suffix, icann := publicsuffix.PublicSuffix(name)
	level := -1
	if icann {
		level = strings.Count(suffix, ".") + 1
	}
	if len(parts) <= level || level == -1 {
		return apex, sub, fmt.Errorf(`Invalid TLD {"parts":%s,"tld_level":%d,"allowUnknownTLD":false}`,
			jsonStrings(parts), level)
	}
	return apex, sub, nil
}

// jsonStrings renders a []string as a JSON array, for the tld-extract message.
func jsonStrings(ss []string) string {
	b, _ := json.Marshal(ss)
	return string(b)
}

// parseHost splits a hostname into its registrable apex (eTLD+1) and the subdomain
// labels below it. On a parse failure the whole name is treated as the apex.
func parseHost(name string) (apex, sub string) {
	if isInternalDomain(name) {
		apex = internalApex(name)
	} else if e, err := publicsuffix.EffectiveTLDPlusOne(name); err == nil {
		apex = e
	} else {
		apex = name
	}
	if name == apex {
		return apex, ""
	}
	sub = strings.TrimSuffix(name, "."+apex)
	if sub == name { // apex not a suffix (shouldn't happen) => no sub
		sub = ""
	}
	return apex, sub
}

// canonical is the stable JSON encoding of a value, used as a dedupe key.
func canonical(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
