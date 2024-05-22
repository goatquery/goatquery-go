package goatquery

type PagedResponse[T any] struct {
	Count *int64 `json:"count,omitempty"`
	Value []T    `json:"value"`
}
