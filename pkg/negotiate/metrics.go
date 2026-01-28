package negotiate

import (
	"strings"
	"time"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
)

// NewSyncJobWithMetrics instruments an existing SyncJob with Prometheus histograms and
// returns the same job instance. It wraps the job's delegates to record durations.
//
// Metrics exported (metric names use the job Name with hyphens/spaces replaced by underscores):
// - <name>_sync_run_seconds{job, status}: duration per full sync run (from InputDelegate to CompletedDelegate)
// - <name>_sync_item_seconds{job, status}: duration per processed item T (around OutputDelegate)
func NewSyncJobWithMetrics[T any](job *SyncJob[T], r prometheus.Registerer) *SyncJob[T] {
	// Sanitize metric name components
	metricPrefix := strings.ReplaceAll(job.Name(), "-", "_")
	metricPrefix = strings.ReplaceAll(metricPrefix, " ", "_")

	// Histograms for run-level and item-level durations
	runHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    metricPrefix + "_sync_run_seconds",
		Help:    "Duration of sync job runs in seconds",
		Buckets: []float64{0.0001, 0.001, 0.01, 0.1, 0.25, 0.5, 0.75, 1, 2, 4, 8, 10, 20, 30, 60, 120, 300},
	}, []string{"job", "status"})

	itemHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    metricPrefix + "_sync_item_seconds",
		Help:    "Duration of individual sync items in seconds",
		Buckets: []float64{0.0001, 0.001, 0.01, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.75, 1, 1.5, 2, 4, 8, 10},
	}, []string{"job", "status"})

	// Best-effort registration (ignore AlreadyRegistered errors)
	if r == nil {
		r = prometheus.DefaultRegisterer
	}
	_ = r.Register(runHistogram)
	_ = r.Register(itemHistogram)

	// Capture start time per run when InputDelegate is called.
	var runStart time.Time

	// Wrap InputDelegate to mark run start and observe error-only runs
	originalInput := job.InputDelegate
	job.InputDelegate = func() ([]T, error) {
		runStart = time.Now()
		input, err := originalInput()
		// If InputDelegate errors, negotiate loop skips CompletedDelegate; record run duration here.
		if err != nil && (*[2]uintptr)(unsafe.Pointer(&err))[1] != 0 {
			runHistogram.With(prometheus.Labels{"job": job.Name(), "status": "error"}).Observe(time.Since(runStart).Seconds())
		}
		return input, err
	}

	// Wrap OutputDelegate to measure per-item durations
	originalOutput := job.OutputDelegate
	job.OutputDelegate = func(i T) error {
		start := time.Now()
		err := originalOutput(i)
		status := "ok"
		if err != nil && (*[2]uintptr)(unsafe.Pointer(&err))[1] != 0 {
			status = "error"
		}
		itemHistogram.With(prometheus.Labels{"job": job.Name(), "status": status}).Observe(time.Since(start).Seconds())
		return err
	}

	// Wrap CompletedDelegate to measure full run duration and preserve original behavior
	originalCompleted := job.CompletedDelegate
	job.CompletedDelegate = func(items []T, err error) error {
		// Invoke the original CompletedDelegate if present and combine errors for status
		finalErr := err
		var originalReturn error
		if originalCompleted != nil {
			if e := originalCompleted(items, err); e != nil {
				originalReturn = e
				if finalErr == nil {
					finalErr = e
				}
			}
		}
		status := "ok"
		if finalErr != nil && (*[2]uintptr)(unsafe.Pointer(&finalErr))[1] != 0 {
			status = "error"
		}
		if !runStart.IsZero() {
			runHistogram.With(prometheus.Labels{"job": job.Name(), "status": status}).Observe(time.Since(runStart).Seconds())
		}
		// Preserve original CompletedDelegate return semantics
		if originalCompleted != nil {
			return originalReturn
		}
		return err
	}

	return job
}

func (s *SyncJob[T]) WithMetrics(r prometheus.Registerer) *SyncJob[T] {
	return NewSyncJobWithMetrics(s, r)
}
