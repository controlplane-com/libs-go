package env

// validate.go attaches validation to the canonical env value types via the
// Validator interface (ValidationRule), the go-libs home of the Joi customs that
// used to live in data-service/pkg/handlers/schema_env.go (wired there via
// RegisterCustom). It ports env.ts's EnvName / EnvVar.value customs verbatim.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	// envNameRe mirrors env.ts EnvName's regex (knative restriction).
	envNameRe = regexp.MustCompile(`^[-._a-zA-Z][-._a-zA-Z0-9]*$`)
	// envAllowedInternalRefs mirrors env.ts ALLOWED_INTERNAL_REFERENCES.
	envAllowedInternalRefs = []string{"reference", "secret"}
	// envAllowedManifestRefs mirrors env.ts ALLOWED_MANIFEST_REFERENCES.
	envAllowedManifestRefs = map[string]bool{
		"metadata.name":      true,
		"metadata.namespace": true,
		"spec.nodeName":      true,
		"status.hostIP":      true,
		"status.podIP":       true,
		"status.podIPs":      true,
	}
	// secretNameRe mirrors base.ts Name's regex (used by validateSecretReference
	// via Name.validate: regex + max 64).
	secretNameRe = regexp.MustCompile(`^[a-z]([-a-z0-9])*[a-z0-9]$`)
)

// ValidationRule ports EnvCollection (env.ts:44-60):
// `Joi.array().items(EnvVar).sort({by:'name'}).custom(duplicate-name check)`.
//
// The sort was missing, so an env list persisted in submitted order here and in
// name order there — a difference in the STORED document, since the array is
// stored sorted. Order matters: TS sorts BEFORE the duplicate scan.
func (EnvCollection) ValidationRule() *validate.Rule {
	return validate.Array(validate.Reflect(EnvVar{})).
		SortBy("name").
		CustomTyped(validate.CustomT(envDuplicateNames))
}

// envDuplicateNames ports the collection custom: two entries sharing a name is an
// error (not a silent dedupe).
func envDuplicateNames(env *EnvCollection, h *validate.Helper) error {
	seen := map[string]bool{}
	for _, e := range *env {
		name := string(e.Name)
		if seen[name] {
			return h.Errorf("any.custom", "Duplicate variables named %s", name)
		}
		seen[name] = true
	}
	return nil
}

// ValidationRule ports EnvName = Joi.string().trim().max(120).custom(...) plus the
// EnvVar.name field-level .disallow('K_SERVICE','K_CONFIGURATION','K_REVISION').
// trim/max/required are declarative on the EnvVar.Name field tag; the two customs
// live here in tag order (envName, then the reserved-name disallow).
func (EnvName) ValidationRule() *validate.Rule {
	return validate.String().Custom(envNameCustom).
		Invalid("K_SERVICE", "K_CONFIGURATION", "K_REVISION")
}

// ValidationRule ports EnvVar.value's custom chain: internal reference, manifest
// reference, secret reference — in that order. max(4096)/allow(”) are declarative
// on the EnvVar.Value field tag.
func (EnvValue) ValidationRule() *validate.Rule {
	return validate.String().
		Custom(validateGenericInternalReference(envAllowedInternalRefs)).
		Custom(validateManifestReference).
		Custom(validateSecretReference)
}

// envNameCustom ports EnvName's custom (CPLN_ reserved + name regex).
func envNameCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if strings.HasPrefix(s, "CPLN_") {
		return nil, h.Errorf("any.custom", "Environment variable names starting with CPLN_ are reserved for system use")
	}
	if !envNameRe.MatchString(s) {
		return nil, h.Errorf("any.custom", "Invalid environment variable name: %q. Environment variable names must contain only alphanumeric characters and the '.', '-', or '_' characters.", s)
	}
	return s, nil
}

// validateGenericInternalReference ports env.ts validateGenericInternalReference.
func validateGenericInternalReference(allowed []string) validate.CustomFn {
	return func(value any, h *validate.Helper) (any, error) {
		v, _ := value.(string)
		if strings.HasPrefix(v, "cpln://") {
			parts := strings.Split(v, "/")
			internalReference := ""
			if len(parts) > 2 {
				internalReference = parts[2]
			}
			found := false
			for _, a := range allowed {
				if a == internalReference {
					found = true
					break
				}
			}
			if !found {
				return nil, h.Errorf("any.custom", "Invalid reference: %s", internalReference)
			}
			slashIndex := len("cpln://") + len(internalReference)
			if slashIndex >= len(v) || v[slashIndex] != '/' {
				return nil, h.Errorf("any.custom", "Internal reference (%s) must be followed by a slash (/)", internalReference)
			}
			val := ""
			if len(parts) > 3 {
				val = parts[3]
			}
			if val == "" {
				return nil, h.Errorf("any.custom", "Must include a value after internal reference (%s)", internalReference)
			}
			if strings.HasPrefix(v, " ") || strings.HasSuffix(v, " ") {
				return nil, h.Errorf("any.custom", "Internal reference (%s) must not start or end with a space", internalReference)
			}
		}
		return v, nil
	}
}

// validateSecretReference ports env.ts validateSecretReference.
func validateSecretReference(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if strings.HasPrefix(v, "cpln://secret/") {
		parts := strings.Split(v, "/")
		val := ""
		if len(parts) > 3 {
			val = parts[3]
		}
		secretName := strings.Split(val, ".")[0]
		if !validSecretName(secretName) {
			return nil, h.Errorf("any.custom", "the secret reference: %s is not a valid secret name", secretName)
		}
	}
	return v, nil
}

// validateManifestReference ports env.ts validateManifestReference.
func validateManifestReference(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if strings.HasPrefix(v, "cpln://reference/") {
		parts := strings.Split(v, "/")
		ref := ""
		if len(parts) > 3 {
			ref = parts[3]
		}
		if !envAllowedManifestRefs[ref] {
			return nil, h.Errorf("any.custom", "Env reference %s now allowed", ref)
		}
	}
	return v, nil
}

// validSecretName mirrors Name.validate (base.ts Name regex + max 64).
func validSecretName(s string) bool {
	return len(s) <= 64 && secretNameRe.MatchString(s)
}
