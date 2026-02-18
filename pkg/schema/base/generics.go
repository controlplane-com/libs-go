package base

type GenericList[T any] struct {
	Kind     ListKind     `json:"kind,omitempty"`
	ItemKind ListItemKind `json:"itemKind,omitempty"`
	Items    []T          `json:"items,omitempty"`
	Links    []Link       `json:"links,omitempty"`
}
