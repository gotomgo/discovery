package discovery

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// to support concurrency default discovery is an atomic.Value
var defaultD atomic.Value
var createLock sync.Mutex

// GetDefaultDiscovery returns the value for the default discovery
func GetDefaultDiscovery() Discovery {
	d := defaultD.Load()
	if d == nil {
		return nil
	}
	return d.(Discovery)
}

// GetDefaultDiscoveryOrPanic returns the value for the default discovery
// and panics if the default discovery has not been set
func GetDefaultDiscoveryOrPanic() Discovery {
	d := defaultD.Load()
	if d == nil {
		panic(fmt.Errorf("default Discovery Service not initialized"))
	}
	return d.(Discovery)
}

// GetOrCreateDefaultDiscovery gets the current default discovery,
// creating it if it has not been set
//
//	Params
//	  resolver - optional ItemResolver
//
//	Notes
//	  This code avoids race conditions that could occur if two
//	  entities attempt to create the default discover concurrently
func GetOrCreateDefaultDiscovery(resolver ItemResolver) Discovery {
	createLock.Lock()
	defer createLock.Unlock()

	d := GetDefaultDiscovery()
	if d == nil {
		d = SetDefaultDiscovery(NewDiscovery(resolver))
	}

	return d
}

// SetDefaultDiscovery stores the default discovery
func SetDefaultDiscovery(d Discovery) Discovery {
	defaultD.Store(d)
	return d
}

// CreateSuperDiscovery creates a new discovery that "super classes" the
// default discovery
//
//	Params
//	  resolver - optional ItemResolver
func CreateSuperDiscovery(resolver ItemResolver) Discovery {
	return NewDiscoveryWithBase(GetDefaultDiscoveryOrPanic(), resolver)
}

// GetDefaultResolver returns the resolver for default discovery
func GetDefaultResolver() ItemResolver {
	return GetDefaultDiscoveryOrPanic().(ItemDiscoveryManagement).GetResolver()
}

// AddItem is a helper method for adding an item to default discovery
//
//	Notes
//		If default discovery has not been assigned, it is created With
//		a default resolver.
func AddItem(itemType reflect.Type, item interface{}) error {
	return GetOrCreateDefaultDiscovery(nil).(ItemDiscoveryManagement).AddItem(itemType, item)
}

func GetRequiredItem[T any](d Discovery, itemType reflect.Type) T {
	if d == nil {
		d = GetDefaultDiscoveryOrPanic()
	}

	return d.GetRequiredItem(itemType).(T)
}

func GetItem[T any](d Discovery, itemType reflect.Type) (T, error) {
	if d == nil {
		d = GetDefaultDiscoveryOrPanic()
	}

	item, err := d.GetItem(itemType)
	if err != nil {
		var zero T
		return zero, err
	}

	return item.(T), nil
}

func Resolve[T any](d Discovery, mapping ResolverMapping) T {
	item, err := GetRequiredItem[ItemResolver](d, ItemResolverType).ResolveMapping(d, mapping)
	if err != nil {
		panic(fmt.Errorf("error resolving '%t': %s", mapping.Type, err))
	}

	return item.(T)
}
