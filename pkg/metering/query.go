package metering

import (
	"errors"
	"github.com/controlplane-com/libs-go/pkg/math"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"time"
)

type AggregationType string

const (
	AggregateByTimeStep     = "timeStep"
	GroupByConsumptionQuery = "consumptionQuery"
)

type QueryRequest struct {
	StartTime           time.Time          `json:"startTime"`
	EndTime             time.Time          `json:"endTime"`
	TimeStep            timeUtils.TimeStep `json:"timeStep"`
	GroupBy             []string           `json:"groupBy,omitempty"`
	AggregateByTimeStep bool               `json:"aggregateByTimeStep"`

	ConsumptionQueries []*ConsumptionQuery `json:"consumptionQueries"`
}

func (q *QueryRequest) TimeSegment() timeUtils.TimeSegment {
	return timeUtils.TimeSegment{
		Start: q.StartTime,
		End:   q.EndTime,
	}
}

func (q *QueryRequest) SetTimeSegment(t timeUtils.TimeSegment) error {
	if t.End.Before(t.Start) {
		return errors.New("invalid time segment. End time cannot be before start time")
	}
	q.StartTime = t.Start
	q.EndTime = t.End
	return nil
}

func (q *QueryRequest) Clone() *QueryRequest {
	cloned := &QueryRequest{
		StartTime:           q.StartTime,
		EndTime:             q.EndTime,
		TimeStep:            q.TimeStep,
		GroupBy:             make([]string, len(q.GroupBy)),
		AggregateByTimeStep: q.AggregateByTimeStep,
	}
	copy(cloned.GroupBy, q.GroupBy)
	for _, c := range q.ConsumptionQueries {
		cloned.ConsumptionQueries = append(cloned.ConsumptionQueries, c.Clone())
	}
	return cloned
}

func matchGroupKeys(thisKeyMap map[string]any, otherKeyMap map[string]any) bool {
	if thisKeyMap == nil && otherKeyMap == nil {
		return true
	}

	for otherKey, otherValue := range otherKeyMap {
		if v, ok := thisKeyMap[otherKey]; !ok || v != otherValue {
			return false
		}
	}
	return true
}

type ConsumptionQuery struct {
	FilterBy        map[string]string `json:"filterBy"`
	AggregateBy     []string          `json:"aggregateBy,omitempty"`
	ChargeableItems []string          `json:"chargeableItems,omitempty"`
}

func (q *ConsumptionQuery) AggregatesByTimeStep() bool {
	for _, a := range q.AggregateBy {
		if a == AggregateByTimeStep {
			return true
		}
	}
	return false
}

func (q *ConsumptionQuery) DoesAggregateBy(field string) bool {
	for _, a := range q.AggregateBy {
		if a == field {
			return true
		}
	}
	return false
}

func (q *ConsumptionQuery) EnsureAggregatesBy(field string) {
	if q.DoesAggregateBy(field) {
		return
	}
	q.AggregateBy = append(q.AggregateBy, field)
}

func (q *ConsumptionQuery) EnsureDoesNotAggregateBy(field string) {
	if !q.DoesAggregateBy(field) {
		return
	}
	var newAggregateBy []string
	for _, a := range q.AggregateBy {
		if a != field {
			newAggregateBy = append(newAggregateBy, a)
		}
	}
}

func (q *ConsumptionQuery) Clone() *ConsumptionQuery {
	cloned := &ConsumptionQuery{
		FilterBy:        map[string]string{},
		AggregateBy:     make([]string, len(q.AggregateBy)),
		ChargeableItems: make([]string, len(q.ChargeableItems)),
	}
	copy(cloned.AggregateBy, q.AggregateBy)
	copy(cloned.ChargeableItems, q.ChargeableItems)
	for k, v := range q.FilterBy {
		cloned.FilterBy[k] = v
	}
	return cloned
}

type QueryResult struct {
	Periods          []*ConsumptionPeriod `json:"periods"`
	ConsumptionCount int                  `json:"consumptionCount"`
}

func (q *QueryResult) AllConsumptions() []*Consumption {
	var consumptions []*Consumption
	for _, p := range q.Periods {
		for _, g := range p.Groups {
			consumptions = append(consumptions, g.Consumptions...)
		}
	}
	return consumptions
}

func (q *QueryResult) Merge(other *QueryResult) {
	if other == nil {
		return
	}
	for _, otherPeriod := range other.Periods {
		for _, thisPeriod := range q.Periods {
			if thisPeriod.StartTime == otherPeriod.StartTime && thisPeriod.EndTime == otherPeriod.EndTime {
				q.mergePeriods(thisPeriod, otherPeriod)
			}
		}
	}
}

func (q *QueryResult) mergePeriods(thisPeriod *ConsumptionPeriod, otherPeriod *ConsumptionPeriod) {
	for _, otherGroup := range otherPeriod.Groups {
		for _, thisGroup := range thisPeriod.Groups {
			if matchGroupKeys(thisGroup.Key, otherGroup.Key) {
				thisGroup.Consumptions = append(thisGroup.Consumptions, otherGroup.Consumptions...)
			}
		}
	}
}

type ConsumptionPeriod struct {
	StartTime        time.Time           `json:"startTime"`
	EndTime          time.Time           `json:"endTime"`
	TotalSeconds     math.Float64        `json:"totalSeconds"`
	ElapsedSeconds   math.Float64        `json:"elapsedSeconds,omitempty"`
	ConsumptionCount int                 `json:"consumptionCount"`
	Groups           []*ConsumptionGroup `json:"groups"`
}

func (p *ConsumptionPeriod) TimeSegment() timeUtils.TimeSegment {
	return timeUtils.TimeSegment{Start: p.StartTime, End: p.EndTime}
}

type ConsumptionGroup struct {
	Key          map[string]any `json:"key"`
	Consumptions []*Consumption `json:"consumptions"`
}
