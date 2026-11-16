package math

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloat64JSONRoundTrip(t *testing.T) {
	for _, v := range []Float64{0, 1.5, -2, Inf(1), Inf(-1), NaN()} {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		var got Float64
		require.NoError(t, json.Unmarshal(b, &got), string(b))
		if IsNaN(v) {
			require.True(t, IsNaN(got))
			continue
		}
		require.Equal(t, v, got)
	}
}

func TestFloat64UnmarshalInStruct(t *testing.T) {
	var d struct {
		From Float64 `json:"from"`
		To   Float64 `json:"to"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"from":100,"to":"Infinity"}`), &d))
	require.Equal(t, Float64(100), d.From)
	require.True(t, math.IsInf(float64(d.To), 1))
}

func TestFloat64UnmarshalRejectsOtherStrings(t *testing.T) {
	var f Float64
	require.Error(t, json.Unmarshal([]byte(`"12"`), &f))
	require.Error(t, json.Unmarshal([]byte(`true`), &f))
}

func TestFloat64UnmarshalNullLeavesZero(t *testing.T) {
	var f Float64
	require.NoError(t, json.Unmarshal([]byte(`null`), &f))
	require.Equal(t, Float64(0), f)
}
