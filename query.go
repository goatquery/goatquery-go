// Package goatquery provides query string parsing and evaluation for Go.
package goatquery

// Query represents the parsed query string parameters.
type Query struct {
	// Top limits the number of results returned. Nil means no limit (unless MaxTop is set).
	Top *int
	// Skip specifies the number of results to skip for pagination.
	Skip *int
	// Count requests the total count of matching results in the response.
	Count *bool
	// OrderBy specifies the ordering of results (e.g., "name asc, age desc").
	OrderBy string
	// Search is a free-text search term passed to the SearchFunc callback.
	Search string
	// Filter specifies the filter expression (e.g., "age gt 25 and name eq 'John'").
	Filter string
}

// QueryOptions configures query evaluation behavior.
type QueryOptions struct {
	// MaxTop is the maximum allowed value for Top. If Top exceeds MaxTop, an error is returned.
	// When Top is nil and MaxTop > 0, MaxTop is applied as the default limit.
	MaxTop int

	// MaxPropertyMappingDepth limits how deep nested property paths can be resolved.
	// Defaults to 5 if zero or unset.
	MaxPropertyMappingDepth int
}

// GetMaxPropertyMappingDepth returns the configured depth or the default of 5.
func (o *QueryOptions) GetMaxPropertyMappingDepth() int {
	if o == nil || o.MaxPropertyMappingDepth <= 0 {
		return 5
	}
	return o.MaxPropertyMappingDepth
}
