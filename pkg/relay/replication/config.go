package replication

import (
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	"go.uber.org/zap/zapcore"
	"time"
)

type WalData struct {
	Changes []*Change `json:"change"`
}

type Change struct {
	Action       Action        `json:"kind"`
	Schema       string        `json:"schema"`
	Table        string        `json:"table"`
	ColumnNames  []string      `json:"columnnames"`
	ColumnTypes  []string      `json:"columntypes"`
	ColumnValues []interface{} `json:"columnvalues"`
	OldKeys      *OldKeys      `json:"oldkeys,omitempty"`
}

type OldKeys struct {
	KeyNames  []string      `json:"keynames"`
	KeyTypes  []string      `json:"keytypes"`
	KeyValues []interface{} `json:"keyvalues"`
}

type Config struct {
	LogLevel                    zapcore.Level   `cpln:"default:info;mapper:ZapLogLevelMapper"`
	Host                        string          `json:"-"`
	Port                        string          `json:"-"`
	User                        string          `json:"-"`
	Database                    string          `json:"-"`
	Password                    string          `json:"-"`
	Slot                        string          `json:"slot"`
	WalAcknowledgementFrequency time.Duration   `cpln:"default:30s" json:"-"`
	Destinations                DestinationList `json:"destinations"`
}

func (c Config) Tables() []string {
	tableMap := make(map[string]string)
	for _, d := range c.Destinations {
		tableMap[d.Table] = d.Table
	}
	return pipeline.ExtractMapValues(tableMap)
}

type Action = string

const ActionInsert Action = "insert"
const ActionUpdate Action = "update"
const ActionDelete Action = "delete"

/*
func init() {
	if err := config.Initialize(&Config, logging.ZapLogLevelMapper{}, &config.SliceMapper{}); err != nil {
		panic(err)
	}
	c := config.GetConfig[Config]()
	if err := logging.InitializeLogger(c.LogLevel); err != nil {
		panic(err)
	}
	fmt.Println(config.Summarize(c))
}
*/
