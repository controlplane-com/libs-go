package command

// validate.go attaches validation to the canonical command types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_command.go +
// the command entries in schema_reflect_datanet.go. The registered `command` kind
// reflects Command (generic Object.unknown spec/status); the typed specs/statuses
// mirror the Joi source for parity.
//
// Irreducible bits ported here:
//   - ExpandVolumeSpec / ShrinkVolumeSpec: timeoutSeconds defaults to 600 when not
//     > 0 (was RegisterRefine(timeoutDefaultCustom), command.ts:157-162, :171-176).
//   - ReplaceVolumeStatus / ShrinkVolumeStatus newVolumeAttributes:
//     pattern(String, String().allow('')) — the value's allow('') is not taggable.
//   - DeleteVolumeSetLocationStatus.volumes: pattern(Number, DeleteCloudDevicesStatus)
//     — the NUMBER key rule (reflection hardcodes map keys to String).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"reflect"
)

// ValidationRule reflects Command — the schema the "command" kind validates against.
// Spec/Status are generic maps (Object.unknown); conflictQuery resolves to query's
// own rule via the query.Query Validator.
func (Command) ValidationRule() *validate.Rule { return validate.Reflect(Command{}) }

// timeoutSecondsDefault ports the Expand/ShrinkVolumeSpec custom: timeoutSeconds
// defaults to 600 when absent or not > 0 (command.ts:157-162, :171-176).
func timeoutSecondsDefault(obj any, _ *validate.Helper) (any, error) {
	if m, isMap := obj.(map[string]any); isMap {
		if n, ok := util.NumOf(m["timeoutSeconds"]); !ok || !(n > 0) {
			m["timeoutSeconds"] = float64(600)
		}
		return obj, nil
	}
	f, ok := validate.Field(obj, "timeoutSeconds")
	if !ok {
		return obj, nil
	}
	if f.IsNil() || !(f.Elem().Float() > 0) {
		f.Set(reflect.ValueOf(util.Ptr(600.0)))
	}
	return obj, nil
}

// ValidationRule ports CopyVolumeSpec's custom: timeoutSeconds defaults to 600, the two
// ends may only share a location when the copy goes to another volume set, and
// `locations` — the set the scheduler beams the command to, so both actuators run it —
// is always derived here, never trusted from the caller.
func (CopyVolumeSpec) ValidationRule() *validate.Rule {
	return validate.Reflect(CopyVolumeSpec{}).
		CustomTyped(timeoutSecondsDefault).
		CustomTyped(copyVolumeEnds)
}

func copyVolumeEnds(obj any, h *validate.Helper) (any, error) {
	loc, target := validate.Str(obj, "location"), validate.Str(obj, "targetLocation")
	if loc == target && validate.Str(obj, "targetVolumeSet") == "" {
		return nil, h.Errorf("any.custom", "targetLocation must differ from location unless targetVolumeSet names another volume set")
	}
	locations := []string{loc, target}
	if loc == target {
		locations = []string{loc}
	}
	if m, isMap := obj.(map[string]any); isMap {
		out := make([]any, len(locations))
		for i, l := range locations {
			out[i] = l
		}
		m["locations"] = out
		return obj, nil
	}
	if f, ok := validate.Field(obj, "locations"); ok {
		f.Set(reflect.ValueOf(locations))
	}
	return obj, nil
}

// ValidationRule reflects ExpandVolumeSpec and applies the timeoutSeconds default.
// ValidationRule ports command.ts:404-406. `locations` is
// `Joi.array().items(Joi.string().required())` — a required ITEM, which is not the same
// as min(1): it rejects an explicit `[]` with array.includesRequiredUnknowns, and the tag
// DSL cannot spell "the item is required".
func (DeleteVolumeSetSpec) ValidationRule() *validate.Rule {
	return validate.Reflect(DeleteVolumeSetSpec{}).
		Fork("locations", func(*validate.Rule) *validate.Rule {
			return validate.Array(validate.String().Required())
		})
}

func (ExpandVolumeSpec) ValidationRule() *validate.Rule {
	return validate.Reflect(ExpandVolumeSpec{}).CustomTyped(timeoutSecondsDefault)
}

// ValidationRule reflects ShrinkVolumeSpec and applies the timeoutSeconds default.
func (ShrinkVolumeSpec) ValidationRule() *validate.Rule {
	return validate.Reflect(ShrinkVolumeSpec{}).CustomTyped(timeoutSecondsDefault)
}

// attrPatternRule = pattern(String, String().allow(”)) — the map value's allow(”)
// cannot be carried by a field tag, so each attributes map type delegates here.
func attrPatternRule() *validate.Rule {
	return validate.Object(nil).Optional().Pattern(validate.String(), validate.String().Allow(""))
}

func (ReplaceVolumeStatusNewVolumeAttributes) ValidationRule() *validate.Rule {
	return attrPatternRule()
}

func (ShrinkVolumeStatusNewVolumeAttributes) ValidationRule() *validate.Rule {
	return attrPatternRule()
}

func (ShrinkVolumeStatusOldVolumeAttributes) ValidationRule() *validate.Rule {
	return attrPatternRule()
}

// ValidationRule = pattern(Number, DeleteCloudDevicesStatus) — the Number key rule
// (reflection hardcodes map keys to String, so the map type carries the rule).
func (DeleteVolumeSetLocationStatusVolumes) ValidationRule() *validate.Rule {
	return validate.Object(nil).Pattern(validate.Number(), validate.Reflect(DeleteCloudDevicesStatus{}))
}
