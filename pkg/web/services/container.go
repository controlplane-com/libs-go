package services

import (
	"errors"

	"github.com/controlplane-com/libs-go/pkg/web"
)

var (
	ServiceNotFoundErr error = errors.New("no matching service found")
)

type ServiceContainer interface {
	InjectService(service Service)
	ListServices() []Service
	GetServiceByName(name string) Service
	RemoveService(service Service)
	Start() <-chan error
	Stop()
}

type serviceContainer struct {
	web.Container[Service]
	errChan chan error
}

func NewServiceContainer() ServiceContainer {
	return &serviceContainer{
		Container: web.NewContainer[Service](),
	}
}

func (s *serviceContainer) InjectService(service Service) {
	s.Container.Append(service.Name(), service)
}

func (s *serviceContainer) ListServices() []Service {
	return s.Container.List()
}

func (s *serviceContainer) RemoveService(service Service) {
	s.Container.Delete(service.Name())
}

func (s *serviceContainer) GetServiceByName(name string) Service {
	return s.Container.Get(name)
}

func (s *serviceContainer) Start() <-chan error {
	s.errChan = make(chan error, len(s.List())*2)
	for _, v := range s.List() {
		go func() {
			ch := v.Start()
			for e := range ch {
				if e != nil {
					s.errChan <- e
				}
			}
		}()
	}
	return s.errChan
}
func (s *serviceContainer) Stop() {
	for _, v := range s.List() {
		v.Stop()
	}
	close(s.errChan)
}

func GetService[TService Service](c ServiceContainer) (TService, error) {
	for _, s := range c.ListServices() {
		if serviceOfDesiredType, ok := s.(TService); ok {
			return serviceOfDesiredType, nil
		}
	}

	var zeroValue TService
	return zeroValue, ServiceNotFoundErr
}

func MustGetService[TService Service](c ServiceContainer) TService {
	s, err := GetService[TService](c)
	if err != nil {
		panic(err)
	}
	return s
}
