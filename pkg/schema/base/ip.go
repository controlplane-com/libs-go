package base

// ip.go is the single home of Joi's `Joi.string().ip({version, cidr})`. Every
// caller used to hand-roll it (net.ParseIP + a dot count, or net.ParseCIDR), which
// diverged from Joi at both edges: net.ParseIP ACCEPTS an IPv4-mapped IPv6 literal
// and REJECTS leading-zero octets, where the RFC-3986 grammar Joi compiles does the
// opposite. The regex sources below are @hapi/address's, verbatim (they are RE2-safe:
// no lookaround, no backreferences), so the accept set is Joi's by construction.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// rfc3986 pieces, copied from @hapi/address dist/uri.js `generate()`.
const (
	ipv4Src = `(?:(?:0{0,2}\d|0?[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.){3}(?:0{0,2}\d|0?[1-9]\d|1\d\d|2[0-4]\d|25[0-5])`
	v4Cidr  = `(?:\d|[1-2]\d|3[0-2])`
	v6Cidr  = `(?:0{0,2}\d|0?[1-9]\d|1[01]\d|12[0-8])`
	ipvFutS = `v[\dA-Fa-f]+\.[\w\-\.~!\$&'\(\)\*\+,;=:]+`
)

// ipv6Src is @hapi/address's rfc3986.ipv6address. h16/ls32 expanded exactly as the
// module generates them.
var ipv6Src = func() string {
	h16 := `[\dA-Fa-f]{1,4}`
	ls32 := `(?:` + h16 + `:` + h16 + `|` + ipv4Src + `)`
	return `(?:(?:` + h16 + `:){6}` + ls32 +
		`|::(?:` + h16 + `:){5}` + ls32 +
		`|(?:` + h16 + `)?::(?:` + h16 + `:){4}` + ls32 +
		`|(?:(?:` + h16 + `:){0,1}` + h16 + `)?::(?:` + h16 + `:){3}` + ls32 +
		`|(?:(?:` + h16 + `:){0,2}` + h16 + `)?::(?:` + h16 + `:){2}` + ls32 +
		`|(?:(?:` + h16 + `:){0,3}` + h16 + `)?::` + h16 + `:` + ls32 +
		`|(?:(?:` + h16 + `:){0,4}` + h16 + `)?::` + ls32 +
		`|(?:(?:` + h16 + `:){0,5}` + h16 + `)?::` + h16 +
		`|(?:(?:` + h16 + `:){0,6}` + h16 + `)?::)`
}()

func ipVersionSrc(v string) string {
	switch v {
	case "ipv4":
		return ipv4Src
	case "ipv6":
		return ipv6Src
	default: // ipvfuture
		return ipvFutS
	}
}

// ipRegexFor ports @hapi/address ipRegex(): one alternation per requested version,
// each carrying the version's CIDR suffix per the cidr mode.
func ipRegexFor(versions []string, cidr string) *regexp.Regexp {
	vs := versions
	if len(vs) == 0 {
		vs = []string{"ipv4", "ipv6", "ipvfuture"}
	}
	parts := make([]string, 0, len(vs))
	for _, v := range vs {
		ip := ipVersionSrc(v)
		if cidr == "forbidden" {
			parts = append(parts, ip)
			continue
		}
		suffix := v6Cidr
		if v == "ipv4" {
			suffix = v4Cidr
		}
		cidrPart := `\/` + suffix
		if cidr == "required" {
			parts = append(parts, ip+cidrPart)
			continue
		}
		parts = append(parts, ip+`(?:`+cidrPart+`)?`)
	}
	return regexp.MustCompile(`^(?:` + strings.Join(parts, "|") + `)$`)
}

var ipRegexCache = map[string]*regexp.Regexp{}

func ipRegexCached(versions []string, cidr string) *regexp.Regexp {
	key := cidr + "|" + strings.Join(versions, ",")
	if re, ok := ipRegexCache[key]; ok {
		return re
	}
	re := ipRegexFor(versions, cidr)
	ipRegexCache[key] = re
	return re
}

// IPRule ports Joi.string().ip({version, cidr}). versions may be empty (Joi's
// "all versions" default, which reports string.ip rather than string.ipVersion);
// cidr is "optional" (Joi's default) | "required" | "forbidden".
//
// The error is Joi's, down to the context: string.ipVersion carries {cidr, version}
// and string.ip carries {cidr}; the label — not the value — is what the message
// quotes, so the caller can see WHICH field is wrong.
func IPRule(versions []string, cidr string) *validate.Rule {
	if cidr == "" {
		cidr = "optional"
	}
	re := ipRegexCached(versions, cidr)
	vlist := append([]string(nil), versions...)
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if re.MatchString(s) {
			return s, nil
		}
		if len(vlist) > 0 {
			return nil, h.WithCtx(
				h.Errorf("string.ipVersion", "%q must be a valid ip address of one of the following versions [%s] with a %s CIDR",
					h.Label(), strings.Join(vlist, ", "), cidr),
				"cidr", cidr, "version", vlist)
		}
		return nil, h.WithCtx(
			h.Errorf("string.ip", "%q must be a valid ip address with a %s CIDR", h.Label(), cidr),
			"cidr", cidr)
	})
}
