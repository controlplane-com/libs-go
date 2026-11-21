package base

// validate.go attaches validation to the canonical base types via the Validator
// interface (ValidationRule), the go-libs home of what used to be
// data-service/pkg/handlers/schema_base.go. It ports base.ts's shared building
// blocks — Name, Kind, Tags, description, LocalLink, ImageLink, Base, Link/Links,
// Regex — onto the base structs, plus the exported ToKind / LocalLinkRule /
// ImageLinkRule builders that per-kind link types compose with. Link expansion +
// kind matching delegate to linkalgebra.go (validate-native errors); image-string
// format validation (checkImageAndTag) is ported here.

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- Name (base.ts:291) ---

var nameRe = regexp.MustCompile(`^[a-z]([-a-z0-9])*[a-z0-9]$`)

// ValidationRule ports Name = Joi.string().regex(...).meta({unmodifiable}).max(64).
func (Name) ValidationRule() *validate.Rule {
	return validate.String().Regex(nameRe).Max(64).Unmodifiable()
}

// --- Kind (base.ts:252) ---

var kindValues = []any{
	"org", "cloudaccount", "policy", "user", "group", "resource", "task",
	"permissions", "serviceaccount", "secret", "location", "gvc", "workload",
	"quota", "identity", "deployment", "event", "domain", "image", "ipset",
	"resourcepolicy", "agent", "accessreport", "policymembership", "dbcluster",
	"auditctx", "memcachecluster", "spicedbcluster", "tenant", "mk8s", "command",
	"imagesummary", "volumeset", "sandboximage", "certificategenerator",
}

// ValidationRule ports Kind = Joi.string().valid(...).
func (Kind) ValidationRule() *validate.Rule { return validate.String().Valid(kindValues...) }

// IsValidKind reports whether s is one of the Kind enum values (base.ts:252
// Kind.valid). Exported so callers can validate a kind against the FULL enum —
// e.g. /admin/org/-addPolicy, which accepts any Kind (bb.Kind.validate), not only
// the subset a normal org bootstrap seeds builtin policies for.
func IsValidKind(s string) bool {
	for _, k := range kindValues {
		if k == s {
			return true
		}
	}
	return false
}

// ValidationRule ports CloudProvider = Joi.string().valid('aws','gcp','azure','ngs').
func (CloudProvider) ValidationRule() *validate.Rule {
	return validate.String().Valid("aws", "gcp", "azure", "ngs")
}

// --- Tags (base.ts:305) ---

var tags2Labels = map[string]bool{"cpln/logMetadata": true}

const maxTags = 50

// ValidationRule ports Tags = Joi.object().pattern(Joi.string(), Joi.any()).custom(...).
func (Tags) ValidationRule() *validate.Rule {
	return validate.Object(nil).Pattern(validate.String(), validate.Any()).
		CustomTyped(tagsCustom).Meta(map[string]any{"type": "Tags"})
}

// tagsCustom ports the Tags .custom: null→{}, drop nulls, bool/date→string, reject
// non-primitive values, enforce 63-char cap on tags2Labels, max 50 tags. It serves
// the generic map and, on the typed walk, a pointer to whichever string-keyed map
// type carries the tags (base.Tags, a row's tag column), so it works by reflection.
func tagsCustom(value any, h *validate.Helper) (any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return value, nil
	}
	for _, key := range rv.MapKeys() {
		p := key.String()
		ev := rv.MapIndex(key)
		var v any
		if ev.IsValid() && (ev.Kind() != reflect.Interface || !ev.IsNil()) {
			v = ev.Interface()
		}
		if v == nil {
			rv.SetMapIndex(key, reflect.Value{})
			continue
		}
		switch t := v.(type) {
		case bool:
			v = strconv.FormatBool(t)
		case time.Time:
			v = t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
		case string:
		case float64, float32, int, int64:
		default:
			// base.ts:328 interpolates the value into a template literal, so an
			// object renders as [object Object] and an array as its joined elements
			// — NOT as Go's map[b:1] / [1 2].
			return nil, h.Errorf("any.custom", "A tag's value must be primitive, found a '%s' for %s", JSString(v), p)
		}
		if tags2Labels[p] {
			// base.ts:334 tests `value[p].length` — UTF-16 code units, not bytes.
			// len() rejected any accented value over ~31 characters that TS accepts.
			if s, ok := v.(string); ok && validate.JSLen(s) > 63 {
				return nil, h.Errorf("any.custom", "The value of %s must be shorter than 63 characters", p)
			}
		}
		rv.SetMapIndex(key, reflect.ValueOf(&v).Elem())
	}
	if rv.Len() > maxTags {
		return nil, h.Errorf("any.custom", "You cannot have more than %d tags per object", maxTags)
	}
	return value, nil
}

// --- LocalLink (base.ts:152) ---

var localLinkRe = regexp.MustCompile(`(/org/[^/]+/.*)|(//.+)`)

// LocalLinkRule ports LocalLink = Joi.string().trim().regex(...).custom(expand).
// Expansion + the lowercase check delegate to linkalgebra.go.
func LocalLinkRule() *validate.Rule {
	return validate.String().Trim().Regex(localLinkRe).Custom(localLinkExpandCustom)
}

// LocalLinkShapeRule is the SHAPE half of LocalLinkRule — the trim + pattern, without
// the expansion custom. The central pre-schema link pass (data-service links.go) uses
// it so a value that is not link-shaped answers with Joi's `string.pattern.base`, as it
// does in TS where expansion happens INSIDE the schema and so never sees a value the
// pattern already rejected.
func LocalLinkShapeRule() *validate.Rule {
	return validate.String().Trim().Regex(localLinkRe)
}

// ValidationRule makes the canonical LocalLink type self-validating.
func (LocalLink) ValidationRule() *validate.Rule { return LocalLinkRule() }

func localLinkExpandCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	org, gvc := ctxOrgGvc(h)
	expanded, err := ExpandLocalLink(s, org, gvc)
	if err != nil {
		return nil, h.Errorf("any.custom", "%s", err.Error())
	}
	return expanded, nil
}

// ToKind ports base.ts toKind(...) as a .custom() to append to a LocalLink/
// ImageLink rule. It kind-checks the already-expanded link via linkalgebra.go.
func ToKind(kinds ...string) validate.CustomFn {
	return func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		org, gvc := ctxOrgGvc(h)
		if !LinkKindOK(s, kinds, org, gvc) {
			orig, _ := h.Original.(string)
			if orig == "" {
				orig = s
			}
			return nil, h.Errorf("any.custom", "Link %s is not pointing to a resource of kind %s in the same org/gvc", orig, KindsJSON(kinds))
		}
		return s, nil
	}
}

// --- ImageLink (base.ts:199) ---

// ImageLinkRule ports ImageLink = Joi.string().trim().custom(expand).custom(kind).
func ImageLinkRule() *validate.Rule {
	return validate.String().Trim().Custom(imageLinkExpandCustom).Custom(imageLinkKindCustom)
}

// ValidationRule makes the canonical ImageLink type self-validating.
func (ImageLink) ValidationRule() *validate.Rule { return ImageLinkRule() }

func imageLinkExpandCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if !strings.HasPrefix(s, "/") {
		if err := checkImageAndTag(s); err != nil {
			// returned as-is: a sub-schema failure carries the nested error Joi puts
			// in context.error, which h.Errorf would flatten to its message.
			return nil, err
		}
		return s, nil
	}
	org, gvc := ctxOrgGvc(h)
	expanded, err := ExpandImageLink(s, org, gvc)
	if err != nil {
		return nil, h.Errorf("any.custom", "%s", err.Error())
	}
	return expanded, nil
}

func imageLinkKindCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if !strings.HasPrefix(s, "/") {
		if err := checkImageAndTag(s); err != nil {
			// returned as-is: a sub-schema failure carries the nested error Joi puts
			// in context.error, which h.Errorf would flatten to its message.
			return nil, err
		}
		return s, nil
	}
	org, gvc := ctxOrgGvc(h)
	if !LinkKindOK(s, []string{"image"}, org, gvc) {
		orig, _ := h.Original.(string)
		if orig == "" {
			orig = s
		}
		return nil, h.Errorf("any.custom", "Link %s is not pointing to a resource of kind %s in the same org/gvc", orig, KindsJSON([]string{"image"}))
	}
	return s, nil
}

// --- image string format validation (base.ts checkImageAndTag) ---

var (
	imageNameRe = regexp.MustCompile(`^[a-z0-9\-./_]+(?::)?[a-z0-9\-./_]+$`)
	// The JS pattern carries a `(?![0-9]+$)` lookahead RE2 cannot express; the
	// all-digits case is guarded separately and the JS source is what gets reported.
	imageNameJS = `^(?![0-9]+$)[a-z0-9\-./_]+(?::)?[a-z0-9\-./_]+$`
	imageTagRe  = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_\-.]*$`)

	// The rules base.ts spells as ImageName / ImageTag. An all-digits name cannot be
	// expressed in RE2, so it is rejected by a Custom that raises the SAME
	// string.pattern.base error the regex would, keeping Joi's rule order (pattern
	// before max) intact.
	imageNameRule = validate.String().
			Custom(func(v any, h *validate.Helper) (any, error) {
			if s, _ := v.(string); allDigitsRe.MatchString(s) {
				// Joi's template is {{#label}}, not the literal "value": at a nested
				// path (a container's image) the caller has to be told WHICH field.
				return nil, h.WithCtx(h.Errorf("string.pattern.base",
					`%q with value %q fails to match the required pattern: /%s/`, h.Label(), s, imageNameJS),
					"regex", map[string]any{}, "value", s)
			}
			return v, nil
		}).
		RegexAs(imageNameRe, imageNameJS).
		Max(128)

	imageTagRule = validate.String().Regex(imageTagRe).Max(128)
	allDigitsRe  = regexp.MustCompile(`^[0-9]+$`)
	sha256HexRe  = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

func checkImageAndTag(image string) error {
	masked := strings.Replace(image, "@sha256:", "@sha256@", 1)
	parts := strings.Split(masked, ":")
	switch len(parts) {
	case 1:
		return checkImageName(parts[0])
	case 2:
		possiblePort := strings.SplitN(parts[1], "/", 2)[0]
		if allDigitsRe.MatchString(possiblePort) || possiblePort == "" {
			return checkImageName(parts[0] + ":" + parts[1])
		}
		if err := checkImageName(parts[0]); err != nil {
			return err
		}
		return checkImageTag(parts[1])
	case 3:
		up := upToSlash(parts[1])
		if isNaNNumber(up) || !strings.Contains(parts[1], "/") {
			return errStr(`too many ":"`)
		}
		if err := checkImageName(parts[0] + ":" + parts[1]); err != nil {
			return err
		}
		return checkImageTag(parts[2])
	default:
		return errStr("Container images may not contain more than two colon characters.")
	}
}

func checkImageName(maybeName string) error {
	justName := maybeName
	if i := strings.Index(maybeName, "@sha256@"); i > 0 {
		justName = maybeName[:i]
		checksum := maybeName[i+8:]
		if !sha256HexRe.MatchString(checksum) {
			return errStr("The content hash sha256:" + checksum + " is not valid")
		}
	}
	// base.ts checkImageName throws ImageName.validate(justName)'s own Joi error, so
	// the message must be the engine's, not a hand-written lookalike: the label is
	// "value" (a standalone validate), the pattern error quotes the JS source, and
	// regex is checked before max because Joi applies rules in declaration order.
	_, verr := validate.Validate(imageNameRule, justName, nil, validate.Prefs{})
	return validate.Nested(justName, verr)
}

func checkImageTag(maybeTag string) error {
	justTag := maybeTag
	if i := strings.Index(maybeTag, "@sha256@"); i > 0 {
		justTag = maybeTag[:i]
		checksum := maybeTag[i+8:]
		if !sha256HexRe.MatchString(checksum) {
			return errStr("The content hash sha256:" + checksum + " is not valid")
		}
	}
	_, verr := validate.Validate(imageTagRule, justTag, nil, validate.Prefs{})
	return validate.Nested(justTag, verr)
}

func upToSlash(s string) string {
	if i := strings.Index(s, "/"); i >= 0 {
		return s[:i]
	}
	return s
}

// isNaNNumber mirrors JS isNaN(Number(s)): empty/whitespace → 0 (not NaN).
func isNaNNumber(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return false
	}
	_, err := strconv.ParseFloat(t, 64)
	return err != nil
}

func errStr(s string) error { return &imageErr{s} }

type imageErr struct{ msg string }

func (e *imageErr) Error() string { return e.msg }

// --- Regex (base.ts:440) — Go's regexp IS RE2. ---

// RegexRule ports Regex = Joi.string().custom(RE2-compile-check).
func RegexRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if _, err := regexp.Compile(s); err != nil {
			return nil, h.Errorf("any.custom", `"%s" is not a valid RE2 regex pattern: %s`, s, err.Error())
		}
		return s, nil
	})
}

// ValidationRule makes the canonical Regex type self-validating.
func (Regex) ValidationRule() *validate.Rule { return RegexRule() }

// --- Link / Links (base.ts:350) ---

// LinkRule ports Link = Joi.object({rel: required trim, href: required trim}).
func LinkRule() *validate.Rule {
	// Ordered: Object() sorts (a Go map has no order), and rel/href IS the Joi
	// declaration order that decides which key an abortEarly failure reports.
	return validate.ObjectOrdered([]string{"rel", "href"}, map[string]*validate.Rule{
		"rel":  validate.String().Required().Trim(),
		"href": validate.String().Required().Trim(),
	})
}

// ValidationRule makes the canonical Link type self-validating.
func (Link) ValidationRule() *validate.Rule { return LinkRule() }

// LinksRule ports Links = Joi.array().items(Link).
func LinksRule() *validate.Rule { return validate.Array(LinkRule()) }

// ValidationRule makes the canonical Links type self-validating.
func (Links) ValidationRule() *validate.Rule { return LinksRule() }

// DescriptionRule ports description: Joi.string().trim().max(250).allow(”).
func DescriptionRule() *validate.Rule {
	return validate.String().Trim().Max(250).Allow("")
}

// --- Base (base.ts:358) ---

// BaseRule ports Base = Joi.object({id, name, kind, version, description, tags,
// created, lastModified, links}). Resources compose it via Reflect(self) plus
// their spec/status keys.
func BaseRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"id":           validate.String().UUID().ReadOnly(),
		"name":         Name("").ValidationRule(),
		"kind":         Kind("").ValidationRule(),
		"version":      validate.Number().Positive().Integer().ReadOnly(),
		"description":  DescriptionRule(),
		"tags":         Tags(nil).ValidationRule().Default(func() any { return map[string]any{} }),
		"created":      validate.DateRule().ReadOnly(),
		"lastModified": validate.DateRule().ReadOnly(),
		"links":        LinksRule().ReadOnly(),
	})
}

// ValidationRule makes the canonical Base type self-validating.
func (Base) ValidationRule() *validate.Rule { return BaseRule() }

// DescriptionDefaultsToName ports the common createSchema fork
// `description: DescriptionRule().Default(Joi.ref('name'))`: when description is
// absent/empty, default it to the name. Shared across kinds, so it reads the two
// fields by json name; attach with `.CustomTyped(base.DescriptionDefaultsToName)`
// on a resource's ValidationRule.
func DescriptionDefaultsToName(obj any, h *validate.Helper) (any, error) {
	if m, isMap := obj.(map[string]any); isMap {
		if v, present := m["description"]; present && v != nil {
			return obj, nil
		}
		if name, ok := m["name"].(string); ok && name != "" {
			m["description"] = name
		}
		return obj, nil
	}
	desc, ok := validate.Field(obj, "description")
	if !ok || desc.Kind() != reflect.String || h.Present("description") {
		return obj, nil
	}
	if name, ok := validate.Field(obj, "name"); ok && name.Kind() == reflect.String && name.String() != "" {
		desc.SetString(name.String())
	}
	return obj, nil
}

// --- helper: read org/gvc from the validation context ---

func ctxOrgGvc(h *validate.Helper) (string, string) {
	if h.Ctx == nil {
		return "", ""
	}
	return h.Ctx.Org, h.Ctx.Gvc
}

// ErrorAtKey anchors an error at a CHILD key of the object a container custom runs
// on, for the Joi error types whose context carries no `value` (any.required,
// any.unknown). Helper.ErrorAt always stamps one, so the stamp is dropped here —
// without this, a hand-raised `"provider" is required` either blamed the whole object
// (path []) or carried a context key Joi does not emit.
func ErrorAtKey(h *validate.Helper, key, typ, format string, args ...any) *validate.Error {
	e := h.ErrorAt([]any{key}, nil, typ, format, args...)
	delete(e.Ctx, "value")
	return e
}
