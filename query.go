package goatquery

type Query struct {
	Top     int
	Skip    int
	Count   bool
	OrderBy string
	Select  string
	Search  string
	Filter  string
}

type QueryOptions struct {
	MaxTop int
}
