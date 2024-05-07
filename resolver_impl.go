package discovery

import (
	"reflect"
	"sync"

	"github.com/gotomgo/coreutils/errors"
)

// BaseItemResolver provides item creation mappings
type BaseItemResolver struct {
	lock       sync.Mutex
	mappings   map[reflect.Type]ResolverMapping
	aoMappings map[reflect.Type][]AOResolverMapping
}

// ensure we are an implementation of AOItemResolver
var _ AOItemResolver = &BaseItemResolver{}

// NewBaseItemResolver creates an instance of BaseItemResolver
func NewBaseItemResolver() *BaseItemResolver {
	return &BaseItemResolver{
		mappings:   map[reflect.Type]ResolverMapping{},
		aoMappings: map[reflect.Type][]AOResolverMapping{},
	}
}

// addMapping adds a ResolverMapping to the BaseItemResolver
func (r *BaseItemResolver) addMapping(mapping ResolverMapping) {
	r.mappings[mapping.Type] = mapping
}

// AddMapping adds a ResolverMapping to the BaseItemResolver
func (r *BaseItemResolver) AddMapping(mapping ResolverMapping) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.addMapping(mapping)
}

// AddMappings adds an [] of ResolverMapping's to the BaseItemResolver
func (r *BaseItemResolver) AddMappings(mappings []ResolverMapping) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, mapping := range mappings {
		r.addMapping(mapping)
	}
}

// AddMappingsVar adds a variadic list of ResolverMapping's to the BaseItemResolver
func (r *BaseItemResolver) AddMappingsVar(mappings ...ResolverMapping) {
	r.AddMappings(mappings)
}

// GetMapping returns a ResolverMapping for itemType, if available
func (r *BaseItemResolver) GetMapping(itemType reflect.Type) (ResolverMapping, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	result, ok := r.mappings[itemType]
	return result, ok
}

// ResolveItem returns an instance of itemType via its creator
func (r *BaseItemResolver) ResolveItem(d Discovery, itemType reflect.Type) (interface{}, error) {
	creator, ok := r.GetMapping(itemType)
	if !ok {
		return nil, nil
	}

	result, err := creator.Creator(d)
	if errors.IsError(err) {
		err = ErrItemNotResolved.Instance(itemType.Name(), err).WithInner(err)
		return nil, err
	}

	return r.WrapAO(d, itemType, result)
}

func (r *BaseItemResolver) ResolveMapping(d Discovery, mapping ResolverMapping) (interface{}, error) {
	result, err := mapping.Creator(d)
	if errors.IsError(err) {
		err = ErrItemNotResolved.Instance(mapping.Type.Name(), err).WithInner(err)
		return nil, err
	}

	return r.WrapAO(d, mapping.Type, result)
}

// GetAOMappings returns an []AOMapping for itemType, if available
func (r *BaseItemResolver) GetAOMappings(itemType reflect.Type) (result []AOResolverMapping, ok bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.getAOMappings(itemType)
}

// GetAOMappings returns an []AOMapping for itemType, if available
func (r *BaseItemResolver) getAOMappings(itemType reflect.Type) (result []AOResolverMapping, ok bool) {
	result, ok = r.aoMappings[itemType]
	return
}

// addMapping adds a ResolverMapping to the BaseItemResolver
func (r *BaseItemResolver) addAOMapping(mapping AOResolverMapping) {
	var mappings []AOResolverMapping
	mappings = r.aoMappings[mapping.Type]
	r.aoMappings[mapping.Type] = append(mappings, mapping)
}

// AddAOMapping adds an AOMapping to the AOItemResolver
func (r *BaseItemResolver) AddAOMapping(mapping AOResolverMapping) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.addAOMapping(mapping)
}

// AddAOMappings adds a collection of AOMapping to the AOItemResolver
func (r *BaseItemResolver) AddAOMappings(mappings []AOResolverMapping) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, mapping := range mappings {
		r.addAOMapping(mapping)
	}
}

// WrapAO wraps a core item with 0 or more AO items
func (r *BaseItemResolver) WrapAO(d Discovery, itemType reflect.Type, item interface{}) (result interface{}, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// we need to return the core item in the case there are no AO mapping
	result = item

	if mappings, ok := r.getAOMappings(itemType); ok {
		// create item wrappers in reverse order of registration so that what
		// is registered first is 1st wrapper, 2nd is 2nd, and so on
		for i := len(mappings) - 1; i >= 0; i-- {
			if result, err = mappings[i].Creator(d, result); errors.IsError(err) {
				err = ErrItemNotResolved.Instancef("ao mapping %s failed resolve: %s", mappings[i].Type, err).WithInner(err)
				result = nil
				return
			}
		}
	}

	return
}
