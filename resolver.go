package discovery

import "reflect"

// Resolver is the signature for a function that resolves an item
type Resolver func(discovery Discovery) (interface{}, error)

// ResolverMapping binds an item type with a function tha can instance it
type ResolverMapping struct {
	Type    reflect.Type
	Creator Resolver
}

// ItemResolver is used during discovery to attempt to resolve an item that
//
//	has not been previously resolved, or when the InstanceItem option is specified
type ItemResolver interface {
	ResolveItem(d Discovery, itemType reflect.Type) (interface{}, error)
	ResolveMapping(d Discovery, mapping ResolverMapping) (interface{}, error)
	AddMapping(mapping ResolverMapping)
	AddMappingsVar(mappings ...ResolverMapping)
	AddMappings(mapping []ResolverMapping)
}

// ItemResolverType is the reflected type of ItemResolver
var ItemResolverType = reflect.TypeOf((*ItemResolver)(nil)).Elem()
