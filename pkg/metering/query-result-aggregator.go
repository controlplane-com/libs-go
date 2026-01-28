package metering

import (
	"encoding/json"
	"time"

	"github.com/controlplane-com/libs-go/pkg/math"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"golang.org/x/exp/slices"
)

type queryResultMap map[time.Time]map[string]map[string]*Consumption
type QueryResultAggregator struct {
	m        queryResultMap
	TimeStep timeUtils.TimeStep
	timeUtils.TimeSegment
	AggregateByTimeStep bool
}

func NewQueryResultAggregator(aggregationPeriod timeUtils.TimeSegment, timeStep timeUtils.TimeStep, aggregateByTimeStep bool) QueryResultAggregator {
	return QueryResultAggregator{m: queryResultMap{}, TimeStep: timeStep, TimeSegment: aggregationPeriod, AggregateByTimeStep: aggregateByTimeStep}
}

func (q QueryResultAggregator) AggregateQueryResults(results ...*QueryResult) error {
	for _, r := range results {
		for _, p := range r.Periods {
			var timeStepStart time.Time
			if q.AggregateByTimeStep {
				timeStepStart = timeUtils.AlignTimeWithStepStart(p.StartTime, q.TimeStep)
			} else {
				timeStepStart = q.Start
			}
			periodEntry, ok := q.m[timeStepStart]
			if !ok {
				periodEntry = map[string]map[string]*Consumption{}
				q.m[timeStepStart] = periodEntry
			}
			for _, g := range p.Groups {
				b, err := json.Marshal(g.Key)
				if err != nil {
					return err
				}
				groupKey := string(b)
				groupEntry, ok := periodEntry[groupKey]
				if !ok {
					groupEntry = map[string]*Consumption{}
					periodEntry[groupKey] = groupEntry
				}
				for _, c := range g.Consumptions {
					b, err := json.Marshal(c.Tags)
					if err != nil {
						return err
					}
					consumptionKey := string(b)
					if c.Charges != nil {
						consumptionKey += c.Charges.Currency
					}
					consumptionEntry, ok := groupEntry[consumptionKey]
					if !ok {
						groupEntry[consumptionKey] = c
						continue
					}
					groupEntry[consumptionKey], err = consumptionEntry.Add(c)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (q QueryResultAggregator) QueryResult(timeStep timeUtils.TimeStep) (*QueryResult, error) {
	queryResult := &QueryResult{}
	for startTime, groupMap := range q.m {
		var periodStart, periodEnd time.Time
		if q.AggregateByTimeStep {
			periodStart, periodEnd = startTime, timeStep.Advance(startTime, 1)
		} else {
			periodStart, periodEnd = startTime, q.End
		}
		finalPeriod := &ConsumptionPeriod{
			StartTime:    periodStart,
			EndTime:      periodEnd,
			TotalSeconds: math.Float64(periodEnd.Sub(periodStart).Seconds()),
		}
		now := time.Now()
		if finalPeriod.EndTime.After(now) {
			finalPeriod.ElapsedSeconds = math.Float64(now.Sub(finalPeriod.StartTime).Seconds())
		}
		for periodKey, consumptionMap := range groupMap {
			var key map[string]any
			err := json.Unmarshal([]byte(periodKey), &key)
			if err != nil {
				return nil, err
			}
			finalGroup := &ConsumptionGroup{
				Key: key,
			}
			for _, consumption := range consumptionMap {
				finalPeriod.ConsumptionCount++
				queryResult.ConsumptionCount++
				if consumption.Charges != nil {
					consumption.Charges.ExtractDetailsFromMap()
				}
				if finalPeriod.TimeSegment().Contains(now, true) {
					q.projectConsumption(finalPeriod, consumption)
				}
				finalGroup.Consumptions = append(finalGroup.Consumptions, consumption)
			}
			finalPeriod.Groups = append(finalPeriod.Groups, finalGroup)
		}
		queryResult.Periods = append(queryResult.Periods, finalPeriod)
	}
	slices.SortFunc(queryResult.Periods, func(a *ConsumptionPeriod, b *ConsumptionPeriod) int {
		return int(a.StartTime.Sub(b.StartTime).Milliseconds())
	})

	return queryResult, nil
}

func (q QueryResultAggregator) projectConsumption(period *ConsumptionPeriod, consumption *Consumption) {
	ratio := period.TotalSeconds / period.ElapsedSeconds
	consumption.ProjectedValue = consumption.Value * ratio
	if consumption.Charges != nil {
		consumption.ProjectedCharge = consumption.Charge * ratio
	}
}
