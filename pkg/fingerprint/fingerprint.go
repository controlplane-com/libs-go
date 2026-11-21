// Package fingerprint ports nodelibs/misc fingerprint.ts: hashes of a value's canonical (RFC 8785) JSON.
package fingerprint

import (
	"bytes"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"strings"
)

var crockford = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ").WithPadding(base32.NoPadding)

// Canonical is the value's JSON with sorted keys and JSON.stringify's string escaping, as canonicalize writes it.
func Canonical(v any) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(generic); err != nil {
		return nil
	}
	return unescapeLineSeparators(bytes.TrimSuffix(buf.Bytes(), []byte("\n")))
}

// unescapeLineSeparators writes U+2028/U+2029 raw, as JSON.stringify does; Go always escapes them.
func unescapeLineSeparators(in []byte) []byte {
	if !bytes.Contains(in, []byte(`\u202`)) {
		return in
	}
	out := make([]byte, 0, len(in))
	for i := 0; i < len(in); i++ {
		if in[i] != '\\' || i+1 >= len(in) {
			out = append(out, in[i])
			continue
		}
		if esc := in[i:min(i+6, len(in))]; string(esc) == `\u2028` || string(esc) == `\u2029` {
			if esc[5] == '8' {
				out = append(out, "\u2028"...)
			} else {
				out = append(out, "\u2029"...)
			}
			i += 5
			continue
		}
		out = append(out, in[i], in[i+1])
		i++
	}
	return out
}

// Value is fingerprintValue: base64 sha1 of the canonical JSON, its padding '=' dropped.
func Value(v any) string {
	sum := sha1.Sum(Canonical(v))
	return strings.Replace(base64.StdEncoding.EncodeToString(sum[:]), "=", "", 1)
}

// ValueURL is fingerprintValueUrl: lowercase Crockford base32 of the sha1, cut to maxLength when it is positive.
func ValueURL(v any, maxLength int) string {
	sum := sha1.Sum(Canonical(v))
	hash := strings.ToLower(crockford.EncodeToString(sum[:]))
	if maxLength > 0 && maxLength < len(hash) {
		hash = hash[:maxLength]
	}
	return hash
}
