package metering

import (
	"fmt"
	"time"
)

type GetBucketRequest struct {
	Id int64 `json:"id"`
}

type Bucket struct {
	Id               int64     `json:"id"`
	PartitionsEnding time.Time `json:"partitionsEnding" gorm:"type:timestamptz"`
	Created          time.Time `json:"created" gorm:"type:timestamptz;default:now()"`
	Orgs             []*Org    `json:"orgs,omitempty"`
}

func (b *Bucket) SchemaName() string {
	if b == nil || b.Id == 0 {
		return "public"
	}
	return fmt.Sprintf("org_bucket_%d", b.Id)
}

func (b *Bucket) CheckpointName() string {
	return b.SchemaName()
}
