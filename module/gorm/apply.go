// Package gorm provides GoatQuery integration with GORM, translating parsed
// filter, orderby, top, skip, count, and search queries into GORM scopes.
package gorm

import (
	"fmt"
	"reflect"

	goatquery "github.com/goatquery/goatquery-go"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/parser"
	"gorm.io/gorm"
)

// SearchFunc is a user-provided function that applies search filtering to a GORM query.
type SearchFunc = func(db *gorm.DB, searchTerm string) *gorm.DB

// Apply applies the query parameters (filter, search, count, orderby, skip, top) to a GORM query.
// Pipeline order: validate → filter → search → count → orderby → skip → top.
func Apply[T any](db *gorm.DB, query goatquery.Query, searchFunc SearchFunc, options *goatquery.QueryOptions) (*gorm.DB, *int64, error) {
	maxDepth := 5
	if options != nil {
		maxDepth = options.GetMaxPropertyMappingDepth()
	}

	if options != nil && options.MaxTop > 0 && query.Top != nil && *query.Top > options.MaxTop {
		return nil, nil, fmt.Errorf("The value for query parameter 'top' exceeds the maximum allowed value of '%d'.", options.MaxTop)
	}

	var model T
	t := reflect.TypeOf(model)
	namer := db.Statement.NamingStrategy
	joined := make(map[string]bool)

	// Filter
	if query.Filter != "" {
		l := lexer.NewLexer(query.Filter)
		p := parser.NewParser(l)

		filter, err := p.ParseFilter()
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to parse filter: %w", err)
		}

		db, err = evaluateFilter(filter, db, namer, t, maxDepth, joined, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	// Search
	if searchFunc != nil && query.Search != "" {
		db = searchFunc(db, query.Search)
	}

	// Count
	var count *int64
	if query.Count != nil && *query.Count {
		var c int64
		if err := db.Model(new(T)).Count(&c).Error; err != nil {
			return nil, nil, fmt.Errorf("Failed to count: %w", err)
		}
		count = &c
	}

	// OrderBy
	if query.OrderBy != "" {
		l := lexer.NewLexer(query.OrderBy)
		p := parser.NewParser(l)

		statements, err := p.ParseOrderBy()
		if err != nil {
			return nil, count, fmt.Errorf("Failed to parse orderby: %w", err)
		}

		db, err = evaluateOrderBy(statements, db, namer, t, maxDepth, joined)
		if err != nil {
			return nil, count, err
		}
	}

	// Skip
	if query.Skip != nil && *query.Skip > 0 {
		db = db.Offset(*query.Skip)
	}

	// Top
	if query.Top != nil && *query.Top > 0 {
		db = db.Limit(*query.Top)
	} else if (query.Top == nil || *query.Top <= 0) && options != nil && options.MaxTop > 0 {
		db = db.Limit(options.MaxTop)
	}

	return db, count, nil
}
