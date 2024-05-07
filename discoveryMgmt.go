package discovery

import "reflect"

// ItemDiscoveryManagement provides the ability to add and remove discovery items
type ItemDiscoveryManagement interface {
	GetResolver() ItemResolver

	AddItem(itemType reflect.Type, item interface{}) error
	RemoveItem(itemType reflect.Type)
}

// ItemDiscoveryManagementType is the reflected type of ItemDiscoveryManagement
var ItemDiscoveryManagementType = reflect.TypeOf((*ItemDiscoveryManagement)(nil)).Elem()
