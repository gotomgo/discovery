package discovery

import (
	"container/list"
	"reflect"
	"sync"

	"github.com/gotomgo/coreutils/errors"
)

// ItemDiscovery is the default implementation of Discovery and NamedDiscovery
type ItemDiscovery struct {
	lock sync.RWMutex

	items         map[reflect.Type]interface{}
	baseDiscovery Discovery

	listenerLock  sync.Mutex
	typeListeners *list.List

	resolveLock  sync.Mutex
	resolveLocks map[reflect.Type]*sync.Mutex

	activeResolvers *list.List

	resolver ItemResolver
}

var _ Discovery = &ItemDiscovery{}
var _ ItemDiscoveryManagement = &ItemDiscovery{}

// NewDiscovery creates a new ItemDiscovery typed as Discovery
//
//	 Params
//	   resolver - optional ItemResolver
//
//		Notes
//			resolver is the item resolver used by discovery. if not specified, an
//			instance of BaseItemResolver is created for use
//
//			This constructor is more technically correct then
//			NewItemDiscovery which returns *ItemDiscovery
func NewDiscovery(resolver ItemResolver) Discovery {
	return NewItemDiscovery(resolver)
}

// NewDiscoveryWithBase creates a new ItemDiscovery typed as Discovery with
// a base Discovery
//
//	 Params
//	   resolver - optional ItemResolver
//
//		Notes
//			resolver is the item resolver used by discovery. if not specified, an
//			instance of BaseItemResolver is created for use
//
//			This constructor is more technically correct then
//			NewItemDiscoveryWithBase which returns *ItemDiscovery
func NewDiscoveryWithBase(baseD Discovery, resolver ItemResolver) Discovery {
	return NewItemDiscoveryWithBase(baseD, resolver)
}

// NewItemDiscovery creates and instance of ItemDiscovery
//
//	 Params
//	   resolver - optional ItemResolver
//
//		Notes
//			resolver is the item resolver used by discovery. if not specified, an
//			instance of BaseItemResolver is created for use
//
//			This constructor is considered *deprecated*. Prefer NewDiscovery
func NewItemDiscovery(resolver ItemResolver) *ItemDiscovery {
	if resolver == nil {
		resolver = NewBaseItemResolver()
	}

	return &ItemDiscovery{
		items:           map[reflect.Type]interface{}{},
		resolver:        resolver,
		resolveLocks:    map[reflect.Type]*sync.Mutex{},
		activeResolvers: &list.List{},
		typeListeners:   &list.List{},
	}
}

// NewItemDiscoveryWithBase creates and instance of ItemDiscovery with a
// base Discovery
//
//	Notes
//		resolver is the item resolver used by discovery. if not specified, an
//		instance of BaseItemResolver is created for use
//
//		This constructor is considered *deprecated*. Prefer NewDiscoveryWithBase
func NewItemDiscoveryWithBase(baseD Discovery, resolver ItemResolver) *ItemDiscovery {
	if resolver == nil {
		resolver = NewBaseItemResolver()
	}

	return &ItemDiscovery{
		baseDiscovery:   baseD,
		items:           map[reflect.Type]interface{}{},
		resolver:        resolver,
		resolveLocks:    map[reflect.Type]*sync.Mutex{},
		activeResolvers: &list.List{},
		typeListeners:   &list.List{},
	}
}

//	--------------------------------------------------------------------------
//	ItemDiscoveryManagement implementation
//	--------------------------------------------------------------------------

// AddItem adds an item for discovery by type
func (d *ItemDiscovery) AddItem(itemType reflect.Type, item interface{}) error {
	ok := reflect.TypeOf(item).ConvertibleTo(itemType)

	if !ok {
		return ErrItemNotItemType.Instance(itemType)
	}

	d.setTypedItem(itemType, item)

	return nil
}

// RemoveItem removes an item from discovery by type
func (d *ItemDiscovery) RemoveItem(itemType reflect.Type) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, ok := d.items[itemType]; ok {
		delete(d.items, itemType)
	}
}

//	--------------------------------------------------------------------------
//	Discovery implementation
//	--------------------------------------------------------------------------

// GetResolver returns the resolver being used for Discovery
func (d *ItemDiscovery) GetResolver() ItemResolver {
	return d.resolver
}

func (d *ItemDiscovery) HasItem(itemType reflect.Type) bool {
	_, ok := d.getTypedItem(itemType)
	return ok
}

func (d *ItemDiscovery) GetItem(itemType reflect.Type) (interface{}, error) {
	return d._getTypedItem(itemType, RoNone)
}

func (d *ItemDiscovery) GetRequiredItem(itemType reflect.Type) interface{} {
	item, err := d._getTypedItem(itemType, RoNone)

	if err != nil {
		panic(err)
	}

	return item
}

func (d *ItemDiscovery) GetItemWithOptions(itemType reflect.Type, options ResolveOptions) (interface{}, error) {
	return d._getTypedItem(itemType, options)
}

func (d *ItemDiscovery) GetRequiredItemWithOptions(itemType reflect.Type, options ResolveOptions) (interface{}, error) {
	return d._getTypedItem(itemType, options)
}

// WrapAO can be used to resolve an AO item wrapper when a item is NOT
// automatically wrapped because the item is created directly and NOT
// through discovery
func (d *ItemDiscovery) WrapAO(itemType reflect.Type, item interface{}) (interface{}, error) {
	return d.resolver.(AOItemResolver).WrapAO(d, itemType, item)
}

func (d *ItemDiscovery) getTypedItem(itemType reflect.Type) (item interface{}, ok bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	item, ok = d.items[itemType]
	return
}

func (d *ItemDiscovery) setTypedItem(itemType reflect.Type, item interface{}) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.items[itemType] = item
}

func (d *ItemDiscovery) _getTypedItem(itemType reflect.Type, options ResolveOptions) (interface{}, error) {
	var item interface{}
	var err error

	if (options & RoInstanceItem) != 0 {
		item, err = d.resolveItem(itemType, nil, nil)
	} else {
		var ok bool

		item, ok = d.getTypedItem(itemType)

		if !ok && ((options & RoDontResolve) == 0) {
			item, err = d.resolveItem(itemType, d.getTypedItem, func(itemType reflect.Type, item interface{}) {
				d.setTypedItem(itemType, item)
			})
		}
	}

	if errors.IsError(err) {
		return nil, err
	}

	if (item == nil) && ((options & RoInstanceItem) == 0) {
		if d.baseDiscovery != nil {
			if item, err = d.baseDiscovery.GetItemWithOptions(itemType, options); errors.IsError(err) {
				return nil, err
			}
		}
	}

	if item == nil {
		return nil, ErrItemNotFound.Instance(itemType)
	}

	return item, nil
}

type resolveCheckBack func(itemType reflect.Type) (interface{}, bool)
type resolveSetItem func(itemType reflect.Type, item interface{})

func (d *ItemDiscovery) resolveItem(itemType reflect.Type, checkBack resolveCheckBack, setItem resolveSetItem) (interface{}, error) {
	if d.resolver == nil {
		return nil, nil
	}

	// fmt.Println("Resolving ", itemType)
	// defer fmt.Println("Resolve complete for ", itemType)

	d.acquireResolveLock(itemType)
	defer d.releaseResolveLock(itemType)

	if checkBack != nil {
		if item, ok := checkBack(itemType); ok {
			return item, nil
		}
	}

	item, err := d.resolver.ResolveItem(d, itemType)

	if (item != nil) && (setItem != nil) {
		setItem(itemType, item)
	}

	return item, err
}

func (d *ItemDiscovery) isResolving(itemType reflect.Type) bool {
	for e := d.activeResolvers.Front(); e != nil; e = e.Next() {
		if e.Value.(reflect.Type) == itemType {
			return true
		}
	}

	return false
}

func (d *ItemDiscovery) removeActiveResolver(itemType reflect.Type) {
	for e := d.activeResolvers.Front(); e != nil; e = e.Next() {
		if e.Value == itemType {
			d.activeResolvers.Remove(e)
			break
		}
	}
}

func (d *ItemDiscovery) acquireResolveLock(itemType reflect.Type) {
	var resLock *sync.Mutex

	d.resolveLock.Lock()

	defer func() {
		d.resolveLock.Unlock()
		if resLock != nil {
			resLock.Lock()
		}
	}()

	if d.isResolving(itemType) {
		err := ErrCircularResolveDependency.Instance(itemType)
		panic(err)
	}

	d.activeResolvers.PushBack(itemType)
	defer d.removeActiveResolver(itemType)

	var ok bool

	// note the lock is taken deferred
	if resLock, ok = d.resolveLocks[itemType]; !ok {
		resLock = &sync.Mutex{}
		d.resolveLocks[itemType] = resLock
	}
}

func (d *ItemDiscovery) releaseResolveLock(itemType reflect.Type) {
	var resLock *sync.Mutex

	d.resolveLock.Lock()

	defer func() {
		d.resolveLock.Unlock()
		if resLock != nil {
			resLock.Unlock()
		}
	}()

	resLock, _ = d.resolveLocks[itemType]
}
