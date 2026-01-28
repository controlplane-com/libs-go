package bucket

import (
	"encoding/json"
	"fmt"
	"time"
)

// Org represents an organization that is assigned to a bucket for data partitioning
type Org struct {
	Name     string    `json:"name" gorm:"primaryKey"`
	Created  time.Time `json:"created" gorm:"type:timestamptz;default:now()"`
	BucketId int       `json:"bucketId"`
	Bucket   *Bucket   `json:"bucket,omitempty"`
}

func (o *Org) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

// Bucket represents a partition bucket that contains multiple organizations
type Bucket struct {
	Id               int       `json:"id"`
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
