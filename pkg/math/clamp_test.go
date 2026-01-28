package math

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

type ClampTest struct {
	X              float64 `json:"x"`
	Min            float64 `json:"min"`
	Max            float64 `json:"max"`
	ExpectedResult float64 `json:"expectedResult"`
}

func TestClamp(t *testing.T) {
	negInf := math.Inf(-1)
	inf := math.Inf(1)
	tests := []ClampTest{
		{
			X:              10,
			Min:            negInf,
			Max:            inf,
			ExpectedResult: 10,
		},
		{
			X:              10,
			Min:            0,
			Max:            inf,
			ExpectedResult: 10,
		},
		{
			X:              10,
			Min:            negInf,
			Max:            5,
			ExpectedResult: 5,
		},
		{
			X:              10,
			Min:            5,
			Max:            50,
			ExpectedResult: 10,
		},
		{
			X:              1,
			Min:            5,
			Max:            50,
			ExpectedResult: 5,
		},
		{
			X:              51,
			Min:            5,
			Max:            50,
			ExpectedResult: 50,
		},
	}
	for _, test := range tests {
		result := Clamp(test.X, test.Min, test.Max)
		if result != test.ExpectedResult {
			j, _ := json.MarshalIndent(test, "", "   ")
			fmt.Printf("With test: %s.\nExpected %f, got %f\n", string(j), test.ExpectedResult, result)
			t.FailNow()
		}
	}
}
