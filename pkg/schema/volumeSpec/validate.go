package volumeSpec

// validate.go attaches validation to the canonical volumeSpec types via the
// Validator interface — the go-libs home of what used to be the volUriRef/volUri/
// volPath RegisterCustom + volName RegisterPattern registrations in
// pkg/handlers/schema_volumespec.go + schema_a_reflect_workload.go.
//
// The three registry-backed field tags on VolumeSpec are replaced by strongly-
// typed nominal string types:
//   Uri  (was custom=volUriRef,custom=volUri) → VolumeSpecUri
//   Path (was custom=volPath)                 → VolumeSpecPath
//   Name (was pattern=volName)                → VolumeSpecName
// Declarative modifiers (required on Uri, max=63 on Name) stay on the field tags.

import (
	"path"
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- named regexes (ported from schema_volumespec.go) ---

var (
	volUriSchemeRe   = regexp.MustCompile(`^(s3|gs|azureblob|azurefs|cpln|scratch|k8s)://.+`)
	volNameRe        = regexp.MustCompile(`^[a-z0-9-]+$`)
	volReservedPaths = map[string]bool{"/dev": true, "/dev/log": true, "/tmp": true, "/var": true, "/var/log": true}
)

// volUriRefAllowed mirrors validateGenericInternalReference([]string{"volumeset", "secret"}).
var volUriRefAllowed = []string{"volumeset", "secret"}

func (VolumeSpecUri) ValidationRule() *validate.Rule {
	return validate.String().Custom(volUriRefCustom).Custom(volumeURICustom)
}

func (VolumeSpecPath) ValidationRule() *validate.Rule {
	return validate.String().Custom(volumePathCustom)
}

func (VolumeSpecName) ValidationRule() *validate.Rule {
	return validate.String().Regex(volNameRe)
}

// --- ported custom fns ---

// volUriRefCustom ports validateGenericInternalReference(["volumeset", "secret"]).
func volUriRefCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if strings.HasPrefix(v, "cpln://") {
		parts := strings.Split(v, "/")
		internalReference := ""
		if len(parts) > 2 {
			internalReference = parts[2]
		}
		found := false
		for _, a := range volUriRefAllowed {
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

// volumeURICustom ports the volUri custom from schema_volumespec.go.
func volumeURICustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if !volUriSchemeRe.MatchString(v) {
		return nil, h.Errorf("any.custom", "URI %s is not in the form s3://bucket, gs://bucket, azureblob://storageAccount/container, azurefs://storageAccount/share, cpln://, scratch://", v)
	}
	v = strings.TrimRight(v, "/")

	seg := func(s string) int { return len(strings.Split(s, "/")) }

	if strings.HasPrefix(v, "azureblob://") {
		if seg(v) > 4 {
			return nil, h.Errorf("any.custom", "%s is invalid: at most the storage account and container can be specified like azureblob://storageAccount/container", v)
		}
	}
	if strings.HasPrefix(v, "azurefs://") {
		if seg(v) > 4 {
			return nil, h.Errorf("any.custom", "%s is invalid: at most the storage account and file share can be specified like azurefs://storageAccount/share", v)
		}
		if seg(v) == 3 {
			return nil, h.Errorf("any.custom", "%s is invalid: missing share name after the storageaccount azurefs://storageAccount/share", v)
		}
	}
	if strings.HasPrefix(v, "s3://") || strings.HasPrefix(v, "gs://") {
		if seg(v) > 3 {
			return nil, h.Errorf("any.custom", "%s is invalid: at most the bucket can be specified like s3://bucket or gs://bucket", v)
		}
	}
	if strings.HasPrefix(v, "cpln://") {
		pieces := strings.Split(v, "/")
		p2 := ""
		if len(pieces) > 2 {
			p2 = pieces[2]
		}
		if p2 != "secret" && p2 != "volumeset" {
			return nil, h.Errorf("any.custom", "%s is invalid: the uri must begin with either cpln://secret or cpln://volumeset", v)
		}
	}
	if strings.HasPrefix(v, "k8s://") {
		pieces := strings.Split(v, "/")
		p2 := ""
		if len(pieces) > 2 {
			p2 = pieces[2]
		}
		if p2 != "secret" {
			return nil, h.Errorf("any.custom", "%s is invalid: the uri must begin with either k8s://secret", v)
		}
	}
	if strings.HasPrefix(v, "scratch://") {
		if seg(v) > 3 {
			return nil, h.Errorf("any.custom", "%s is invalid: at most the volume name can be specified like scratch://temp1 or scratch://logs", v)
		}
	}
	return v, nil
}

// volumePathCustom ports the volPath custom from schema_volumespec.go.
func volumePathCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if !path.IsAbs(v) {
		return nil, h.Errorf("any.custom", "Path %s must be absolute", v)
	}
	v = path.Clean(v)
	v = strings.TrimRight(v, "/")
	// "/" trims to "" (volumeSpec.ts:92 _.trimEnd), which the workload custom then
	// rejects as "volumes[].path is required" — remapping it back to "/" would mount
	// over the container root.
	if volReservedPaths[v] {
		return nil, h.Errorf("any.custom", "The path %s is reserved and cannot be used as mount point", v)
	}
	return v, nil
}
