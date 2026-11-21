package wirejson

import "testing"

// A valid surrogate pair is how every ASCII-safe JSON encoder (python json.dumps by
// default, Java, JS JSON.stringify with an ASCII filter) writes a non-BMP character.
// The pair branch used to skip a hex digit, so these failed to parse at all.
func TestSurrogatePairs(t *testing.T) {
	cases := map[string]string{
		"\"\\ud83d\\ude00\"":   "\U0001f600",
		"\"a\\ud83d\\ude00b\"": "a\U0001f600b",
		"\"\\u00e9\"":          "\u00e9",
		"\"\\ud83d\"":          "\ufffd",
		"\"\\ud83dx\"":         "\ufffdx",
	}
	for in, want := range cases {
		var v any
		if err := Decode([]byte(in), &v); err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		got, _ := v.(string)
		if got != want {
			t.Errorf("%s -> %q, want %q", in, got, want)
		}
	}
}
