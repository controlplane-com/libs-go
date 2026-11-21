package validate

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// --- fixtures ---

type wrap[T any] struct {
	val    T
	parsed bool
	null   bool
}

func (w *wrap[T]) InnerForWrite() any { w.parsed, w.null = true, false; return &w.val }
func (w *wrap[T]) SetNull()           { w.null, w.parsed = true, false; var z T; w.val = z }
func (w *wrap[T]) IsNull() bool       { return w.null }
func (w *wrap[T]) MarshalJSON() ([]byte, error) {
	if w.null || !w.parsed {
		return []byte("null"), nil
	}
	return json.Marshal(w.val)
}

type extras map[string]any

func (extras) ExtrasMarker() {}

type wName string

func (wName) ValidationRule() *Rule { return String().Regex(reNameTest).Max(10) }

var reNameTest = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type wPort struct {
	Number   float64 `json:"number" validate:"port,required"`
	Protocol string  `json:"protocol,omitzero" validate:"valid=http|tcp,default=http"`
}

type wSpec struct {
	Type      string    `json:"type,omitzero" validate:"valid=serverless|standard,default=serverless"`
	Replicas  *float64  `json:"replicas,omitzero" validate:"integer,min=0"`
	Ports     []wPort   `json:"ports,omitzero" validate:"uniqueBy=number,max=3"`
	Tags      []string  `json:"tags,omitzero" validate:"sort,dedupe"`
	Labels    wLabels   `json:"labels,omitzero"`
	Note      string    `json:"note,omitzero" validate:"trim,allow="`
	Debug     *bool     `json:"debug,omitzero"`
	Anything  any       `json:"anything,omitzero"`
	Created   time.Time `json:"created,omitzero" validate:"date"`
	Extras    extras    `json:"-"`
	MaxRef    float64   `json:"maxRef,omitzero" validate:"minRef=minRef"`
	MinRef    float64   `json:"minRef,omitzero"`
	Alt       string    `json:"alt,omitzero"`
	private   int
	Forbidden string `json:"forbidden,omitzero" validate:"forbidden"`
}

type wLabels map[string]string

func (wLabels) ValidationRule() *Rule {
	return Object(nil).Pattern(String().Max(5), String().Lowercase())
}

func (wSpec) ValidationRule() *Rule {
	r := Reflect(wSpec{}).Unknown(true)
	r = r.Keys(map[string]*Rule{
		"alt":   Alternatives(String().Valid("a"), Number()),
		"ports": Array(Reflect(wPort{})).SortBy("number").Unique("number").Max(3),
	})
	return RefineT(r, func(s *wSpec, h *Helper) error {
		if s.Type == "standard" && s.Replicas == nil {
			one := 1.0
			s.Replicas = &one
		}
		if s.Note == "boom" {
			return h.Errorf("any.custom", "note says boom")
		}
		return nil
	})
}

type wBase struct {
	Name    wName  `json:"name,omitzero" validate:"required"`
	Version int    `json:"version,omitzero" validate:"integer,positive"`
	Kind    string `json:"kind,omitzero" validate:"valid=thing"`
}

type wRow struct {
	wBase
	Spec   *wrap[wSpec] `json:"spec,omitzero"`
	Status wrap[wSpec]  `json:"status,omitzero"`
	Id     string       `json:"id,omitzero" validate:"uuid"`
}

// the schema type mirrors the row without the wrappers, as go-libs models do
type wSchema struct {
	Name    wName  `json:"name,omitzero" validate:"required"`
	Version int    `json:"version,omitzero" validate:"integer,positive"`
	Kind    string `json:"kind,omitzero" validate:"valid=thing"`
	Spec    wSpec  `json:"spec,omitzero"`
	Status  wSpec  `json:"status,omitzero"`
	Id      string `json:"id,omitzero" validate:"uuid"`
}

func rowRule() *Rule { return For[wSchema]().Rule() }

// cSpec/cSchema are wSpec/wSchema without the typed refine, so the generic
// evaluator can run the same graph for the output-equivalence gate.
type cSpec struct {
	Type     string    `json:"type,omitzero" validate:"valid=serverless|standard,default=serverless"`
	Replicas *float64  `json:"replicas,omitzero" validate:"integer,min=0"`
	Ports    []wPort   `json:"ports,omitzero"`
	Tags     []string  `json:"tags,omitzero" validate:"sort,dedupe"`
	Labels   wLabels   `json:"labels,omitzero"`
	Note     string    `json:"note,omitzero" validate:"trim,allow="`
	Debug    *bool     `json:"debug,omitzero"`
	Anything any       `json:"anything,omitzero"`
	Created  time.Time `json:"created,omitzero" validate:"date"`
	Extras   extras    `json:"-"`
	MaxRef   float64   `json:"maxRef,omitzero" validate:"minRef=minRef"`
	MinRef   float64   `json:"minRef,omitzero"`
	Alt      string    `json:"alt,omitzero"`
}

func (cSpec) ValidationRule() *Rule {
	return Reflect(cSpec{}).Unknown(true).Keys(map[string]*Rule{
		"alt":   Alternatives(String().Valid("a"), Number()),
		"ports": Array(Reflect(wPort{})).SortBy("number").Unique("number").Max(3),
	})
}

func (c *cSpec) MarshalJSON() ([]byte, error) {
	type alias cSpec
	b, err := json.Marshal((*alias)(c))
	if err != nil || len(c.Extras) == 0 {
		return b, err
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	for k, v := range c.Extras {
		m[k] = v
	}
	return json.Marshal(m)
}

type cSchema struct {
	Name    wName  `json:"name,omitzero" validate:"required"`
	Version int    `json:"version,omitzero" validate:"integer,positive"`
	Kind    string `json:"kind,omitzero" validate:"valid=thing"`
	Spec    cSpec  `json:"spec,omitzero"`
	Status  cSpec  `json:"status,omitzero"`
	Id      string `json:"id,omitzero" validate:"uuid"`
}

func decodeRow(t *testing.T, body string, prefs Prefs) (*wRow, *Error) {
	t.Helper()
	row := &wRow{}
	d, err := DecodeInto(row, Bytes(body))
	if err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return row, d.Validate(rowRule(), &Context{Method: "POST"}, prefs)
}

func wantErr(t *testing.T, verr *Error, typ, msg string) {
	t.Helper()
	if verr == nil {
		t.Fatalf("expected %s error %q, got success", typ, msg)
	}
	if verr.Type != typ || !strings.Contains(verr.Message, msg) {
		t.Fatalf("expected %s %q, got %s %q", typ, msg, verr.Type, verr.Message)
	}
}

// --- tests ---

func TestTypedDecodeFillsTheStruct(t *testing.T) {
	row, verr := decodeRow(t, `{"name":"abc","version":3,"spec":{"type":"standard","ports":[{"number":9},{"number":3,"protocol":"tcp"}],"tags":["b","a","b"],"labels":{"k":"UP"},"note":"  hi  ","debug":false,"anything":{"x":[1]},"created":"2024-01-02T03:04:05Z"}}`, Prefs{})
	if verr != nil {
		t.Fatal(verr)
	}
	if row.Name != "abc" || row.Version != 3 {
		t.Fatalf("base: %+v", row.wBase)
	}
	s := row.Spec.val
	if s.Type != "standard" || s.Replicas == nil || *s.Replicas != 1 {
		t.Fatalf("refine default not applied: %+v", s)
	}
	if len(s.Ports) != 2 || s.Ports[0].Number != 3 || s.Ports[0].Protocol != "tcp" || s.Ports[1].Protocol != "http" {
		t.Fatalf("ports not sorted/defaulted: %+v", s.Ports)
	}
	if !reflect.DeepEqual(s.Tags, []string{"a", "b"}) {
		t.Fatalf("tags not sorted/deduped: %v", s.Tags)
	}
	if s.Labels["k"] != "up" {
		t.Fatalf("map value not lowercased: %v", s.Labels)
	}
	if s.Note != "hi" {
		t.Fatalf("note not trimmed: %q", s.Note)
	}
	if s.Debug == nil || *s.Debug {
		t.Fatalf("explicit false lost: %v", s.Debug)
	}
	if !reflect.DeepEqual(s.Anything, map[string]any{"x": []any{1.0}}) {
		t.Fatalf("any field: %#v", s.Anything)
	}
	if !s.Created.Equal(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("created: %v", s.Created)
	}
	if row.Status.parsed {
		t.Fatalf("absent status was materialized: %+v", row.Status)
	}
}

func TestTypedPresenceSemantics(t *testing.T) {
	_, verr := decodeRow(t, `{"version":1}`, Prefs{})
	wantErr(t, verr, "any.required", `"name" is required`)

	_, verr = decodeRow(t, `{"name":null}`, Prefs{})
	wantErr(t, verr, "string.base", `"name" must be a string`)

	_, verr = decodeRow(t, `{"name":"a","version":0}`, Prefs{})
	wantErr(t, verr, "number.positive", `"version" must be a positive number`)

	_, verr = decodeRow(t, `{"name":"a","version":1.5}`, Prefs{})
	wantErr(t, verr, "number.integer", `"version" must be an integer`)

	_, verr = decodeRow(t, `{"name":"a","version":"2"}`, Prefs{})
	if verr != nil {
		t.Fatalf("a numeric string coerces like Joi convert:true: %v", verr)
	}

	_, verr = decodeRow(t, `{"name":"a","typo":1}`, Prefs{})
	wantErr(t, verr, "object.unknown", `"typo" is not allowed`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"forbidden":"x"}}`, Prefs{})
	wantErr(t, verr, "any.unknown", `"spec.forbidden" is not allowed`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"ports":[{"number":1},{"number":1}]}}`, Prefs{})
	wantErr(t, verr, "array.unique", `"spec.ports[1]" contains a duplicate value`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"ports":[{"protocol":"tcp"}]}}`, Prefs{})
	wantErr(t, verr, "any.required", `"spec.ports[0].number" is required`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"ports":[7]}}`, Prefs{})
	wantErr(t, verr, "object.base", `"spec.ports[0]" must be of type object`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"labels":{"toolong":"x"}}}`, Prefs{})
	wantErr(t, verr, "object.unknown", `"spec.labels.toolong" is not allowed`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"minRef":5,"maxRef":3}}`, Prefs{})
	wantErr(t, verr, "number.min", `"spec.maxRef" must be greater than or equal to ref:minRef`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"note":"boom"}}`, Prefs{})
	wantErr(t, verr, "any.custom", `"spec" failed custom validation because note says boom`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"alt":"b"}}`, Prefs{})
	wantErr(t, verr, "alternatives.types", `"spec.alt" must be one of [a, number]`)

	_, verr = decodeRow(t, `{"name":"a","spec":{"created":"nope"}}`, Prefs{})
	wantErr(t, verr, "date.base", `"spec.created" must be a valid date`)

	_, verr = decodeRow(t, `{"name":"a","spec":"str"}`, Prefs{})
	wantErr(t, verr, "object.base", `"spec" must be of type object`)

	_, verr = decodeRow(t, `[1]`, Prefs{})
	wantErr(t, verr, "object.base", `"value" must be of type object`)

}

func TestTypedNullAndExtras(t *testing.T) {
	_, verr := decodeRow(t, `{"name":"a","status":null}`, Prefs{})
	wantErr(t, verr, "object.base", `"status" must be of type object`)

	row, verr := decodeRow(t, `{"name":"a","spec":{"note":"","zzz":{"deep":true},"anything":null}}`, Prefs{})
	if verr != nil {
		t.Fatal(verr)
	}
	if row.Spec.val.Extras["zzz"] == nil {
		t.Fatalf("unknown key not kept in extras: %v", row.Spec.val.Extras)
	}
	if row.Spec.val.Note != "" {
		t.Fatalf("allowed empty string: %q", row.Spec.val.Note)
	}
	// the same body with the stripUnknown PREF: the spec rule carries an explicit
	// .unknown(true), which outranks the preference in Joi
	// (`prefs.stripUnknown && !schema._flags.unknown`), so the key survives.
	row, verr = decodeRow(t, `{"name":"a","spec":{"zzz":1}}`, Prefs{StripUnknown: true})
	if verr != nil {
		t.Fatal(verr)
	}
	if row.Spec.val.Extras["zzz"] == nil {
		t.Fatalf("schema .unknown(true) lost to the stripUnknown pref: %v", row.Spec.val.Extras)
	}
}

func TestTypedSeamSetFieldsCountAsPresent(t *testing.T) {
	row := &wRow{}
	d, err := DecodeInto(row, Bytes(`{"spec":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	row.Name = "seamed" // a pre-schema hook wrote it
	if verr := d.Validate(rowRule(), nil, Prefs{}); verr != nil {
		t.Fatal(verr)
	}
	if row.Spec.val.Type != "serverless" {
		t.Fatalf("nested default: %+v", row.Spec.val)
	}
}

func TestTypedDuplicateKeysLastWins(t *testing.T) {
	row, verr := decodeRow(t, `{"name":"a","name":"b","version":1,"version":2}`, Prefs{})
	if verr != nil {
		t.Fatal(verr)
	}
	if row.Name != "b" || row.Version != 2 {
		t.Fatalf("%+v", row.wBase)
	}
}

func TestTypedTreeSourceMatchesBytes(t *testing.T) {
	body := `{"name":"abc","version":"3","spec":{"type":"standard","ports":[{"number":9},{"number":3}],"tags":["b","a"],"note":" x ","labels":{"k":"UP"}}}`
	var tree any
	if err := json.Unmarshal([]byte(body), &tree); err != nil {
		t.Fatal(err)
	}
	a := &wRow{}
	da, _ := DecodeInto(a, Bytes(body))
	b := &wRow{}
	db, _ := DecodeInto(b, Tree{tree})
	if verr := da.Validate(rowRule(), nil, Prefs{}); verr != nil {
		t.Fatal(verr)
	}
	if verr := db.Validate(rowRule(), nil, Prefs{}); verr != nil {
		t.Fatal(verr)
	}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatalf("bytes and tree sources disagree:\n%s\n%s", ja, jb)
	}
}

func TestTypedMatchesGenericOutput(t *testing.T) {
	// the generic evaluator's normalized map, pushed through the type the way the
	// old write path did, must equal the typed result byte for byte
	bodies := []string{
		`{"name":"abc","version":3,"spec":{"type":"standard","ports":[{"number":9},{"number":3,"protocol":"tcp"}],"tags":["b","a","b"],"labels":{"k":"UP"},"note":"  hi  ","debug":false,"anything":{"x":[1]},"created":"2024-01-02T03:04:05Z"}}`,
		`{"name":"a","spec":{"note":"","debug":true},"status":{"type":"standard"}}`,
		`{"name":"a","version":"7","spec":{"minRef":1,"maxRef":2,"alt":"a"}}`,
	}
	for _, body := range bodies {
		var tree any
		_ = json.Unmarshal([]byte(body), &tree)
		rule := For[cSchema]().Rule()
		out, verr := Validate(rule, tree, &Context{Method: "POST"}, Prefs{})
		if verr != nil {
			t.Fatalf("generic: %v", verr)
		}
		gb, _ := json.Marshal(out)
		var viaType cSchema
		if err := json.Unmarshal(gb, &viaType); err != nil {
			t.Fatal(err)
		}
		want, _ := json.Marshal(&viaType)

		typed := &cSchema{}
		d, err := DecodeInto(typed, Bytes(body))
		if err != nil {
			t.Fatal(err)
		}
		if verr := d.Validate(rule, &Context{Method: "POST"}, Prefs{}); verr != nil {
			t.Fatalf("typed: %v", verr)
		}
		got, _ := json.Marshal(typed)
		if string(got) != string(want) {
			t.Fatalf("typed output differs from the generic path for %s\n got: %s\nwant: %s", body, got, want)
		}
	}
}

func TestTypedRefusesMapCustoms(t *testing.T) {
	type bad struct {
		A string `json:"a"`
	}
	rule := Reflect(bad{}).Refine(func(m map[string]any, h *Helper) error { return nil })
	d, _ := DecodeInto(&bad{}, Bytes(`{"a":"x"}`))
	defer func() {
		if recover() == nil {
			t.Fatal("a map refine on the typed walk must not silently no-op")
		}
	}()
	_ = d.Validate(rule, nil, Prefs{})
}

func TestGenericRefusesTypedCustoms(t *testing.T) {
	type good struct {
		A string `json:"a"`
	}
	rule := RefineT(Reflect(good{}), func(g *good, h *Helper) error { return nil })
	defer func() {
		if recover() == nil {
			t.Fatal("a typed refine on the generic evaluator must not silently no-op")
		}
	}()
	_, _ = Validate(rule, map[string]any{"a": "x"}, nil, Prefs{})
}

func TestTypedSyntaxError(t *testing.T) {
	if _, err := DecodeInto(&wRow{}, Bytes(`{"name":`)); err == nil {
		t.Fatal("truncated body accepted")
	}
	if _, err := DecodeInto(&wRow{}, Bytes(`{} x`)); err == nil {
		t.Fatal("trailing garbage accepted")
	}
}

// A default that refers to a malformed sibling (description defaulting to a numeric
// name) cannot be stored in the typed field. Joi does not validate defaults, so the
// sibling's own error is the answer — not a panic from the assignment.
func TestTypedDefaultRefToMalformedSiblingIsNotFatal(t *testing.T) {
	type row struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	}
	rule := ObjectOrdered([]string{"description", "name"}, map[string]*Rule{
		"description": String().Default(RefTo("name")),
		"name":        String().Required(),
	})
	var r row
	d, err := DecodeInto(&r, Bytes(`{"name":123}`))
	if err != nil {
		t.Fatal(err)
	}
	verr := d.Validate(rule, &Context{}, Prefs{})
	if verr == nil || verr.Message != `"name" must be a string` {
		t.Fatalf("want the name error, got %v", verr)
	}
}

// Joi puts the failing value in context.value on EVERY error; the typed walk's three
// typeBaseError seams answered without one, so the same body reported a different
// context on the write path than on the tree path (live: cloudaccount `data`, whose
// row column is a JSONB wrapper under an `any` rule).
func TestTypedTypeBaseErrorsCarryTheValue(t *testing.T) {
	t.Run("rule accepts a shape the field cannot hold", func(t *testing.T) {
		type row struct {
			Blob map[string]string `json:"blob,omitzero"`
		}
		rule := Object(map[string]*Rule{"blob": Any()})
		var r row
		d, err := DecodeInto(&r, Bytes(`{"blob":[1]}`))
		if err != nil {
			t.Fatal(err)
		}
		verr := d.Validate(rule, &Context{}, Prefs{})
		wantErr(t, verr, "object.base", `"blob" must be of type object`)
		if got := verr.Ctx["value"]; !reflect.DeepEqual(got, []any{1.0}) {
			t.Fatalf("context.value = %#v, want [1]", got)
		}
	})

	t.Run("rule type the field cannot hold", func(t *testing.T) {
		type row struct {
			Items []string `json:"items,omitzero"`
		}
		rule := Object(map[string]*Rule{"items": String()})
		var r row
		d, err := DecodeInto(&r, Bytes(`{"items":["a"]}`))
		if err != nil {
			t.Fatal(err)
		}
		verr := d.Validate(rule, &Context{}, Prefs{})
		wantErr(t, verr, "string.base", `"items" must be a string`)
		if got := verr.Ctx["value"]; !reflect.DeepEqual(got, []any{"a"}) {
			t.Fatalf("context.value = %#v, want [a]", got)
		}
	})

	t.Run("root body is not an object", func(t *testing.T) {
		type row struct {
			A string `json:"a,omitzero"`
		}
		var r row
		d, err := DecodeInto(&r, Bytes(`[1]`))
		if err != nil {
			t.Fatal(err)
		}
		verr := d.Validate(Object(map[string]*Rule{"a": String()}), &Context{}, Prefs{})
		wantErr(t, verr, "object.base", `"value" must be of type object`)
		if _, has := verr.Ctx["value"]; !has {
			t.Fatalf("root object.base carries no context.value: %#v", verr.Ctx)
		}
	})
}

// A custom whose return value the field cannot hold used to reach assign()'s panic —
// a 500 on the write path. Which custom branch runs is request-data driven, so this is
// a body-reachable panic, not a schema-load assert.
func TestTypedCustomReturningAnUnstorableValueIsNotFatal(t *testing.T) {
	blob := func(any, *Helper) (any, error) { return map[string]any{"x": 1.0}, nil }

	t.Run("scalar field", func(t *testing.T) {
		type row struct {
			Note string `json:"note,omitzero"`
		}
		var r row
		d, err := DecodeInto(&r, Bytes(`{"note":"hi"}`))
		if err != nil {
			t.Fatal(err)
		}
		verr := d.Validate(Object(map[string]*Rule{"note": String().Custom(blob)}), &Context{}, Prefs{})
		wantErr(t, verr, "string.base", `"note" must be a string`)
	})

	t.Run("array element", func(t *testing.T) {
		type row struct {
			Tags []string `json:"tags,omitzero"`
		}
		var r row
		d, err := DecodeInto(&r, Bytes(`{"tags":["a"]}`))
		if err != nil {
			t.Fatal(err)
		}
		verr := d.Validate(Object(map[string]*Rule{"tags": Array(String().Custom(blob))}), &Context{}, Prefs{})
		wantErr(t, verr, "string.base", `"tags[0]" must be a string`)
	})
}

// TS materializes an omitted subtree before Joi runs (volumeset.ts:914
// `if (!req.body.status) { req.body.status = {}; }`) so the schema's nested defaults
// land there and an apply-merge REPLACES the stored subtree instead of copying it back.
// A Go struct field is always physically there, so presence — not assignment — is what
// has to be said, and a nil pointer/wrapper has to be allocated for the walk to descend.
func TestMarkPresentMaterializesAnAbsentSubtree(t *testing.T) {
	row := &wRow{}
	d, err := DecodeInto(row, Bytes(`{"name":"abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Present("spec") {
		t.Fatal("precondition: spec must start absent")
	}
	f, ok := Field(row, "spec")
	if !ok {
		t.Fatal("no spec field")
	}
	if _, ok := PeelForWrite(f); !ok {
		t.Fatal("PeelForWrite must allocate the nil wrapper")
	}
	d.MarkPresent("spec")
	if !d.Present("spec") {
		t.Fatal("spec must read as present after MarkPresent")
	}
	if row.Spec == nil {
		t.Fatal("PeelForWrite must have allocated the *wrap")
	}
	if verr := d.Validate(rowRule(), nil, Prefs{}); verr != nil {
		t.Fatal(verr)
	}
	if row.Spec.val.Type != "serverless" {
		t.Fatalf("the nested default must land in the materialized subtree: %+v", row.Spec.val)
	}
}

// MarkPresent must never DEMOTE a key: a body that carried the key, or a seam that set
// it, stays present.
func TestMarkPresentIsIdempotent(t *testing.T) {
	row := &wRow{}
	d, err := DecodeInto(row, Bytes(`{"name":"abc","spec":{"type":"standard"}}`))
	if err != nil {
		t.Fatal(err)
	}
	d.MarkPresent("spec")
	d.MarkPresent("nosuchkey") // must not panic
	if !d.Present("spec") {
		t.Fatal("spec must stay present")
	}
	if row.Spec == nil || row.Spec.val.Type != "standard" {
		t.Fatalf("MarkPresent must not disturb the decoded value: %+v", row.Spec)
	}
}
