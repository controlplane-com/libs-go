package prometheus

import (
	"testing"
	"time"

	promModel "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func at(t time.Time) promModel.Time { return promModel.TimeFromUnixNano(t.UnixNano()) }

var instant = time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)

func series(org string, samples ...promModel.SamplePair) *promModel.SampleStream {
	return &promModel.SampleStream{
		Metric: promModel.Metric{"org": promModel.LabelValue(org)},
		Values: samples,
	}
}

func sample(t time.Time, value float64) promModel.SamplePair {
	return promModel.SamplePair{Timestamp: at(t), Value: promModel.SampleValue(value)}
}

// A matrix is the same answer with its history attached. Dropping it meters the org at zero for
// the hour, which is what was happening to hundreds of orgs an hour.
func TestInstantVectorTakesTheSampleAtTheInstantFromAMatrix(t *testing.T) {
	matrix := promModel.Matrix{
		series("a", sample(instant.Add(-2*time.Hour), 1), sample(instant.Add(-time.Hour), 2), sample(instant, 3)),
		series("b", sample(instant.Add(-time.Hour), 7)),
	}

	vector, err := InstantVector(matrix, at(instant))
	require.NoError(t, err)
	require.Len(t, vector, 2)
	assert.Equal(t, promModel.SampleValue(3), vector[0].Value, "the value as of the instant, not the first one")
	assert.Equal(t, at(instant), vector[0].Timestamp)
	assert.Equal(t, promModel.SampleValue(7), vector[1].Value, "the last sample at or before the instant")
	assert.Equal(t, promModel.LabelValue("a"), vector[0].Metric["org"])
}

// Billing an hour with a value from after it would charge for time that has not happened.
func TestInstantVectorIgnoresSamplesAfterTheInstant(t *testing.T) {
	matrix := promModel.Matrix{
		series("a", sample(instant.Add(time.Hour), 99)),
		series("b", sample(instant.Add(-time.Minute), 4), sample(instant.Add(time.Hour), 99)),
	}

	vector, err := InstantVector(matrix, at(instant))
	require.NoError(t, err)
	require.Len(t, vector, 1, "a series with nothing at or before the instant contributes nothing")
	assert.Equal(t, promModel.SampleValue(4), vector[0].Value)
}

func TestInstantVectorPassesAVectorThrough(t *testing.T) {
	vector := promModel.Vector{{Metric: promModel.Metric{"org": "a"}, Value: 5, Timestamp: at(instant)}}
	got, err := InstantVector(vector, at(instant))
	require.NoError(t, err)
	assert.Equal(t, vector, got)
}

func TestInstantVectorHandlesNothingAtAll(t *testing.T) {
	empty, err := InstantVector(promModel.Matrix{}, at(instant))
	require.NoError(t, err)
	assert.Empty(t, empty)

	none, err := InstantVector(nil, at(instant))
	require.NoError(t, err)
	assert.Empty(t, none)

	emptySeries, err := InstantVector(promModel.Matrix{series("a")}, at(instant))
	require.NoError(t, err)
	assert.Empty(t, emptySeries)
}

// A scalar carries no series to bill against. There is nothing to salvage, and pretending it was
// empty is how an hour goes unbilled without anyone noticing.
func TestInstantVectorRefusesAnAnswerWithNoSeries(t *testing.T) {
	_, err := InstantVector(&promModel.Scalar{Value: 1, Timestamp: at(instant)}, at(instant))
	require.ErrorContains(t, err, "unusable instant query response")

	_, err = InstantVector(&promModel.String{Value: "nope", Timestamp: at(instant)}, at(instant))
	require.ErrorContains(t, err, "unusable instant query response")
}
