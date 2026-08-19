package data_service

import (
	"fmt"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
)

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

	//We need a new collection. An empty page is not the end of the list as long
	//as the server hands out a next link — stopping there would silently
	//truncate the listing.
	for i.nextUrl != "" {
		fetched := i.nextUrl
		_, err := i.client.Get(i.nextUrl, &list)
		if err != nil {
			i.err = err
			return false
		}
		i.items = list.Items
		i.nextUrl = GetLinkByRel("next", list.Links)
		if i.nextUrl == fetched {
			i.err = fmt.Errorf("pagination is not advancing at %s", fetched)
			return false
		}
		if len(i.items) > 0 {
			i.currentItemIndex = 0
			return true
		}
	}
	return false
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
