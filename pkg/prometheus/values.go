package prometheus

import (
	"fmt"

	promModel "github.com/prometheus/common/model"
)

// InstantVector reduces whatever an instant query answered with to the vector the caller asked
// for. The prometheus API contract says an instant query returns a vector, but the query path in
// front of it does not always agree: a matrix comes back for the same query often enough that
// treating it as unusable meters those orgs at zero, silently, for the hour.
//
// A matrix is the same answer with its history attached, so the sample to take is each series'
// last one at or before the instant — exactly what the vector would have held.
func InstantVector(value promModel.Value, at promModel.Time) (promModel.Vector, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case promModel.Vector:
		return typed, nil
	case promModel.Matrix:
		vector := make(promModel.Vector, 0, len(typed))
		for _, series := range typed {
			sample, ok := lastSampleAt(series, at)
			if !ok {
				continue
			}
			vector = append(vector, &promModel.Sample{
				Metric:    series.Metric,
				Value:     sample.Value,
				Timestamp: sample.Timestamp,
			})
		}
		return vector, nil
	default:
		// A scalar or a string carries no series to bill anything against, so there is nothing to
		// salvage and the caller has to know it asked the wrong question.
		return nil, fmt.Errorf("unusable instant query response of type %T", value)
	}
}

func lastSampleAt(series *promModel.SampleStream, at promModel.Time) (promModel.SamplePair, bool) {
	if series == nil || len(series.Values) == 0 {
		return promModel.SamplePair{}, false
	}
	for i := len(series.Values) - 1; i >= 0; i-- {
		if series.Values[i].Timestamp <= at {
			return series.Values[i], true
		}
	}
	// Every sample is newer than the instant asked about. Billing the hour with a value from after
	// it would charge for time that has not happened.
	return promModel.SamplePair{}, false
}
