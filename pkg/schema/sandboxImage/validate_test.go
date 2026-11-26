package sandboxImage

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestSpecMatchesJoi pins SandboxImageSpec to verdicts CAPTURED from the built
// sandboxImage.ts (node -e over build/cjs/src/sandboxImage.js) with the org context
// that expands an ImageLink. Every case is a recorded (type, message) pair.
func TestSpecMatchesJoi(t *testing.T) {
	long := strings.Repeat("a", 128)
	const pat = `"tag" with value %s fails to match the required pattern: ` +
		`/^[a-zA-Z0-9_][a-zA-Z0-9_\-.]*$/`
	for _, c := range []struct {
		body string
		typ  string // "" = accepted
		msg  string
		want string // normalized JSON when accepted
	}{
		{body: `{}`, want: `{"tag":"latest","type":"dev"}`},
		{body: `{"base":{"image":"coke.registry.cpln.io/app:1"}}`,
			want: `{"base":{"image":"coke.registry.cpln.io/app:1"},"tag":"latest","type":"dev"}`},
		{body: `{"base":{"image":"//image/base:1"}}`,
			want: `{"base":{"image":"/org/o/image/base:1"},"tag":"latest","type":"dev"}`},
		{body: `{"base":{"image":"//image/base"}}`,
			want: `{"base":{"image":"/org/o/image/base"},"tag":"latest","type":"dev"}`},
		{body: `{"base":{}}`, want: `{"base":{},"tag":"latest","type":"dev"}`},
		{body: `{"base":{"image":"x/y:1"},"tag":"v2"}`,
			want: `{"base":{"image":"x/y:1"},"tag":"v2","type":"dev"}`},
		{body: `{"tag":"_ok-1.2"}`, want: `{"tag":"_ok-1.2","type":"dev"}`},
		{body: `{"tag":"` + long + `"}`, want: `{"tag":"` + long + `","type":"dev"}`},
		{body: `{"type":"session"}`, want: `{"tag":"latest","type":"session"}`},
		// the old spelling is read as session by the handler, before the schema pass
		{body: `{"type":"agent"}`, typ: "any.only", msg: `"type" must be one of [dev, session]`},
		{body: `{"type":"bogus"}`, typ: "any.only", msg: `"type" must be one of [dev, session]`},
		{body: `{"tag":"bad tag!"}`, typ: "string.pattern.base", msg: sprintfPat(pat, `"bad tag!"`)},
		{body: `{"tag":"-lead"}`, typ: "string.pattern.base", msg: sprintfPat(pat, `"-lead"`)},
		{body: `{"tag":""}`, typ: "string.empty", msg: `"tag" is not allowed to be empty`},
		{body: `{"tag":"` + long + `a"}`, typ: "string.max",
			msg: `"tag" length must be less than or equal to 128 characters long`},
		// the recipe surface is gone: runtimes/packages are unknown keys. One at a time,
		// because the engine still blames multiple unknown keys lexicographically.
		{body: `{"base":{"image":"x/y:1"},"runtimes":[{"name":"python"}]}`,
			typ: "object.unknown", msg: `"runtimes" is not allowed`},
		{body: `{"base":{"image":"x/y:1"},"packages":["jq"]}`,
			typ: "object.unknown", msg: `"packages" is not allowed`},
		{body: `{"base":{"image":123}}`, typ: "string.base", msg: `"base.image" must be a string`},
	} {
		row, verr := validate.Decode[SandboxImageSpec](validate.Bytes(c.body), &validate.Context{Org: "o"}, validate.Prefs{})
		gotTyp, gotMsg := "", ""
		if verr != nil {
			gotTyp, gotMsg = verr.Details()[0].Type, verr.Details()[0].Message
		}
		if gotTyp != c.typ || gotMsg != c.msg {
			t.Errorf("%s:\n got  %q %q\n want %q %q", c.body, gotTyp, gotMsg, c.typ, c.msg)
			continue
		}
		if c.typ != "" {
			continue
		}
		out, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.want {
			t.Errorf("%s:\n got  %s\n want %s", c.body, out, c.want)
		}
	}
}

// TestStatusMatchesJoi pins SandboxImageStatus. lastBuildAt is a bare Joi.date(), so a
// non-ISO string is `date.base` (not `date.format`) yet still renders as ISO millis.
func TestStatusMatchesJoi(t *testing.T) {
	for _, c := range []struct {
		body string
		typ  string
		msg  string
		want string
	}{
		{body: `{}`, want: `{}`},
		{body: `{"phase":"ready"}`, want: `{"phase":"ready"}`},
		{body: `{"imageLink":"//image/foo:1"}`, want: `{"imageLink":"/org/o/image/foo:1"}`},
		{body: `{"imageLink":"/org/o/image/foo:1"}`, want: `{"imageLink":"/org/o/image/foo:1"}`},
		{body: `{"imageRef":"r"}`, want: `{"imageRef":"r"}`},
		{body: `{"buildId":"b"}`, want: `{"buildId":"b"}`},
		{body: `{"builtFingerprint":"f"}`, want: `{"builtFingerprint":"f"}`},
		// Joi answers {"error":""}: `allow('')` makes the empty string legal, but the
		// branch-wide `allow=` + omitzero convention drops it. See the report.
		{body: `{"error":""}`, want: `{}`},
		{body: `{"error":"boom"}`, want: `{"error":"boom"}`},
		{body: `{"buildDuration":12.5}`, want: `{"buildDuration":12.5}`},
		{body: `{"lastBuildAt":"2026-01-02T03:04:05Z"}`, want: `{"lastBuildAt":"2026-01-02T03:04:05.000Z"}`},
		{body: `{"lastBuildAt":"2026-01-02"}`, want: `{"lastBuildAt":"2026-01-02T00:00:00.000Z"}`},
		{body: `{"phase":"compiling"}`, typ: "any.only",
			msg: `"phase" must be one of [pending, queued, building, ready, failed]`},
		{body: `{"phase":""}`, typ: "any.only",
			msg: `"phase" must be one of [pending, queued, building, ready, failed]`},
		{body: `{"buildDuration":"x"}`, typ: "number.base", msg: `"buildDuration" must be a number`},
		{body: `{"lastBuildAt":"nope"}`, typ: "date.base", msg: `"lastBuildAt" must be a valid date`},
	} {
		row, verr := validate.Decode[SandboxImageStatus](validate.Bytes(c.body), &validate.Context{Org: "o"}, validate.Prefs{})
		gotTyp, gotMsg := "", ""
		if verr != nil {
			gotTyp, gotMsg = verr.Details()[0].Type, verr.Details()[0].Message
		}
		if gotTyp != c.typ || gotMsg != c.msg {
			t.Errorf("%s:\n got  %q %q\n want %q %q", c.body, gotTyp, gotMsg, c.typ, c.msg)
			continue
		}
		if c.typ != "" {
			continue
		}
		out, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.want {
			t.Errorf("%s:\n got  %s\n want %s", c.body, out, c.want)
		}
	}
}

// TestSandboxImageMatchesJoi pins the resource schema: spec is required, tags default
// to {}, and the create fork requires `name`.
func TestSandboxImageMatchesJoi(t *testing.T) {
	for _, c := range []struct {
		body string
		typ  string
		msg  string
		want string
	}{
		{body: `{"name":"si1","spec":{}}`, want: `{"name":"si1","tags":{},"spec":{"tag":"latest","type":"dev"}}`},
		{body: `{"spec":{}}`, want: `{"tags":{},"spec":{"tag":"latest","type":"dev"}}`},
		{body: `{"name":"si1","spec":{},"status":{}}`,
			want: `{"name":"si1","tags":{},"spec":{"tag":"latest","type":"dev"},"status":{}}`},
		{body: `{"name":"si1"}`, typ: "any.required", msg: `"spec" is required`},
	} {
		row, verr := validate.Decode[SandboxImage](validate.Bytes(c.body), &validate.Context{Org: "o"}, validate.Prefs{})
		gotTyp, gotMsg := "", ""
		if verr != nil {
			gotTyp, gotMsg = verr.Details()[0].Type, verr.Details()[0].Message
		}
		if gotTyp != c.typ || gotMsg != c.msg {
			t.Errorf("%s:\n got  %q %q\n want %q %q", c.body, gotTyp, gotMsg, c.typ, c.msg)
			continue
		}
		if c.typ != "" {
			continue
		}
		out, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.want {
			t.Errorf("%s:\n got  %s\n want %s", c.body, out, c.want)
		}
	}
}

func sprintfPat(pat, value string) string {
	return strings.Replace(pat, "%s", value, 1)
}
