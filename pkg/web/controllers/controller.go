package controllers

import (
	"github.com/gorilla/mux"
	"github.com/controlplane-com/libs-go/pkg/web/services"
)

type RouterHost interface {
	Router() *mux.Router
}

type Controller interface {
	services.ServiceContainer
	services.Service
	RouterHost
}

type defaultController struct {
	services.ServiceContainer
	RouterHost
	name string
}

func NewController(name string, serviceContainer services.ServiceContainer, router RouterHost) Controller {
	return &defaultController{ServiceContainer: serviceContainer, RouterHost: router}
}

func (d *defaultController) Name() string {
	return d.name
}
