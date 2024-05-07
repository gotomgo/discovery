package discovery

import (
	"net/http"

	"github.com/gotomgo/coreutils/errors"
)

const (
	// ErrItemNotFoundID represents a service that cannot be found because
	// there is no mapping, or existing service item
	ErrItemNotFoundID = "discovery/item/notfound"

	// ErrItemNotResolvedID represents a service that has a mapping but
	// failed to resolve
	ErrItemNotResolvedID = "discovery/item/resolve/failed"

	// ErrItemNotItemTypeID represent an item that does not implement
	// the specified item type
	ErrItemNotItemTypeID = "discovery/item/must-be-item-type"

	// ErrCircularResolveDependencyID indicates a circular dependency between
	// items
	ErrCircularResolveDependencyID = "discovery/item/resolve/circular"
)

var (
	ErrItemNotFound = errors.NewErrorTemplate(
		ErrItemNotFoundID,
		"item '%s' not found",
		http.StatusInternalServerError,
		false)

	ErrItemNotResolved = errors.NewErrorTemplate(
		ErrItemNotResolvedID,
		"item '%s' not resolved: %s",
		http.StatusInternalServerError,
		false)

	ErrItemNotItemType = errors.NewErrorTemplate(
		ErrItemNotItemTypeID,
		"item does not implement type %s",
		http.StatusInternalServerError,
		false)

	ErrCircularResolveDependency = errors.NewErrorTemplate(
		ErrCircularResolveDependencyID,
		"item type %s has a circular resolve dependency",
		http.StatusInternalServerError,
		false)
)
