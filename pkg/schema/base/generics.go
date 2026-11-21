package base

type GenericList[T any] struct {
	Kind     ListKind     `json:"kind,omitzero"`
	ItemKind ListItemKind `json:"itemKind,omitzero"`
	Items    []T          `json:"items,omitzero"`
	Links    []Link       `json:"links,omitzero"`
}
