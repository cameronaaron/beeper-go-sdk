package beeperdesktop

import (
	"context"

	"github.com/cameronaaron/beeper-go-sdk/internal"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

// Iterator provides a way to iterate through paginated results
type Iterator[T any] struct {
	iterator *internal.Iterator[T]
}

// NewIterator creates a new iterator for paginated results
func NewIterator[T any](client *BeeperDesktop, path string, params map[string]interface{}) *Iterator[T] {
	return &Iterator[T]{
		iterator: internal.NewIterator[T](client, path, params),
	}
}

// Next returns the next item in the iteration
func (it *Iterator[T]) Next(ctx context.Context) (*T, error) {
	return it.iterator.Next(ctx)
}

// HasNext returns true if there are more items to iterate
func (it *Iterator[T]) HasNext() bool {
	return it.iterator.HasNext()
}

// ToSlice collects all remaining items into a slice
func (it *Iterator[T]) ToSlice(ctx context.Context) ([]T, error) {
	return it.iterator.ToSlice(ctx)
}

// NewMessageIterator creates an iterator for message search results
func (c *BeeperDesktop) NewMessageIterator(params resources.MessageSearchParams) *Iterator[resources.Message] {
	paramMap := map[string]interface{}{
		"accountIDs":         params.AccountIDs,
		"chatIDs":            params.ChatIDs,
		"chatType":           params.ChatType,
		"cursor":             params.Cursor,
		"dateAfter":          params.DateAfter,
		"dateBefore":         params.DateBefore,
		"direction":          params.Direction,
		"excludeLowPriority": params.ExcludeLowPriority,
		"includeMuted":       params.IncludeMuted,
		"limit":              params.Limit,
		"mediaTypes":         params.MediaTypes,
		"query":              params.Query,
		"senderIDs":          params.SenderIDs,
	}
	return NewIterator[resources.Message](c, "/v0/search-messages", paramMap)
}

// NewChatIterator creates an iterator for chat search results
func (c *BeeperDesktop) NewChatIterator(params resources.ChatSearchParams) *Iterator[resources.Chat] {
	paramMap := map[string]interface{}{
		"accountIDs":   params.AccountIDs,
		"chatType":     params.ChatType,
		"includeMuted": params.IncludeMuted,
		"limit":        params.Limit,
		"cursor":       params.Cursor,
		"scope":        params.Scope,
		"query":        params.Query,
	}
	return NewIterator[resources.Chat](c, "/v0/search-chats", paramMap)
}
