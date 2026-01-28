package web

import "github.com/controlplane-com/libs-go/pkg/database"

type RepositoryContainer interface {
	Container[database.Repository]
}

func NewRepositoryContainer() RepositoryContainer {
	return NewContainer[database.Repository]()
}

func GetRepository[TRepo database.Repository](c RepositoryContainer) (TRepo, error) {
	for _, s := range c.List() {
		if serviceOfDesiredType, ok := s.(TRepo); ok {
			return serviceOfDesiredType, nil
		}
	}

	var zeroValue TRepo
	return zeroValue, ContainerItemNotFoundErr
}

func MustGetRepository[TRepo database.Repository](c RepositoryContainer) TRepo {
	r, err := GetRepository[TRepo](c)
	if err != nil {
		panic(err)
	}
	return r
}
