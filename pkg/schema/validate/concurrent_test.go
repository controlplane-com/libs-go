package validate

import (
	"sync"
	"testing"
)

// A schema's rule graph is built LAZILY on first use and then shared by every
// concurrent request that validates that type. This pins that the build publishes
// safely.
//
// Before the fix, ApplyCtx/Rule() guarded the build with `if s.Root == nil` around
// the sync.Once — a double-checked lock with no atomics. A second goroutine could see
// a non-nil s.Root that the builder had published without the happens-before Once.Do
// provides, skip the Do, and read a rule graph whose interior writes were not visible
// to it. `go test -race` reports it here; the data-service saw ~200 race reports a
// minute under concurrent `-query` traffic, and a sibling instance of the same class
// (an unsynchronized package-level map on the query path) killed the process outright
// with "fatal error: concurrent map writes".
//
// The barrier is what makes it bite: every goroutine must be waiting BEFORE the first
// build starts, or they all arrive after it and observe a finished graph.
type concCase struct {
	Name   string   `json:"name" validate:"required,min=1,max=64"`
	Count  int      `json:"count" validate:"min=1,max=100"`
	Kind   string   `json:"kind" validate:"valid=alpha|beta|gamma"`
	Tags   []string `json:"tags"`
	Nested *concSub `json:"nested"`
}

type concSub struct {
	Inner string `json:"inner" validate:"min=2"`
	Extra int    `json:"extra" validate:"min=0,max=9"`
}

func TestSchemaBuildIsRaceFreeUnderConcurrentFirstUse(t *testing.T) {
	s := For[concCase]()
	// A schema cached by an earlier test would already be built, and the race is in the
	// FIRST use. This type is used nowhere else, so the cache entry is fresh here.
	if s.Root != nil {
		t.Fatal("schema was already built; this test must be the first user of concCase")
	}

	const n = 64
	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)
	done.Add(n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer done.Done()
			start.Wait()
			obj := map[string]any{
				"name":   "n",
				"count":  3,
				"kind":   "beta",
				"tags":   []any{"a", "b"},
				"nested": map[string]any{"inner": "xy", "extra": 4},
			}
			errs[i] = s.ApplyCtx(obj, nil, Prefs{})
		}(i)
	}
	start.Done()
	done.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected validation error: %v", i, err)
		}
	}
	if s.Rule() == nil {
		t.Fatal("the rule graph was never built")
	}
}

// The same check through Rule(), the other entry point that used to carry the
// pre-check — a caller deriving a variant of a schema races the builder identically.
func TestSchemaRuleAccessorIsRaceFreeUnderConcurrentFirstUse(t *testing.T) {
	type ruleCase struct {
		Name string `json:"name" validate:"required,min=1"`
	}
	s := For[ruleCase]()

	const n = 64
	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)
	done.Add(n)
	got := make([]*Rule, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer done.Done()
			start.Wait()
			got[i] = s.Rule()
		}(i)
	}
	start.Done()
	done.Wait()

	for i, r := range got {
		if r == nil {
			t.Fatalf("goroutine %d got a nil rule", i)
		}
		if r != got[0] {
			t.Fatalf("goroutine %d got a different rule graph — the build ran more than once", i)
		}
	}
}
