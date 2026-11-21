package base

// hostname.go ports Joi's `Joi.string().hostname()`, which is
// `Domain.isValid(value, {minDomainSegments: 1}) || <ip>` (joi lib/types/string.js:391).
// It lived twice before — once in orgLogging as an `any.custom` and once in
// mk8sCommon quoting the VALUE instead of the label — so the same field answered
// with two different error shapes depending on which package declared it.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"golang.org/x/net/idna"
)

// The three segment regexes of @hapi/address dist/domain.js, verbatim. Note the LAST
// segment (the TLD) must START with a letter, which an RFC-1123 regex does not require
// — "a.1b" is a hostname to Go's old check and not to Joi's.
var (
	domainControlRe   = regexp.MustCompile(`[\x00-\x20@:/\\#!$&'()*+,;=?]`)
	tldSegmentRe      = regexp.MustCompile(`^[a-zA-Z](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
	domainSegmentRe   = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
	hostnameIPRe      = ipRegexFor(nil, "forbidden") // joi internals.ipRegex
	domainPunyProfile = idna.New(idna.MapForLookup(), idna.Transitional(false))
)

// DomainIsValid ports @hapi/address analyzeDomain (the no-tlds path): the shared
// predicate behind Joi's string.domain and string.hostname. minSegments is Joi's
// minDomainSegments (2 by default, 1 for hostname()).
func DomainIsValid(domain string, minSegments int) bool {
	if domain == "" {
		return false
	}
	if validate.JSLen(domain) > 256 {
		return false
	}
	if domainControlRe.MatchString(domain) {
		return false
	}
	// `new URL('http://'+domain).host` — lowercases and punycodes; a value the URL
	// parser rejects is analyzed as-is.
	if ascii, err := domainPunyProfile.ToASCII(domain); err == nil && ascii != "" {
		domain = strings.ToLower(ascii)
	}
	segments := strings.Split(domain, ".")
	if len(segments) < minSegments {
		return false
	}
	for i, seg := range segments {
		if seg == "" || len(seg) > 63 {
			return false
		}
		if i < len(segments)-1 {
			if !domainSegmentRe.MatchString(seg) {
				return false
			}
		} else if !tldSegmentRe.MatchString(seg) {
			return false
		}
	}
	return true
}

// HostnameRule ports Joi.string().hostname(): a domain with at least one segment, or
// an IP literal. The error is Joi's `string.hostname`, quoting the LABEL.
func HostnameRule() *validate.Rule {
	return validate.String().Custom(HostnameCustom)
}

// HostnameCustom is HostnameRule's check, exported for the leaf types that chain a
// second custom after it (orgLogging's EsAmazonHost).
func HostnameCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if DomainIsValid(s, 1) || hostnameIPRe.MatchString(s) {
		return s, nil
	}
	// WithCtx seeds the context so the engine's one value-stamping seam can add
	// `value` — an error left with a nil context never gets it.
	return nil, h.WithCtx(h.Errorf("string.hostname", "%q must be a valid hostname", h.Label()))
}
