package userSettings

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestUserSettingsMatchesJoi pins UserSettings to verdicts captured from the built
// userSettings.ts (node over build/cjs/src/userSettings.js).
func TestUserSettingsMatchesJoi(t *testing.T) {
	long := strings.Repeat("a", 65)
	for _, c := range []struct {
		name string
		body string
		typ  string // "" = accepted
		msg  string
		want string // normalized JSON when accepted; "" = the body itself
	}{
		{name: "arbitrary nested content under global and orgs.<org> is accepted as-is",
			body: `{"global":{"theme":"dark","nested":{"list":[1,"two",null],"deep":{"x":true}}},"orgs":{"coke":{"pinned":["gvc/a"],"filters":{"kind":"workload"}},"pepsi-2":{}}}`},
		{name: "an empty document", body: `{}`},
		{name: "the echoed read envelope",
			body: `{"kind":"userSettings","version":0,"links":[{"rel":"self","href":"/user/a@example.com/settings"}],"global":{},"orgs":{}}`},
		{name: "a string version converts", body: `{"version":"3"}`, want: `{"version":3}`},
		{name: "a fractional version", body: `{"version":1.5}`, typ: "number.integer", msg: `"version" must be an integer`},
		{name: "a negative version", body: `{"version":-1}`, typ: "number.min", msg: `"version" must be greater than or equal to 0`},
		{name: "a boolean version", body: `{"version":true}`, typ: "number.base", msg: `"version" must be a number`},
		{name: "a non-numeric version", body: `{"version":"x"}`, typ: "number.base", msg: `"version" must be a number`},
		{name: "an unknown top-level key", body: `{"theme":"dark"}`, typ: "object.unknown", msg: `"theme" is not allowed`},
		{name: "global is a string", body: `{"global":"dark"}`, typ: "object.base", msg: `"global" must be of type object`},
		{name: "global is an array", body: `{"global":[]}`, typ: "object.base", msg: `"global" must be of type object`},
		{name: "global is null", body: `{"global":null}`, typ: "object.base", msg: `"global" must be of type object`},
		{name: "orgs is an array", body: `{"orgs":[]}`, typ: "object.base", msg: `"orgs" must be of type object`},
		{name: "an org entry is a string", body: `{"orgs":{"coke":"x"}}`, typ: "object.base", msg: `"orgs.coke" must be of type object`},
		{name: "an org entry is null", body: `{"orgs":{"coke":null}}`, typ: "object.base", msg: `"orgs.coke" must be of type object`},
		{name: "org key Coke", body: `{"orgs":{"Coke":{}}}`, typ: "object.unknown", msg: `"orgs.Coke" is not allowed`},
		{name: "org key has space", body: `{"orgs":{"has space":{}}}`, typ: "object.unknown", msg: `"orgs.has space" is not allowed`},
		{name: "org key xserve-coke", body: `{"orgs":{"xserve-coke":{}}}`, typ: "object.unknown", msg: `"orgs.xserve-coke" is not allowed`},
		{name: "org key index", body: `{"orgs":{"index":{}}}`, typ: "object.unknown", msg: `"orgs.index" is not allowed`},
		{name: "org key -coke", body: `{"orgs":{"-coke":{}}}`, typ: "object.unknown", msg: `"orgs.-coke" is not allowed`},
		{name: "org key too long", body: `{"orgs":{"` + long + `":{}}}`, typ: "object.unknown", msg: `"orgs.` + long + `" is not allowed`},
		{name: "another kind", body: `{"kind":"user"}`, typ: "any.only", msg: `"kind" must be [userSettings]`},
		{name: "an unparseable created", body: `{"created":"garbage"}`, typ: "date.base", msg: `"created" must be a valid date`},
		{name: "created normalizes", body: `{"created":"2026-01-01T00:00:00Z"}`, want: `{"created":"2026-01-01T00:00:00.000Z"}`},
		{name: "a numeric lastModified", body: `{"lastModified":12345}`, want: `{"lastModified":"1970-01-01T00:00:12.345Z"}`},
		{name: "a link without href", body: `{"links":[{"rel":"self"}]}`, typ: "any.required", msg: `"links[0].href" is required`},
	} {
		t.Run(c.name, func(t *testing.T) {
			var obj map[string]any
			if err := json.Unmarshal([]byte(c.body), &obj); err != nil {
				t.Fatal(err)
			}
			verr := validate.For[UserSettings]().ApplyCtx(obj, &validate.Context{}, validate.Prefs{})
			gotTyp, gotMsg := "", ""
			if verr != nil {
				d := verr.(*validate.Error).Details()
				gotTyp, gotMsg = d[0].Type, d[0].Message
			}
			if gotTyp != c.typ || gotMsg != c.msg {
				t.Fatalf("got %q %q, want %q %q", gotTyp, gotMsg, c.typ, c.msg)
			}
			if c.typ != "" {
				return
			}
			want := c.want
			if want == "" {
				want = c.body
			}
			var wantObj map[string]any
			_ = json.Unmarshal([]byte(want), &wantObj)
			got, _ := json.Marshal(obj)
			exp, _ := json.Marshal(wantObj)
			if string(got) != string(exp) {
				t.Fatalf("value:\n got  %s\n want %s", got, exp)
			}
		})
	}
}
