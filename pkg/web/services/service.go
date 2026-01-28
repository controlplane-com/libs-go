package services

import "github.com/controlplane-com/libs-go/pkg/threading"

type Service interface {
	threading.BackgroundJob
}
