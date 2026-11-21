package validate

import (
	"encoding/json"
	"regexp"
	"testing"
)

// The expectations here are Joi's own `schema.describe()` output for the
// equivalent TS schema (taken from the data-service oracle's
// GET /catalog/schema/{kind}), so they pin key ORDER as well as content.
func describeJSON(t *testing.T, r *Rule) string {
	t.Helper()
	b, err := json.Marshal(r.Describe())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func TestDescribeScalarRules(t *testing.T) {
	cases := []struct {
		name string
		rule *Rule
		want string
	}{{
		name: "string trim+max+allow empty (Base.description)",
		rule: String().Trim().Max(250).Allow(""),
		want: `{"type":"string","rules":[{"name":"trim","args":{"enabled":true}},{"name":"max","args":{"limit":250}}],"allow":[""]}`,
	}, {
		name: "uuid + readOnly meta (Base.id)",
		rule: String().UUID().ReadOnly(),
		want: `{"type":"string","rules":[{"name":"guid"}],"metas":[{"readOnly":true}]}`,
	}, {
		name: "positive integer + readOnly (Base.version)",
		rule: Number().Positive().Integer().ReadOnly(),
		want: `{"type":"number","rules":[{"name":"sign","args":{"sign":"positive"}},{"name":"integer"}],"metas":[{"readOnly":true}]}`,
	}, {
		name: "valid() sets only + allow",
		rule: String().Valid("plain", "base64").Default("plain").Optional(),
		want: `{"type":"string","flags":{"only":true,"default":"plain","presence":"optional"},"allow":["plain","base64"]}`,
	}, {
		name: "pattern renders a JS regex literal",
		rule: String().Regex(regexp.MustCompile(`^[a-z]([-a-z0-9])*[a-z0-9]$`)).Max(64).Unmodifiable(),
		want: `{"type":"string","rules":[{"name":"pattern","args":{"regex":"/^[a-z]([-a-z0-9])*[a-z0-9]$/"}},{"name":"max","args":{"limit":64}}],"metas":[{"unmodifiable":true}]}`,
	}, {
		name: "pattern escapes the slashes that would end the literal",
		rule: String().Trim().Regex(regexp.MustCompile(`(/org/[^/]+/.*)|(//.+)`)),
		want: `{"type":"string","rules":[{"name":"trim","args":{"enabled":true}},{"name":"pattern","args":{"regex":"/(\\/org\\/[^/]+\\/.*)|(\\/\\/.+)/"}}]}`,
	}, {
		name: "case-insensitive regex becomes the i flag",
		rule: String().Regex(regexp.MustCompile(`(?i)^sha(256|512):[a-f0-9]+$`)),
		want: `{"type":"string","rules":[{"name":"pattern","args":{"regex":"/^sha(256|512):[a-f0-9]+$/i"}}]}`,
	}, {
		name: "description + label flags",
		rule: Boolean().Description("Enable it").Required(),
		want: `{"type":"boolean","flags":{"presence":"required","description":"Enable it"}}`,
	}, {
		name: "numeric bound by ref",
		rule: Number().MaxRef("maxScale"),
		want: `{"type":"number","rules":[{"name":"max","args":{"limit":{"ref":{"path":["maxScale"]}}}}]}`,
	}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := describeJSON(t, c.rule); got != c.want {
				t.Errorf("describe mismatch\n got: %s\nwant: %s", got, c.want)
			}
		})
	}
}

func TestDescribeContainers(t *testing.T) {
	// Base.links — an array of the ordered rel/href object.
	links := Array(ObjectOrdered([]string{"rel", "href"}, map[string]*Rule{
		"rel":  String().Required().Trim(),
		"href": String().Required().Trim(),
	}))
	want := `{"type":"array","items":[{"type":"object","keys":{"rel":{"type":"string","flags":{"presence":"required"},"rules":[{"name":"trim","args":{"enabled":true}}]},"href":{"type":"string","flags":{"presence":"required"},"rules":[{"name":"trim","args":{"enabled":true}}]}}}]}`
	if got := describeJSON(t, links); got != want {
		t.Errorf("links describe mismatch\n got: %s\nwant: %s", got, want)
	}

	// Base.tags — a pattern map with a custom; a function default is unencodable,
	// so Joi reports the flags object with no `default` key at all.
	tags := Object(nil).Pattern(String(), Any()).
		CustomTyped(func(v any, _ *Helper) (any, error) { return v, nil }).
		Meta(map[string]any{"type": "Tags"}).
		Default(func() any { return map[string]any{} })
	want = `{"type":"object","flags":{},"rules":[{"name":"custom","args":{}}],"metas":[{"type":"Tags"}],"patterns":[{"schema":{"type":"string"},"rule":{"type":"any"}}]}`
	if got := describeJSON(t, tags); got != want {
		t.Errorf("tags describe mismatch\n got: %s\nwant: %s", got, want)
	}

	// sort({by}) reports the FULL Joi ref descriptor, unlike a min/max limit.
	env := Array(Object(map[string]*Rule{"name": String()})).SortBy("name").Dedupe("name")
	want = `{"type":"array","rules":[{"name":"sort","args":{"options":{"order":"ascending","by":{"adjust":null,"in":false,"iterables":null,"map":null,"separator":".","type":"value","ancestor":0,"path":["name"],"depth":1,"key":"name","root":"name","display":"ref:.name"}}}},{"name":"custom","args":{}}],"items":[{"type":"object","keys":{"name":{"type":"string"}}}]}`
	if got := describeJSON(t, env); got != want {
		t.Errorf("sort-by describe mismatch\n got: %s\nwant: %s", got, want)
	}

	// .unique() takes no args; .unique(by) carries the (unencodable) comparator.
	if got := describeJSON(t, Array(String()).Max(3).Unique()); got != `{"type":"array","rules":[{"name":"max","args":{"limit":3}},{"name":"unique"}],"items":[{"type":"string"}]}` {
		t.Errorf("unique describe mismatch: %s", got)
	}

	// dependencies precede keys.
	xor := Object(map[string]*Rule{"recommended": String(), "exact": String()}).Xor("recommended", "exact")
	want = `{"type":"object","dependencies":[{"rel":"xor","peers":["recommended","exact"]}],"keys":{"exact":{"type":"string"},"recommended":{"type":"string"}}}`
	if got := describeJSON(t, xor); got != want {
		t.Errorf("xor describe mismatch\n got: %s\nwant: %s", got, want)
	}
}

// A literal `is:` in TS compiles to any().valid(override, lit).required(); the Go
// builders spell it as a typed one-value .Valid(), so Describe reconstructs it.
func TestDescribeConditionalIsLiteral(t *testing.T) {
	alt := Alternatives().Conditional("encoding", AltCase{
		Is:        String().Valid("base64"),
		Then:      String().Required(),
		Otherwise: String().Required(),
	})
	want := `{"type":"alternatives","matches":[{"ref":{"path":["encoding"]},"is":{"type":"any","flags":{"only":true,"presence":"required"},"allow":[{"override":true},"base64"]},"then":{"type":"string","flags":{"presence":"required"}},"otherwise":{"type":"string","flags":{"presence":"required"}}}]}`
	if got := describeJSON(t, alt); got != want {
		t.Errorf("conditional describe mismatch\n got: %s\nwant: %s", got, want)
	}

	// A multi-value `is` was written as a real schema in TS too, so it stays itself.
	when := Any().When("op", When{
		Is:   Any().Valid("=", "!="),
		Then: Any().Required(),
	})
	want = `{"type":"any","whens":[{"ref":{"path":["op"]},"is":{"type":"any","flags":{"only":true},"allow":["=","!="]},"then":{"type":"any","flags":{"presence":"required"}}}]}`
	if got := describeJSON(t, when); got != want {
		t.Errorf("when describe mismatch\n got: %s\nwant: %s", got, want)
	}
}

// Rules are reported in APPLICATION order, which the flattened Rule fields lose.
func TestDescribeRuleOrderFollowsApplication(t *testing.T) {
	a := describeJSON(t, String().Trim().Max(10))
	b := describeJSON(t, String().Max(10).Trim())
	if a == b {
		t.Fatalf("rule order not recorded: %s", a)
	}
	if want := `{"type":"string","rules":[{"name":"trim","args":{"enabled":true}},{"name":"max","args":{"limit":10}}]}`; a != want {
		t.Errorf("got %s want %s", a, want)
	}
}

// A fluent call must not append to a shared rule's slice (Joi schemas are immutable).
func TestDescribeCloneIsolatesRules(t *testing.T) {
	base := String().Trim()
	_ = base.Max(1)
	_ = base.Max(2)
	if got := describeJSON(t, base); got != `{"type":"string","rules":[{"name":"trim","args":{"enabled":true}}]}` {
		t.Errorf("base rule mutated: %s", got)
	}
}

// A hand-built Object(map) loses its source order; attaching it to a struct type
// restores the declaration order Joi reports (and abortEarly blames).
type orderProbe struct {
	Region string `json:"region"`
	Tags   string `json:"awsTags"`
	Image  string `json:"image"`
}

func (orderProbe) ValidationRule() *Rule {
	return Object(map[string]*Rule{
		"region":  String(),
		"awsTags": String(),
		"image":   String(),
	})
}

func TestDescribeKeyOrderFollowsStruct(t *testing.T) {
	want := `{"type":"object","keys":{"region":{"type":"string"},"awsTags":{"type":"string"},"image":{"type":"string"}}}`
	if got := describeJSON(t, For[orderProbe]().Rule()); got != want {
		t.Errorf("key order not recovered\n got: %s\nwant: %s", got, want)
	}
}
