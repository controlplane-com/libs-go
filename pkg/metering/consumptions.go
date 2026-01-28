package metering

import (
	"errors"

	dynamic_objects "github.com/controlplane-com/libs-go/pkg/dynamic-objects"
	"github.com/controlplane-com/libs-go/pkg/math"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	mapset "github.com/deckarep/golang-set/v2"
)

type WorkloadConsumptionTags struct {
	BaseConsumptionTags
	Workload string `json:"workload"`
}

type BaseConsumptionTags struct {
	Feature      string `json:"feature"`            //VolumeSets, Workloads, Ingress, SpiceDb, etc.
	Metric       string `json:"metric"`             //Required. CPU, Memory, Egress, StorageCapacity, etc.
	Org          string `json:"org"`                //Required. Everything lives in an org
	Gvc          string `json:"gvc,omitempty"`      //Optional. Not all resources are gvc-scoped
	Name         string `json:"name"`               //Required. All metered entities have a name.
	Location     string `json:"location,omitempty"` //Optional. It's possible that some objects will be global
	LocationType string `json:"locationType"`
}

type VolumeSetConsumptionTags struct {
	WorkloadConsumptionTags
	VolumeIndex      int    `json:"volumeIndex"`
	Driver           string `json:"driver"`
	PerformanceClass string `json:"performanceClass"`
}

type VolumeSetSnapshotConsumptionTags struct {
	VolumeSetConsumptionTags
	SnapshotName string `json:"snapshotName"`
}

type Charges struct {
	Charge    math.Float64   `json:"total"`
	Currency  string         `json:"currency,omitempty"`
	RatePlan  []string       `json:"ratePlans"`
	Details   []ChargeDetail `json:"details,omitempty" gorm:"type:jsonb"`
	detailMap map[string]ChargeDetail
}

func (c *Charges) Add(other *Charges) *Charges {
	var sum Charges
	sum.detailMap = c.detailMap
	if sum.detailMap == nil {
		sum.detailMap = map[string]ChargeDetail{}
		for _, d := range c.Details {
			sum.detailMap[d.getDetailMapKey()] = d
		}
	}
	sum.Charge = c.Charge + other.Charge
	sum.Currency = c.Currency
	uniqueRatePlans := mapset.NewThreadUnsafeSet[string](c.RatePlan...)
	uniqueRatePlans.Append(other.RatePlan...)
	sum.RatePlan = uniqueRatePlans.ToSlice()

	for _, d := range other.Details {
		key := d.getDetailMapKey()
		if existingDetail, ok := sum.detailMap[d.RatePlanItem]; !ok {
			sum.detailMap[key] = d
		} else {
			sum.detailMap[key] = existingDetail.Add(d)
		}
	}
	return &sum
}

func (c *Charges) ExtractDetailsFromMap() {
	if c.detailMap != nil {
		c.Details = pipeline.ExtractMapValues(c.detailMap)
	}
}

type ChargeDetail struct {
	Charge           math.Float64 `json:"amount"`
	Units            math.Float64 `json:"units"`
	Rate             math.Float64 `json:"rate"`
	RatePlanItem     string       `json:"ratePlanItem"`
	ChargeableItemId string       `json:"chargeableItemId"`
	From             math.Float64 `json:"from"`
	To               math.Float64 `json:"to"`
}

func (c ChargeDetail) Add(other ChargeDetail) ChargeDetail {
	return ChargeDetail{
		Charge:       c.Charge + other.Charge,
		Units:        c.Units + other.Units,
		Rate:         c.Rate,
		RatePlanItem: c.RatePlanItem,
		From:         c.From,
		To:           c.To,
	}
}

func (c ChargeDetail) getDetailMapKey() string {
	if c.RatePlanItem != "" {
		return c.RatePlanItem
	}
	return c.ChargeableItemId
}

type Consumption struct {
	*Charges        `gorm:"-"`
	ProjectedCharge math.Float64 `json:"projectedTotal,omitempty" gorm:"-"`
	Value           math.Float64 `json:"value"`
	ProjectedValue  math.Float64 `json:"projectedValue,omitempty" gorm:"-"`
	Job             string       `json:"job,omitempty"`
	Tags            any          `json:"tags" gorm:"type:jsonb"`
	queryIndex      int
}

func (c *Consumption) Add(other *Consumption) (*Consumption, error) {
	sum := &Consumption{
		Tags:       c.Tags,
		Job:        c.Job,
		queryIndex: c.queryIndex,
	}
	if c.Charges == nil {
		c.Charges = &Charges{}
	}
	if other.Charges == nil {
		other.Charges = &Charges{}
	}
	if c.Charges.Currency != other.Charges.Currency {
		return nil, errors.New("consumptions with mismatching currencies cannot be added")
	}
	sum.Charges = c.Charges.Add(other.Charges)
	sum.Value = c.Value + other.Value
	sum.ProjectedValue = c.ProjectedValue + other.ProjectedValue
	return sum, nil
}

func (c *Consumption) FingerprintTags() string {
	return dynamic_objects.ObjectToString(c.Tags)
}

func (c *Consumption) SetQueryIndex(i int) {
	c.queryIndex = i
}

func (c *Consumption) GetQueryIndex() int {
	return c.queryIndex
}
