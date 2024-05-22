package gorm

import (
	"fmt"

	"github.com/goatquery/goatquery-go"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/parser"
	"gorm.io/gorm"
)

type SearchFunc = func(db *gorm.DB, searchTerm string) *gorm.DB

func Apply(db *gorm.DB, query goatquery.Query, model interface{}, searchFunc SearchFunc, options *goatquery.QueryOptions) (*gorm.DB, *int64, error) {
	if options != nil && query.Top > options.MaxTop {
		return nil, nil, fmt.Errorf("The value supplied for the query parameter 'Top' was greater than the maximum top allowed for this resource")
	}

	// Order by
	if query.OrderBy != "" {
		l := lexer.NewLexer(query.OrderBy)
		p := parser.NewParser(l)

		statements := p.ParseOrderBy()

		for _, statement := range statements {
			sql := fmt.Sprintf("%s %s", statement.TokenLiteral(), statement.Direction)

			db = db.Order(sql)
		}
	}

	// Search
	if searchFunc != nil && query.Search != "" {
		db = searchFunc(db, query.Search)
	}

	// Count
	var count int64
	if query.Count {
		db.Model(&model).Count(&count)
	}

	// Skip
	if query.Skip > 0 {
		db = db.Offset(query.Skip)
	}

	// Top
	if query.Top > 0 {
		db = db.Limit(query.Top)
	}

	if query.Top <= 0 && options != nil && options.MaxTop > 0 {
		db = db.Limit(options.MaxTop)
	}

	if query.Count {
		return db, &count, nil
	}

	return db, nil, nil
}
