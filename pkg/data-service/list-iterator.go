package data_service

import "github.com/controlplane-com/types-go/pkg/base"

type ListIterator[T any] struct {
	client           *DataServiceClient
	nextUrl          string
	err              error
	items            []T
	currentItemIndex int
}

func NewListIterator[T any](client *DataServiceClient, url string) *ListIterator[T] {
	return &ListIterator[T]{
		client:  client,
		nextUrl: url,
	}
}

func (i *ListIterator[T]) Next() bool {
	var list base.GenericList[T]
	//If we haven't iterated through the most recent collection...
	if i.currentItemIndex < len(i.items)-1 {
		i.currentItemIndex++
		return true
	}

	//We need a new collection...
	if i.nextUrl == "" {
		return false
	}
	_, err := i.client.Get(i.nextUrl, &list)
	if err != nil {
		i.err = err
		return false
	}
	i.items = list.Items
	i.nextUrl = GetLinkByRel("next", list.Links)
	if len(i.items) == 0 {
		return false
	}
	i.currentItemIndex = 0
	return true
}

func (i *ListIterator[T]) Item() T {
	return i.items[i.currentItemIndex]
}

func (i *ListIterator[T]) Error() error {
	return i.err
}

func (i *ListIterator[T]) VisitAll(handler func(t T) error) ([]T, error) {
	var items []T
	for i.Next() {
		t := i.Item()
		if handler != nil {
			if err := handler(t); err != nil {
				return nil, err
			}
		}
		items = append(items, t)
	}
	if err := i.Error(); err != nil {
		return nil, err
	}
	return items, nil
}

func (i *ListIterator[T]) List() ([]T, error) {
	return i.VisitAll(nil)
}
