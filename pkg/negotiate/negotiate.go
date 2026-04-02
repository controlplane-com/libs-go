package negotiate

import (
	"context"
	"time"
	"unsafe"

	"github.com/robfig/cron/v3"
	"github.com/controlplane-com/libs-go/pkg/common"
	leaderElection "github.com/controlplane-com/libs-go/pkg/leader-election"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/controlplane-com/libs-go/pkg/threading"
	"go.uber.org/zap"
)

type OutputDelegate[T any] func(i T) error
type InputDelegate[T any] func() ([]T, error)
type CompletedDelegate[T any] func([]T, error) error

type SyncJob[T any] struct {
	SyncJobOptions[T]
	running    bool
	ctx        context.Context
	cancelFunc context.CancelFunc
	errChan    chan error
	runCount   int
}

type SyncJobOptions[T any] struct {
	NumWorkers          int
	InputDelegate       InputDelegate[T]
	OutputDelegate      threading.Handler[T]
	CompletedDelegate   CompletedDelegate[T]
	ErrorCallback       threading.ErrorCallback[T]
	SyncInterval        time.Duration
	CronSchedule        string
	NextRunDelegate     func() time.Time
	BufferSize          int
	Name                string
	Elector             leaderElection.Elector
	RetryPolicy         threading.RetryPolicy
	MaxRetries          int
	RetryAfterDuration  time.Duration
	OnLeadershipLost    func() // Called when transitioning from leader to non-leader
	FirstRunImmediately bool   // If true, run immediately on startup, otherwise wait for SyncInterval or CronSchedule
}

func NewSyncJob[T any](options SyncJobOptions[T]) *SyncJob[T] {
	s := &SyncJob[T]{
		SyncJobOptions: options,
		errChan:        make(chan error),
	}
	return s
}

func (s *SyncJob[T]) Start() <-chan error {
	s.running = true
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())

	// If CronSchedule is provided and NextRunDelegate isn't yet set,
	// parse the cron spec and compute next run times via the library.
	if s.CronSchedule != "" && s.NextRunDelegate == nil {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		sched, err := parser.Parse(s.CronSchedule)
		if err != nil {
			logging.LoggerWithContext(s.ctx).With("syncJob", s.Name()).
				Errorf("invalid CronSchedule %q: %v; falling back to SyncInterval", s.CronSchedule, err)
		} else {
			s.NextRunDelegate = func() time.Time {
				return sched.Next(time.Now())
			}
			logging.LoggerWithContext(s.ctx).With("syncJob", s.Name()).
				Infow("cron schedule parsed", "spec", s.CronSchedule)
		}
	}

	go s.negotiate()
	return s.errChan
}

func (s *SyncJob[T]) Stop() {
	s.cancelFunc()
	s.running = false
	close(s.errChan)
}

func (s *SyncJob[T]) Name() string {
	return s.SyncJobOptions.Name
}

func (s *SyncJob[T]) logger() *zap.SugaredLogger {
	return logging.LoggerWithContext(s.ctx).With("syncJob", s.Name())
}

func (s *SyncJob[T]) negotiate() {
	startTime := time.Now()
	var role leaderElection.Type
	var prevRole leaderElection.Type
	if s.Elector != nil {
		role = s.Elector.Role()
	} else {
		role = leaderElection.TypeLeader
	}
	prevRole = role

	for {
		logger := s.logger()
		cycleId := common.NewUuid()
		logger = logger.With("cycleId", cycleId)
		if !s.awaitNextRun(s.ctx, logger, startTime, role) {
			break
		}

		startTime = time.Now()
		s.runCount++
		if s.Elector != nil {
			role = s.Elector.Role()

			// Detect leadership loss transition and invoke callback
			if prevRole == leaderElection.TypeLeader && role != leaderElection.TypeLeader {
				if s.OnLeadershipLost != nil {
					s.OnLeadershipLost()
				}
			}
			prevRole = role

			if role != leaderElection.TypeLeader {
				continue
			}
		}

		input, err := s.InputDelegate()
		if err != nil && (*[2]uintptr)(unsafe.Pointer(&err))[1] != 0 {
			logger.Errorf("error gathering input: %v", err)
			continue
		}
		logger.Infof("executing %s", s.SyncJobOptions.Name)
		queue := threading.NewWorkQueue[T](s.OutputDelegate, s.BufferSize, s.NumWorkers)

		// Configure retry policy if specified
		if s.RetryPolicy != 0 {
			queue.RetryPolicy = s.RetryPolicy
		}
		if s.MaxRetries > 0 {
			queue.MaxRetries = s.MaxRetries
		}
		if s.RetryAfterDuration > 0 {
			queue.RetryAfterDuration = s.RetryAfterDuration
		}

		// Configure error callback if specified
		if s.ErrorCallback != nil {
			queue.ErrorCallback = s.ErrorCallback
		}

		queue.Start()
		queue.Enqueue(input...)
		queue.Stop()
		queue.AwaitCompletion()
		if s.CompletedDelegate != nil {
			err = s.CompletedDelegate(input, queue.Error())
			if err != nil {
				logger.Errorf("Error running completed delegate: %v", err)
			}
		}
		logger.Infof("done in %f seconds", time.Now().Sub(startTime).Seconds())
	}
}

func (s *SyncJob[T]) awaitNextRun(ctx context.Context, logger *zap.SugaredLogger, lastStartTime time.Time, role leaderElection.Type) bool {
	if s.FirstRunImmediately && s.runCount == 0 {
		return true
	}
	var timeToWait time.Duration
	if s.NextRunDelegate != nil {
		timeToWait = s.NextRunDelegate().Sub(time.Now())
	}
	if timeToWait <= 0 {
		timeToWait = s.SyncInterval - time.Now().Sub(lastStartTime)
	}
	keepRunning := true
	if timeToWait > 0 {
		//No need to clutter the log with these in non-leader replicas
		if role == leaderElection.TypeLeader {
			logger.Infof("Waiting %f seconds to run", timeToWait.Seconds())
		}
		select {
		case <-time.After(timeToWait):
			keepRunning = true
		case <-ctx.Done():
			keepRunning = false
		}
	}
	return keepRunning
}
