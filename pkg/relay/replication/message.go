package replication

type Message struct {
	Id        string         `json:"id"`
	Created   string         `json:"created"`
	Delivered string         `json:"delivered"`
	Payload   map[string]any `json:"payload" gorm:"type:jsonb"`
}
