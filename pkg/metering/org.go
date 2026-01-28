package metering

import (
	"encoding/json"
	"time"
)

type Org struct {
	Name     string    `json:"name" gorm:"primaryKey"`
	Created  time.Time `json:"created" gorm:"type:timestamptz;default:now()"`
	BucketId int64     `json:"bucketId"`
	Bucket   *Bucket   `json:"bucket,omitempty"`
}

func (o *Org) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
