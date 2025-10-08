package internal

import (
	"context"
	"fmt"
)

// Cursor represents a pagination cursor
type Cursor[T any] struct {
	Items      []T             `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Cursor    *string `json:"cursor,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	Direction *string `json:"direction,omitempty"`
	HasMore   bool    `json:"has_more"`
}

// Iterator provides iteration over paginated results
type Iterator[T any] struct {
	client      RequestClient
	path        string
	params      map[string]interface{}
	cursor      *string
	limit       *int
	direction   *string
	hasMore     bool
	currentIdx  int
	currentPage []T
}

// RequestClient interface for making paginated requests
type RequestClient interface {
	DoRequestWithQuery(ctx context.Context, method, path string, query map[string]interface{}, result interface{}) error
}

// NewIterator creates a new iterator for paginated results
func NewIterator[T any](client RequestClient, path string, params map[string]interface{}) *Iterator[T] {
	limit, _ := params["limit"].(int)
	direction, _ := params["direction"].(string)
	cursor, _ := params["cursor"].(string)

	return &Iterator[T]{
		client:    client,
		path:      path,
		params:    params,
		cursor:    &cursor,
		limit:     &limit,
		direction: &direction,
		hasMore:   true,
	}
}

// Next returns the next item in the iteration
func (it *Iterator[T]) Next(ctx context.Context) (*T, error) {
	// If we have items in current page, return next one
	if it.currentIdx < len(it.currentPage) {
		item := &it.currentPage[it.currentIdx]
		it.currentIdx++
		return item, nil
	}

	// If no more pages, return done
	if !it.hasMore {
		return nil, nil
	}

	// Fetch next page
	if err := it.fetchNextPage(ctx); err != nil {
		return nil, err
	}

	// Return first item from new page
	if len(it.currentPage) > 0 {
		item := &it.currentPage[0]
		it.currentIdx = 1
		return item, nil
	}

	return nil, nil
}

// HasNext returns true if there are more items to iterate
func (it *Iterator[T]) HasNext() bool {
	return it.currentIdx < len(it.currentPage) || it.hasMore
}

// fetchNextPage fetches the next page of results
func (it *Iterator[T]) fetchNextPage(ctx context.Context) error {
	params := make(map[string]interface{})
	for k, v := range it.params {
		params[k] = v
	}

	if it.cursor != nil && *it.cursor != "" {
		params["cursor"] = *it.cursor
	}
	if it.limit != nil && *it.limit > 0 {
		params["limit"] = *it.limit
	}
	if it.direction != nil && *it.direction != "" {
		params["direction"] = *it.direction
	}

	var response Cursor[T]
	if err := it.client.DoRequestWithQuery(ctx, "GET", it.path, params, &response); err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}

	it.currentPage = response.Items
	it.currentIdx = 0

	if response.Pagination != nil {
		it.cursor = response.Pagination.Cursor
		it.hasMore = response.Pagination.HasMore
	} else {
		it.hasMore = false
	}

	return nil
}

// ToSlice collects all remaining items into a slice
func (it *Iterator[T]) ToSlice(ctx context.Context) ([]T, error) {
	var items []T

	for it.HasNext() {
		item, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		items = append(items, *item)
	}

	return items, nil
}
