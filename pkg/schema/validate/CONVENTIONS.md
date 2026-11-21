# pkg/validate — schema port CONVENTIONS

This is the contract every fan-out agent MUST follow when porting a
`nodelibs/schema/src/*.ts` Joi schema to Go. The foundation (engine + base
exemplar + parity harness) is done and green; you author resource schemas on top
of it.

**Golden rule: NO deferral, NO TODO, NO "port a subset", NO stubbing "for now".**
If a Joi feature is hard, the engine already implements it — use it. If you think
the engine is missing something, it almost certainly is not (see the feature list
below); confirm against a `*.test.ts` vector before assuming.

---

## 1. Where the engine lives (and why)

- **Engine**: `pkg/validate` (`rule.go` builders, `eval.go` runtime, `errors`
  via `*validate.Error`, `harness.go` test runner). It has **ZERO** dependency on
  `pkg/handlers` — do not create an import cycle.
- **Schema DEFINITIONS**: authored in `pkg/handlers` so they can call
  `links.go` (`expandLocalLink`, `expandImageLink`, `linkKindOK`, `kindsJSON`,
  …). The engine takes your `.custom()` logic as a `validate.CustomFn` closure;
  the closure lives in handlers and may call anything in handlers.
- **Reference exemplar**: `pkg/handlers/schema_base.go` (port of `base.ts`) and
  its test `pkg/handlers/schema_base_test.go` (port of `base.test.ts`). Copy their
  shape.

## 2. File layout (pick ONE — this is the standard)

One file per resource: **`pkg/handlers/schema_<kind>.go`** exposing a builder
`func <Kind>Schema() *validate.Rule` and an `init()` (or a call from the handler's
existing registration site) that does:

```go
func init() {
    validate.Register("workload", validate.ObjectSchema(WorkloadSchema()))
}
```

`validate.ObjectSchema(root)` wraps a `*Rule` as a `*validate.Schema{Root: root}`.
`base.go:runValidate` picks it up automatically on create AND patch (it runs on
the merged object), building the `validate.Context{Org, Gvc, Method}` for you and
rendering the parity error body.

Do NOT keep hand-rolling the legacy `validate.Schema{Fields: []FieldRule{…}}`
path. Migrate any resource you touch to a `Root` rule. The legacy path still
compiles (existing registrations in workload.go, identity.go, domain.go,
mk8sCluster.go, cloudaccount.go, location.go, image.go, memcacheCluster.go,
serviceaccount.go, agent.go, auditctx.go, user.go, volumeset.go) — replace it,
don't extend it.

## 3. Building a schema — the combinator API

Constructors: `String() Number() Boolean() DateRule() Any() Array(items…)
Object(map[string]*Rule) Alternatives(alts…)`.

Every fluent method returns a **copy** (Joi immutability), so shared sub-schema
symbols (e.g. `NameRule()`, `LocalLinkRule()`) are safe to re-decorate.

| Joi | Go |
|-----|-----|
| `.required()/.optional()/.forbidden()` | `.Required() .Optional() .Forbidden()` |
| `.default(v)` / `.default(Joi.ref('k'))` / `.default(()=>v)` | `.Default(v)` / `.Default(validate.RefTo("k"))` / `.Default(func() any{…})` |
| `.valid(a,b)` | `.Valid(a, b)` (any.only) |
| `.allow('')/.allow(null)` | `.Allow("")` / `.Allow(nil)` (verbatim bypass) |
| `.custom(fn)` | `.Custom(func(v any, h *validate.Helper)(any,error){…})` |
| `.meta({readOnly})/.meta({unmodifiable})` | `.ReadOnly()` / `.Unmodifiable()` |
| `.trim().lowercase().uppercase()` | `.Trim() .Lowercase() .Uppercase()` |
| `.min(n).max(n).length(n)` (string/number/array by type) | `.Min(n) .Max(n) .Length(n)` |
| number `.min(Joi.ref('k'))` | `.MinRef("k")` / `.MaxRef("k")`; float bounds `.NumMin(f)/.NumMax(f)` |
| `.integer().positive().negative().port().precision(n)` | `.Integer() .Positive() .Negative() .Port() .Precision(n)` |
| `.regex(re)` / `.pattern(re,{invert})` | `.Regex(re)` / `.RegexInvert(re)` |
| `.email().uri().uuid()/.guid().domain()` | `.Email() .URI() .UUID() .Domain()` |
| `.date().iso()` | `validate.DateRule().Iso()` (output = ISO-8601 **millis UTC** string) |
| array `.items(…).sort().unique()/.unique('k').single()` | `.Items(…) .Sort() .Unique()/.Unique("k")` (single-item coercion: wrap scalar yourself if needed) |
| object `.pattern(k,v)` | `.Pattern(keyRule, valRule)` |
| object `.unknown(true)` / `.unknown('strip')` | `.Unknown(true)` / `.StripUnknown()` |
| object `.xor(…).oxor(…).nand(…)` | `.Xor(…) .Oxor(…) .Nand(…)` |
| object `.keys(…)/.append(…)/.fork([k],fn)` | `.Keys(map) / .Append(map) / .Fork("k", fn)` |
| `.when(ref,{is,not,then,otherwise})` | `.When("ref", validate.When{Is:…, Not:…, Then:…, Otherwise:…})` |
| `Joi.alternatives().try(…)` | `validate.Alternatives(a, b)` or `.Try(a,b)` |
| `Joi.alternatives().conditional(ref,{is,then,otherwise})` | `validate.Alternatives().Conditional("ref", validate.AltCase{Is,Then,Otherwise})` |

Notes:
- `.description()` is metadata-only in Joi (codegen); **drop it** — it has no
  runtime effect. Same for `.tag()`, `.meta({hidden/private/pointer/genericField})`.
- A Joi `Joi.object()` **rejects unknown keys by default** (because the TS
  service validates with default prefs). The engine does the same. Only add
  `.Unknown(true)`/`.StripUnknown()` if the TS schema/callsite does.
- `String()` rejects `""` by default (Joi `string.empty`). Use `.Allow("")` when
  the TS schema does.

## 4. Shared sub-schemas — reference by symbol, never re-author

Import the owner's exported builder. From `schema_base.go`:
`NameRule() KindRule() CloudProviderRule() TagsRule() DescriptionRule()
LinkRule() LinksRule() LocalLinkRule() ImageLinkRule() ToKind(kinds…) RegexRule()
BaseSchema()`.

Extend base for a resource:

```go
func GvcSchema() *validate.Rule {
    return BaseSchema().Keys(map[string]*validate.Rule{
        "spec": gvcSpec(),
    })
}
```

**Links MUST reuse links.go.** Do not re-implement `//kind/name` expansion. Use
`LocalLinkRule()` / `ImageLinkRule()` and append `.Custom(ToKind("secret"))` for
the kind check — exactly like `base.ts` does `LocalLink.custom(toKind('secret'))`.
The central `applyLinkExpansion` pass (links.go / `linkRegistry`) still runs
BEFORE schema validation on the merged object, so by the time your schema sees a
link it is already absolute; a `LocalLinkRule()` field is then idempotent. If a
resource's links are already covered by `linkRegistry`, you still add the schema
field for shape/type validation but the expansion is a no-op.

## 5. Ownership — see OWNERSHIP.md

Own exactly the files assigned to your partition. If you need a symbol owned by
another partition (e.g. workload references `EnvVar` owned by A), **import it**,
do not fork it. Cross-references are enumerated in OWNERSHIP.md.

## 6. The parity harness — port every `*.test.ts`

Every schema file has a sibling `*.test.ts`. Port ALL its vectors into
`pkg/handlers/schema_<kind>_test.go` using `validate.Case`:

```go
func TestWorkload(t *testing.T) {
    runCases(t, WorkloadSchema(), []validate.Case{
        {Name: "good", Input: map[string]any{…}, Ctx: &validate.Context{Org:"o", Gvc:"g"},
         Want: map[string]any{…normalized…}},
        {Name: "bad", Input: map[string]any{…}, WantErr: "is required"},
    })
}
```

- `Input` uses `map[string]any` / `[]any` / `float64` (JSON shapes) — numbers are
  `float64`, matching what `json.Decode` produces at runtime.
- `WantErr` is a **substring** of the error message (Joi test idiom
  `expect(error.message).toContain(...)`).
- `Want` deep-equals the normalized output (`expect(result.value).toEqual(...)`).
- `Check` is an escape hatch for awkward assertions.
- `runCases` helper is in `schema_base_test.go`; copy it or define your own thin
  `t.Run` loop over `Case.Run(rule)`.

A vector that depends on a resource in another partition (e.g. base.test.ts uses
`Org`) should be reproduced with the minimal local rule that exercises the
base-level behaviour, OR skipped with a comment pointing at the owning partition's
test — never leave it unported silently.

## 7. Resolving ambiguity — use the oracle, don't guess

Primary oracle = the ported `*.test.ts` vectors. When a vector is ambiguous or a
Joi behaviour is unclear, run the live TS data-service and diff:

```bash
# TS oracle (repo convention :3000); Go port has run on :3491.
cd data-service && pnpm start           # or pnpm start:live
# POST the same body to both and diff status + error.details[].
```

Do not invent behaviour. If you find the engine genuinely cannot express a Joi
construct used by a real schema (not comments — envoy has many commented-out
`Joi.link` recursive bits that are NOT active), STOP and report it to the
foundation owner rather than stubbing.

## 8. Error shape (already handled for you)

A `*validate.Error` carries `Message`, `Type` (Joi error type, e.g.
`any.required`, `string.pattern.base`), and `Path` (segments; array indices are
numeric). `base.go:schemaValidationError` renders it to the parity HTTP body
`{status:400, message, details:[{message,path,type,context}]}` (isDataInvalid).
You only produce errors via the engine or `h.Errorf(type, fmt, …)` inside a
custom — never hand-build an `httpx.APIError` for a validation failure.

## 9. Keep it green

`GOFLAGS=-mod=mod go build ./... && GOFLAGS=-mod=mod go test ./...` must pass
after every resource you land. Do not commit.
