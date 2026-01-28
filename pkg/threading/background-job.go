package threading

import (
	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/controlplane-com/libs-go/pkg/process"
)

type BackgroundJob interface {
	Start() <-chan error
	Stop()
	Name() string
}

func StartJobs(jobs []BackgroundJob) {
	logger := logging.Logger().Sugar()

	for _, j := range jobs {
		logging.Logger().Sugar().Infof("Starting %s", j.Name())
		c := j.Start()
		job := j
		GoSafely(func() error {
			err := <-c
			if err != nil {
				logger.Errorf("Error while running %s. %v", job.Name(), err)
			}
			logging.Logger().Sugar().Infof("%s has stopped running. Initiating graceful shutdown.", job.Name())
			process.Term()
			//No need to further propagate the error
			return nil
		})
	}
}
