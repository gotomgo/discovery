package discovery

import "reflect"

// AOResolver is the signature for a function that resolves a wrapper for another item
type AOResolver func(discovery Discovery, item interface{}) (interface{}, error)

// AOResolverMapping binds an item wrapper (AO) with a function that can instance it
type AOResolverMapping struct {
	Type    reflect.Type
	Creator AOResolver
}

// AOItemResolver provides the ability to add and retrieve AO mappings
type AOItemResolver interface {
	GetAOMappings(itemType reflect.Type) ([]AOResolverMapping, bool)
	AddAOMapping(mapping AOResolverMapping)
	AddAOMappings(mapping []AOResolverMapping)
	WrapAO(d Discovery, itemType reflect.Type, item interface{}) (interface{}, error)
}
