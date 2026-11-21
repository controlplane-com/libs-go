package base

import (
	"math"
	"strings"
)

// MemoryToMiB ports convertMemoryStringToNumber, which TS keeps three copies of
// (workload.ts, volumeSet.ts, nodelibs/misc resources.ts) — here it is one.
func MemoryToMiB(memory string) float64 {
	if memory == "" {
		return 0
	}
	parts := splitBeforeMemUnit(memory)
	if len(parts) != 2 {
		return math.Round(jsParseInt(parts[0]) / 1048576)
	}
	switch strings.ToLower(parts[1]) {
	case "ki":
		return math.Round(jsParseInt(parts[0]) / 1024)
	case "k":
		// decimal units floor so a value can never exceed the limit it is checked against
		return math.Floor(jsParseInt(parts[0]) / 1048.576)
	case "mi":
		return jsParseInt(parts[0])
	case "m":
		return math.Floor(jsParseInt(parts[0]) / 1.048576)
	case "gi":
		return jsParseFloat(parts[0]) * 1024.0
	case "g":
		return math.Floor(jsParseFloat(parts[0]) * 1000 / 1.048576)
	}
	return 0
}

// splitBeforeMemUnit is JS `split(/(?=[GMKgmk])/)`: a zero-width match at index 0 does
// not split, and a second unit letter yields three parts, which the caller rejects.
func splitBeforeMemUnit(s string) []string {
	out := []string{}
	start := 0
	for i := 1; i < len(s); i++ {
		if strings.IndexByte("GMKgmk", s[i]) >= 0 {
			out = append(out, s[start:i])
			start = i
		}
	}
	return append(out, s[start:])
}

// jsParseInt takes the leading integer prefix, as JS parseInt does: parseInt("1.5") is
// 1, so "1.5Mi" is one MiB rather than zero.
func jsParseInt(s string) float64 {
	return jsParseNum(s, false)
}

func jsParseFloat(s string) float64 {
	return jsParseNum(s, true)
}

func jsParseNum(s string, frac bool) float64 {
	s = strings.TrimLeft(s, " \t\n\r")
	i, neg := 0, false
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		neg = s[i] == '-'
		i++
	}
	whole := 0.0
	digits := 0
	for ; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		whole = whole*10 + float64(s[i]-'0')
		digits++
	}
	if frac && i < len(s) && s[i] == '.' {
		i++
		scale := 0.1
		for ; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
			whole += float64(s[i]-'0') * scale
			scale /= 10
			digits++
		}
	}
	if digits == 0 {
		return math.NaN() // JS parseInt("Gi")
	}
	if neg {
		return -whole
	}
	return whole
}
