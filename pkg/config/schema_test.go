package config

import (
	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
	"os"
	"testing"
	"time"
)

type exampleSchema struct {
	Foo string        `cpln:"default:Hello, World!"`
	Bar time.Duration `cpln:"default:24h"`
	Baz time.Time     `cpln:"env:SOME_OTHER_KEY;default:2024-04-01T00:00:00Z"`
}

func TestSchema(t *testing.T) {
	s := exampleSchema{}
	err := ParseSchema(&s)
	if err != nil {
		t.FailNow()
	}
	if s.Foo != "Hello, World!" {
		t.FailNow()
	}
	if s.Bar != time.Hour*24 {
		t.FailNow()
	}
	if s.Baz != timeUtils.MustParseTime(time.RFC3339, "2024-04-01T00:00:00Z") {
		t.FailNow()
	}

	err = os.Setenv("SOME_OTHER_KEY", "2024-01-01T00:00:00Z")
	if err != nil {
		t.FailNow()
	}
	err = os.Setenv("BAR", "-1m")
	if err != nil {
		t.FailNow()
	}
	err = ParseSchema(&s)
	if err != nil {
		t.FailNow()
	}
	if s.Foo != "Hello, World!" {
		t.FailNow()
	}
	if s.Bar != -time.Minute {
		t.FailNow()
	}
	if s.Baz != timeUtils.MustParseTime(time.RFC3339, "2024-01-01T00:00:00Z") {
		t.FailNow()
	}
}
