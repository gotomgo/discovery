package discovery

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRemove(t *testing.T) {
	d := NewItemDiscovery(nil)

	if d.GetResolver() == nil {
		t.Error("Expecting Discovery to have a valid Resolver")
	}

	s := "abcd"

	err := d.AddItem(reflect.TypeOf(s), s)

	if err != nil {
		t.Error(err)
	}

	if len(d.items) != 1 {
		t.Error("Expecting discovery to have 1 item")
	}

	item, ok := d.items[reflect.TypeOf(s)]

	if !ok {
		t.Error("Expecting items to contain s")
	}

	s = item.(string)

	if s != "abcd" {
		t.Error("expecting s == 'abcd'")
	}

	d.RemoveItem(reflect.TypeOf((*string)(nil)).Elem())

	if len(d.items) != 0 {
		t.Error("Expecting discovery to be empty")
	}

	err = d.AddItem(reflect.TypeOf((*float64)(nil)).Elem(), s)

	if err == nil {
		t.Error("Expecting that adding string keyed as float64 would fail")
	}

	if len(d.items) != 0 {
		t.Error("Expecting discovery to be empty")
	}
}

func TestDiscovery(t *testing.T) {
	itemDiscovery := NewItemDiscovery(mockResolver)
	var d Discovery = itemDiscovery

	ok := d.HasItem(MockServiceType)
	if ok {
		t.Error("Expecting has item to be false")
	}

	item, err := d.GetItem(MockServiceType)
	assert.NoError(t, err)

	if item == nil {
		t.Error("Expecting MockService to be resolved")
	}

	if item.(*MockService).field != 32 {
		t.Error("Expecting MockService.field == 32")
	}

	item.(*MockService).field = item.(*MockService).field + 1

	ok = d.HasItem(MockServiceType)
	if !ok {
		t.Error("Expecting has item to be true")
	}

	item, err = d.GetItemWithOptions(MockServiceType, RoInstanceItem)
	assert.NoError(t, err)
	assert.NotNil(t, item)

	if item.(*MockService).field != 32 {
		t.Error("Expecting MockService.field == 32")
	}

	item, err = d.GetRequiredItemWithOptions(MockServiceType, RoNone)
	assert.NoError(t, err)
	assert.NotNil(t, item)

	if item.(*MockService).field != 33 {
		t.Error("Expecting MockService.field == 33")
	}
}

func TestItemRequiredPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	var d Discovery = NewItemDiscovery(nil)

	_ = d.GetRequiredItem(MockServiceType)
}

type MockService struct {
	field int
}

type MockResolver struct {
	mappings map[reflect.Type]ResolverMapping
}

var MockServiceType = reflect.TypeOf(&MockService{})

func NewMockResolver() ItemResolver {
	stage := map[reflect.Type]ResolverMapping{}

	MapHelper(stage, ResolverMapping{
		Type: MockServiceType,
		Creator: func(d Discovery) (interface{}, error) {
			return &MockService{field: 32}, nil
		},
	})

	MapHelper(stage, ResolverMapping{
		Type: reflect.TypeOf("abcd"),
		Creator: func(d Discovery) (interface{}, error) {
			return "abcd", nil
		},
	})

	return &MockResolver{
		mappings: stage,
	}
}

var mockResolver ItemResolver = NewMockResolver()

func MapHelper(aMap map[reflect.Type]ResolverMapping, mapping ResolverMapping) {
	aMap[mapping.Type] = mapping
}

func (r *MockResolver) GetItem(d Discovery, itemType reflect.Type) (interface{}, error) {
	creator, ok := r.mappings[itemType]
	if !ok {
		return nil, errors.New("Normally this is not an error, but we are testing")
	}

	return creator.Creator(d)
}

func (r *MockResolver) AddMapping(mapping ResolverMapping) {

}

func (r *MockResolver) AddMappings(mapping []ResolverMapping) {

}

func (r *MockResolver) AddMappingsVar(mappings ...ResolverMapping) {

}
