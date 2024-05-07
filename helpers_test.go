package discovery

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem interface{}
type testItemImpl struct{}

var _ testItem = &testItemImpl{}

var testItemType = reflect.TypeOf((*testItem)(nil)).Elem()

func TestDefaultDiscovery(t *testing.T) {
	// should be nothing set, and GetDefaultDiscovery should still work
	assert.Nil(t, GetDefaultDiscovery())
	assert.Panics(t, func() { GetItem[*testItem](nil, testItemType) })
	assert.Panics(t, func() { GetRequiredItem[*testItem](nil, testItemType) })

	// create a known resolver so we can compare
	resolver := NewBaseItemResolver()
	resolver.AddMapping(ResolverMapping{
		Type: testItemType,
		Creator: func(d Discovery) (interface{}, error) {
			return &testItemImpl{}, nil
		}})

	d := GetOrCreateDefaultDiscovery(resolver)
	assert.NotNil(t, d)

	// default discovery should be what we just created
	d2 := GetDefaultDiscovery()
	assert.NotNil(t, d2)
	assert.Equal(t, d, d2)

	// assert that our resolver was used
	assert.Equal(t, resolver, GetDefaultResolver())

	// verify base of super discovery
	superD := CreateSuperDiscovery(nil)
	assert.NotNil(t, superD)
	assert.Equal(t, d, superD.(*ItemDiscovery).baseDiscovery)

	// re-assert that default discovery has not changed
	d2 = GetDefaultDiscovery()
	assert.NotNil(t, d2)
	assert.Equal(t, d, d2)

	item, err := GetItem[testItem](GetDefaultDiscovery(), testItemType)
	assert.NoError(t, err)
	assert.NotNil(t, item)

	item = GetRequiredItem[testItem](GetDefaultDiscovery(), testItemType)
	assert.NoError(t, err)
	assert.NotNil(t, item)
}

func TestAddItem(t *testing.T) {
	// clear any value for default discovery
	defaultD = atomic.Value{}
	assert.Nil(t, GetDefaultDiscovery())

	AddItem(testItemType, &testItemImpl{})
	assert.NotNil(t, GetDefaultDiscovery())

	item := GetRequiredItem[testItem](GetDefaultDiscovery(), testItemType)
	assert.NotNil(t, item)
}
