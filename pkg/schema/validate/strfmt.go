package validate

// strfmt.go ports the string FORMATS Joi delegates to @sideway/address: .uri(),
// .domain(), .email() and .guid().
//
// The port previously approximated each with a one-line regex — `.uri()` matched a
// scheme prefix and nothing else, `.email()` was `[^\s@]+@[^\s@]+\.[^\s@]+`, `.guid()`
// insisted on the dashed unwrapped form — so Go accepted 12 of 25 invalid addresses and
// 5 of 20 invalid URIs, and rejected GUID spellings Joi takes.
//
// The grammars are transcribed from the library (address/lib/uri.js generate(),
// domain.js analyze(), email.js), with RE2's missing lookaheads written out as explicit
// checks.

import (
	"sync"

	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
)

// --- URI (RFC 3986, address/lib/uri.js) ---

// uriRegex is the default-scheme URI grammar; buildURIRegex re-generates it with a
// caller-supplied scheme alternation for Joi's `.uri({scheme:[…]})`.
var uriRegex = buildURIRegex(`[a-zA-Z][a-zA-Z\d+\-\.]*`)

func buildURIRegex(scheme string) *regexp.Regexp {
	const hexDigit = `\dA-Fa-f`
	const hexDigitOnly = `[` + hexDigit + `]`
	const unreserved = `\w\-\.~`
	const subDelims = `!\$&'\(\)\*\+,;=`
	const pctEncoded = `%` + hexDigit
	const pchar = unreserved + pctEncoded + subDelims + `:@`
	const pcharOnly = `[` + pchar + `]`
	const decOctet = `(?:0{0,2}\d|0?[1-9]\d|1\d\d|2[0-4]\d|25[0-5])`
	ipv4 := `(?:` + decOctet + `\.){3}` + decOctet

	h16 := hexDigitOnly + `{1,4}`
	ls32 := `(?:` + h16 + `:` + h16 + `|` + ipv4 + `)`
	ipv6 := `(?:` +
		`(?:` + h16 + `:){6}` + ls32 + `|` +
		`::(?:` + h16 + `:){5}` + ls32 + `|` +
		`(?:` + h16 + `)?::(?:` + h16 + `:){4}` + ls32 + `|` +
		`(?:(?:` + h16 + `:){0,1}` + h16 + `)?::(?:` + h16 + `:){3}` + ls32 + `|` +
		`(?:(?:` + h16 + `:){0,2}` + h16 + `)?::(?:` + h16 + `:){2}` + ls32 + `|` +
		`(?:(?:` + h16 + `:){0,3}` + h16 + `)?::` + h16 + `:` + ls32 + `|` +
		`(?:(?:` + h16 + `:){0,4}` + h16 + `)?::` + ls32 + `|` +
		`(?:(?:` + h16 + `:){0,5}` + h16 + `)?::` + h16 + `|` +
		`(?:(?:` + h16 + `:){0,6}` + h16 + `)?::` + `)`
	ipvFuture := `v` + hexDigitOnly + `+\.[` + unreserved + subDelims + `:]+`

	userinfo := `[` + unreserved + pctEncoded + subDelims + `:]*`
	ipLiteral := `\[(?:` + ipv6 + `|` + ipvFuture + `)\]`
	regName := `[` + unreserved + pctEncoded + subDelims + `]{1,255}`
	host := `(?:` + ipLiteral + `|` + ipv4 + `|` + regName + `)`
	authority := `(?:` + userinfo + `@)?` + host + `(?::\d*)?`

	segment := pcharOnly + `*`
	segmentNz := pcharOnly + `+`
	pathAbEmpty := `(?:/` + segment + `)*`
	pathAbsolute := `/(?:` + segmentNz + pathAbEmpty + `)?`
	pathRootless := segmentNz + pathAbEmpty
	pathAbNoAuthority := `(?:///` + segment + pathAbEmpty + `)`
	hierPart := `(?:(?://` + authority + pathAbEmpty + `)|` + pathAbsolute + `|` + pathRootless + `|` + pathAbNoAuthority + `)`

	// the `(?=#|$)` guarding query is redundant in RE2: '#' is not a pchar, so the
	// query class cannot swallow the fragment marker.
	query := `[` + pchar + `/\?]*`
	fragment := `[` + pchar + `/\?]*`
	return regexp.MustCompile(`^(?:` + scheme + `:` + hierPart + `)(?:\?` + query + `)?(?:#` + fragment + `)?$`)
}

// uriHTTPGuard is address's `(?!https?:/(?:$|[^/]))(?!https?:///)(?!https?:[^/])` — the
// three lookaheads that forbid http with an empty or malformed authority.
var uriHTTPGuard = regexp.MustCompile(`(?i)^https?:(?:/($|[^/])|///|[^/])`)

func isURI(s string) bool {
	if s == "" || uriHTTPGuard.MatchString(s) {
		return false
	}
	return uriRegex.MatchString(s)
}

// schemeURIRegexes caches the per-scheme grammars; a rule is built once at init but
// .uri({scheme}) can appear on many, and compiling this regex is not cheap.
var schemeURIRegexes sync.Map

// isURIScheme is `.uri({scheme:[…]})`: address regenerates the WHOLE grammar with the
// scheme production replaced by the alternation, so a value with the right scheme and a
// bad authority fails here too — which is why Joi reports uriCustomScheme for it.
func isURIScheme(s, alternation string) bool {
	if s == "" || uriHTTPGuard.MatchString(s) {
		return false
	}
	re, ok := schemeURIRegexes.Load(alternation)
	if !ok {
		re, _ = schemeURIRegexes.LoadOrStore(alternation, buildURIRegex(`(?:`+alternation+`)`))
	}
	return re.(*regexp.Regexp).MatchString(s)
}

// base64Regex is Joi's string.base64 pattern with its defaults (padding required, not
// url-safe) — base64.js's `internals.base64Regex[true][false]`.
var base64Regex = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

func isBase64(s string) bool { return base64Regex.MatchString(s) }

// --- domain (address/lib/domain.js analyze) ---

var (
	domainControlRx = regexp.MustCompile("[\x00-\x20@:/\\\\#!\\$&'\\(\\)\\*\\+,;=\\?]")
	domainSegmentRx = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
	domainTLDRx     = regexp.MustCompile(`^[a-zA-Z](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
)

// whatwgIDNA is the WHATWG URL standard's domain-to-ASCII, which is what
// @sideway/address reaches for a non-ASCII domain: UTS46 mapping, no STD3 rules and
// no hyphen checks, so `_` survives to be caught by domainSegmentRx below.
var whatwgIDNA = idna.New(
	idna.MapForLookup(),
	idna.StrictDomainName(false),
	idna.ValidateLabels(false),
	idna.VerifyDNSLength(false),
	idna.BidiRule(),
)

// punycodeDomain is address/lib/domain.js's internals.punycode: `new URL('http://'+d).host`,
// falling back to the input unchanged when the URL constructor throws. Without it a
// unicode domain was rejected outright — `föö.com` is valid to Joi and was not to us.
func punycodeDomain(domain string) string {
	if a, err := whatwgIDNA.ToASCII(domain); err == nil {
		return a
	}
	return domain
}

// isDomain is Domain.analyze with the library's defaults (minDomainSegments 2). tlds
// enables the IANA allow-list, which Joi turns on for .email() and for any .domain(opts)
// and leaves off for a bare .domain().
func isDomain(domain string, tlds bool) bool {
	if domain == "" || len(domain) > 256 {
		return false
	}
	if !isASCII(domain) {
		// analyze NFC-normalizes before anything else; allowUnicode defaults to true.
		domain = norm.NFC.String(domain)
	}
	if domainControlRx.MatchString(domain) {
		return false
	}
	// `%` is percent-escaped before the URL parse so a stray one does not decode.
	domain = punycodeDomain(strings.ReplaceAll(domain, "%", "%25"))
	segments := strings.Split(domain, ".")
	if len(segments) < 2 {
		return false
	}
	if tlds && !ianaTLDs[strings.ToLower(segments[len(segments)-1])] {
		return false
	}
	for i, seg := range segments {
		if seg == "" || len(seg) > 63 {
			return false
		}
		if i < len(segments)-1 {
			if !domainSegmentRx.MatchString(seg) {
				return false
			}
		} else if !domainTLDRx.MatchString(seg) {
			return false
		}
	}
	return true
}

// --- email (address/lib/email.js) ---

// emailAtextRx is RFC 5321 atext, one local-part atom.
var emailAtextRx = regexp.MustCompile(`^[\w!#\$%&'\*\+\-/=\?\^` + "`" + `\{\|\}~]+$`)

func isEmail(s string) bool {
	if s == "" {
		return false
	}
	ascii := isASCII(s)
	if !ascii {
		// email.js NFC-normalizes the WHOLE address before splitting, not just the domain.
		s = norm.NFC.String(s)
	}
	// the library measures the address in JS string length (UTF-16 code units), not bytes.
	if utf16Len(s) > 254 {
		return false
	}
	at := strings.Count(s, "@")
	if at != 1 {
		return false
	}
	i := strings.Index(s, "@")
	local, domain := s[:i], s[i+1:]
	if local == "" || len(local) > 64 { // TextEncoder length: bytes
		return false
	}
	for _, atom := range strings.Split(local, ".") {
		if atom == "" || !emailLocalAtom(atom, ascii) {
			return false
		}
	}
	// Joi's .email() defaults to tlds:{allow: <built-in list>}; .domain() does not.
	return isDomain(domain, true)
}

// emailLocalAtom is email.js internals.local for one dot-atom. RFC 6531 widens atext
// with UTF8-non-ascii, so a unicode local part (`föö@bar.com`) is valid to Joi — but
// only when the address as a whole is non-ASCII, which is the branch the library takes.
func emailLocalAtom(atom string, ascii bool) bool {
	if ascii {
		return emailAtextRx.MatchString(atom)
	}
	for _, r := range atom {
		if r >= 0x80 { // any UTF8-2/3/4 sequence; atomRx matches every one of them
			continue
		}
		if !emailAtextRx.MatchString(string(r)) {
			return false
		}
	}
	return true
}

// utf16Len is JS's String.prototype.length: astral code points count twice.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r > 0xFFFF {
			n++
		}
	}
	return n
}

// --- guid (joi string.js guid) ---

// guidVariants are Joi's guid regex with the `([:-]?)` capture expanded, since RE2 has
// no backreference for the `\2?` separators that follow.
var guidVariants = func() []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, sep := range []string{"", "-", ":"} {
		out = append(out, regexp.MustCompile(`(?i)^([\[{\(]?)[0-9A-F]{8}`+sep+
			`[0-9A-F]{4}`+optSep(sep)+`[0-9A-F]{4}`+optSep(sep)+`[0-9A-F]{4}`+optSep(sep)+
			`[0-9A-F]{12}([\]}\)]?)$`))
	}
	return out
}()

func optSep(sep string) string {
	if sep == "" {
		return ""
	}
	return sep + "?"
}

var guidBrackets = map[string]string{"{": "}", "[": "]", "(": ")", "": ""}

func isGUID(s string) bool {
	for _, re := range guidVariants {
		m := re.FindStringSubmatch(s)
		if m == nil {
			continue
		}
		if guidBrackets[m[1]] == m[len(m)-1] {
			return true
		}
	}
	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// IsDomain is isDomain for callers outside the engine — the data-service's `domain`
// path-param rule, which Joi writes as an alternatives of hostname and .domain(opts)
// and which had grown its own hand-rolled copy of this grammar.
func IsDomain(domain string, tlds bool) bool { return isDomain(domain, tlds) }
