package discovery

import "reflect"

// Discovery is the primary interface for finding/acquiring items via discovery
type Discovery interface {
	HasItem(itemType reflect.Type) bool

	GetItem(itemType reflect.Type) (interface{}, error)
	GetRequiredItem(itemType reflect.Type) interface{}

	GetItemWithOptions(itemType reflect.Type, options ResolveOptions) (interface{}, error)
	GetRequiredItemWithOptions(itemType reflect.Type, options ResolveOptions) (interface{}, error)

	// WrapAO can be used to resolve an AO item wrapper when an item is NOT
	// automatically wrapped because the item is created directly and NOT via discovery
	WrapAO(itemType reflect.Type, item interface{}) (interface{}, error)
}
