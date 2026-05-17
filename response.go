package goatquery

// PagedResponse wraps a collection result with an optional total count for pagination.
type PagedResponse[T any] struct {
	Count *int64 `json:"count,omitempty"`
	Value []T    `json:"value"`
}
