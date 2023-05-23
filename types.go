package goatquery

type PagedResponse[T any] struct {
	Count *int64 `json:"count,omitempty"`
	Value []T    `json:"value"`
}

type QueryErrorResponse struct {
	Status  uint   `json:"status"`
	Message string `json:"message"`
}

type Query struct {
	Top     int
	Skip    int
	Count   bool
	OrderBy string
	Select  string
	Search  string
	Filter  string
}
