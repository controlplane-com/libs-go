package bucket

import "time"

// MoveOrgRequest contains parameters for moving an organization to a different bucket
type MoveOrgRequest struct {
	SourceBucket      int       `json:"sourceBucket"`
	DestinationBucket int       `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}

// CopyOrgRequest contains parameters for copying organization data between buckets
type CopyOrgRequest struct {
	SourceBucket      int       `json:"sourceBucket"`
	DestinationBucket int       `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}

// ScrubOrgRequest contains parameters for deleting organization data from a bucket
type ScrubOrgRequest struct {
	DestinationBucket int       `json:"destinationBucket"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}
