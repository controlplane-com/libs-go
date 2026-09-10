package negotiate

import (
	"sync/atomic"
	"testing"
	"time"

	leaderElection "github.com/controlplane-com/libs-go/pkg/leader-election"
)

type fakeElector struct {
	role atomic.Value
}

func (e *fakeElector) Role() leaderElection.Type {
	return e.role.Load().(leaderElection.Type)
}

// A replica that starts before it is Ready is not visible to the elector, so its immediate first
// run is skipped; it must retry soon after, not a whole SyncInterval later.
func TestFirstRunImmediatelyRunsOnceLeadershipArrives(t *testing.T) {
	firstRunRetryInterval = 20 * time.Millisecond
	elector := &fakeElector{}
	elector.role.Store(leaderElection.TypeFollower)

	runs := make(chan time.Time, 10)
	job := NewSyncJob(SyncJobOptions[struct{}]{
		Name:                "first-run",
		NumWorkers:          1,
		BufferSize:          1,
		SyncInterval:        time.Hour,
		FirstRunImmediately: true,
		Elector:             elector,
		InputDelegate: func() ([]struct{}, error) {
			runs <- time.Now()
			return nil, nil
		},
		OutputDelegate: func(struct{}) error { return nil },
	})
	job.Start()
	defer job.Stop()

	select {
	case <-runs:
		t.Fatal("ran while still a follower")
	case <-time.After(100 * time.Millisecond):
	}

	elector.role.Store(leaderElection.TypeLeader)
	select {
	case <-runs:
	case <-time.After(2 * time.Second):
		t.Fatal("first run did not happen after becoming leader")
	}

	select {
	case <-runs:
		t.Fatal("kept retrying after the first leader run instead of waiting SyncInterval")
	case <-time.After(200 * time.Millisecond):
	}
}
