package metering

import (
	"errors"
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"time"
)

type TagValuesQueryRequest struct {
	StartTime time.Time         `json:"startTime"`
	EndTime   time.Time         `json:"endTime"`
	Tag       string            `json:"tag"`
	FilterBy  map[string]string `json:"filterBy"`
}

func (r *TagValuesQueryRequest) TimeSegment() timeUtils.TimeSegment {
	return timeUtils.TimeSegment{Start: r.StartTime, End: r.EndTime}
}

func (r *TagValuesQueryRequest) Validate() error {
	if !r.StartTime.Before(r.EndTime) {
		return errors.New("startTime must be before endTime")
	}
	if r.Tag == "" {
		return errors.New("tag cannot be empty")
	}
	if !r.TimeSegment().IsAlignedWithTimeStep(timeUtils.TimeStepHour) {
		return errors.New("startTime and endTime must be aligned with an hour boundary")
	}
	return nil
}

type TagValuesQueryResult struct {
	*TagValuesQueryRequest
	TagValues []string `json:"tagValues"`
}

type MoveOrgRequest struct {
	SourceBucket      int64     `json:"sourceBucket"`
	DestinationBucket int64     `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}

type CopyOrgRequest struct {
	SourceBucket      int64     `json:"sourceBucket"`
	DestinationBucket int64     `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}

type ScrubOrgRequest struct {
	DestinationBucket int64     `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}
