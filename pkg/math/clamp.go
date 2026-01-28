package math

import "math"

func Clamp(x, min, max float64) float64 {
	return math.Max(min, math.Min(max, x))
}
