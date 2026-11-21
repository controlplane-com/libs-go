package base

// linkalgebra.go is the PURE link expansion + kind-matching algebra ported from
// data-service/pkg/handlers/links.go — the Go equivalent of base.ts's LocalLink /
// ImageLink `.custom` expansion and toKind validation. It takes org/gvc as
// parameters (no request context) and returns a plain error, so it can live in
// go-libs beside the canonical structs and be reused by the LocalLink/ImageLink
// ValidationRule methods. The data-service central link pass (links.go) keeps its
// own httpx-flavoured copy until the migration retires it.

import (
	"errors"
	"strings"
)

const internalLinkPrefix = "cpln://internal/"

// internalAliases mirrors internal.ts's registry (the only exposed alias is keda).
var internalAliases = map[string]bool{"keda": true}

// IsInternalLink reports a cpln://internal/<alias> link.
func IsInternalLink(v string) bool { return strings.HasPrefix(v, internalLinkPrefix) }

// validInternalLink ports parseInternalLink: the alias after the prefix must be a
// single registered segment.
func validInternalLink(v string) bool {
	if !IsInternalLink(v) {
		return false
	}
	alias := v[len(internalLinkPrefix):]
	if alias == "" || strings.Contains(alias, "/") {
		return false
	}
	return internalAliases[alias]
}

// isGvcScopedKind reports whether a bare `//<kind>/<name>` link expands into the
// REQUEST gvc (base.ts:95 / :185) — the four gvc-nested resource kinds.
func isGvcScopedKind(k string) bool {
	switch k {
	case "workload", "identity", "dbcluster", "volumeset":
		return true
	}
	return false
}

// linkKindSeg returns the kind segment of a `//<kind>/...` link (value.split('/')[2]).
func linkKindSeg(value string) string {
	seg := strings.Split(value, "/")
	if len(seg) > 2 {
		return seg[2]
	}
	return ""
}

// gvcSegmentOf returns the gvc segment of an absolute /org/<o>/gvc/<g>/... link.
func gvcSegmentOf(link string) string {
	comps := strings.Split(link, "/")
	if len(comps) > 4 {
		return comps[4]
	}
	return ""
}

// expandGvcLocalLink ports base.ts:61: a `//`-prefixed link is rewritten under the
// given gvc unless it already carries its own gvc/ segment.
func expandGvcLocalLink(value, org, gvc string) string {
	if !strings.HasPrefix(value, "//") {
		return value
	}
	result := strings.TrimLeft(value, "/")
	if !strings.HasPrefix(result, "gvc/") {
		result = "gvc/" + gvc + "/" + result
	}
	return "/org/" + org + "/" + result
}

// ExpandOrgLocalLink ports base.ts expandOrgLocalLink: a `//`-prefixed link is
// rewritten under the org; an absolute link passes through.
func ExpandOrgLocalLink(link, org string) string {
	if strings.HasPrefix(link, "//") {
		return "/org/" + org + "/" + strings.Trim(link, "/")
	}
	return link
}

// ExpandLocalLink ports base.ts:152-195 LocalLink.custom: lowercase-check, then
// route the `//kind/name` link to org- or gvc-expansion (gvc-scoped kinds need the
// request gvc), passing through internal + already-absolute same-org links.
func ExpandLocalLink(value, org, gvc string) (string, error) {
	if value != strings.ToLower(value) {
		return "", errors.New("Links must be lowercase")
	}
	if org == "" {
		return "", errors.New("Cannot expand " + value + " to an org-local link")
	}
	if IsInternalLink(value) {
		if !validInternalLink(value) {
			return "", errors.New("link " + value + " references an invalid internal service")
		}
		return value, nil
	}
	if !strings.HasPrefix(value, "//") {
		prefix := "/org/" + org + "/"
		if !strings.HasPrefix(value, prefix) {
			return "", errors.New("Link " + value + " is not pointing to a resource within the same org")
		}
		return value, nil
	}
	if isGvcScopedKind(linkKindSeg(value)) {
		if gvc == "" {
			return "", errors.New("Cannot expand " + value + " to a gvc-local link")
		}
		return expandGvcLocalLink(value, org, gvc), nil
	}
	return ExpandOrgLocalLink(value, org), nil
}

// ExpandImageLink ports base.ts:199-238 ImageLink.custom (the coercion half). A
// literal image (no leading "/") passes through unchanged; a link expands like a
// LocalLink but only workload/identity are gvc-scoped for images.
func ExpandImageLink(value, org, gvc string) (string, error) {
	if !strings.HasPrefix(value, "/") {
		return value, nil
	}
	if ContainsJSWhitespace(value) {
		return "", errors.New("image links cannot contain whitespace")
	}
	if org == "" {
		return "", errors.New("Cannot expand " + value + " to an org-local link")
	}
	if !strings.HasPrefix(value, "//") {
		prefix := "/org/" + org + "/"
		if !strings.HasPrefix(value, prefix) {
			return "", errors.New("Link " + value + " is not pointing to a resource within the same org")
		}
		return value, nil
	}
	if k := linkKindSeg(value); k == "workload" || k == "identity" {
		if gvc == "" {
			return "", errors.New("Cannot expand " + value + " to an gvc-local link")
		}
		return expandGvcLocalLink(value, org, gvc), nil
	}
	return ExpandOrgLocalLink(value, org), nil
}

// LinkKindOK ports base.ts:80 toKind for an ALREADY-expanded (absolute or internal)
// link: valid if it points at one of the listed kinds in the same org/gvc.
func LinkKindOK(value string, kinds []string, org, gvc string) bool {
	for _, k := range kinds {
		if k == "internal" && IsInternalLink(value) {
			return true
		}
	}
	return linkMatchesKind(value, kinds, org, gvc)
}

// linkMatchesKind ports toKind's per-kind prefix match, handling gvc-nested kinds
// and the '!'-prefixed (gvc embedded in the link) convention.
func linkMatchesKind(value string, kinds []string, org, gvc string) bool {
	for _, k := range kinds {
		prefix := "/org/" + org
		nested := k == "workload" || k == "identity" || k == "dbcluster" || k == "volumeset"
		bangNested := k == "!workload" || k == "!identity" || k == "!dbcluster" || k == "!volumeset"
		if nested {
			if gvc != "" {
				prefix = prefix + "/gvc/" + gvc + "/" + k + "/"
			} else {
				prefix = prefix + "/gvc/<gvc>/" + k + "/"
			}
		} else if bangNested {
			thisGvc := gvcSegmentOf(value)
			if thisGvc != "" {
				prefix = prefix + "/gvc/" + thisGvc + "/" + k[1:] + "/"
			} else {
				prefix = prefix + "/gvc/<gvc>/" + k[1:] + "/"
			}
		} else {
			prefix = prefix + "/" + k + "/"
		}
		suffix := ""
		if len(value) >= len(prefix) {
			suffix = value[len(prefix):]
		}
		if nested || bangNested {
			if strings.Contains(suffix, "/") {
				return false
			}
		}
		if strings.HasPrefix(value, prefix) && value != prefix {
			return true
		}
	}
	return false
}

// KindsJSON renders the kind list as toKind does for its error message.
func KindsJSON(kinds []string) string {
	parts := make([]string, len(kinds))
	for i, k := range kinds {
		parts[i] = `"` + k + `"`
	}
	return "[" + strings.Join(parts, ",") + "]"
}
