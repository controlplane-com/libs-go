package math

import (
	"encoding/json"
	"math"
)

type Float64 float64

func (f Float64) MarshalJSON() ([]byte, error) {
	switch {
	case math.IsNaN(float64(f)):
		return []byte(`"NaN"`), nil
	case math.IsInf(float64(f), 1):
		return []byte(`"Infinity"`), nil
	case math.IsInf(float64(f), -1):
		return []byte(`"-Infinity"`), nil
	default:
		return json.Marshal(float64(f)) // default serialization for normal values
	}
}

func Inf(sign int) Float64 {
	return Float64(math.Inf(sign))
}

func NaN() Float64 {
	return Float64(math.NaN())
}

func IsNaN(f Float64) bool {
	return math.IsNaN(float64(f))
}

func IsInf(f Float64, sign int) bool {
	return math.IsInf(float64(f), sign)
}

func IsReal(f Float64) bool {
	return !IsNaN(f) && !IsInf(f, 1) && !IsInf(f, -1)
}

const epsilon Float64 = 0.0001

func Abs(f Float64) Float64 {
	return Float64(math.Abs(float64(f)))
}

func Equal(f1, f2 Float64) bool {
	return EqualWithEpsilon(f1, f2, epsilon)
}

func EqualWithEpsilon(f1, f2 Float64, epsilon Float64) bool {
	return Abs(f2-f1) < epsilon
}
