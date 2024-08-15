package gorm

import (
	"fmt"
	"strings"

	"github.com/goatquery/goatquery-go"
	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/keywords"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/parser"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SearchFunc = func(db *gorm.DB, searchTerm string) *gorm.DB

func LogQuery(db *gorm.DB) *gorm.DB {
	db.Session(&gorm.Session{}).Logger.LogMode(logger.Info)

	return db
}

func Apply(db *gorm.DB, query goatquery.Query, model interface{}, searchFunc SearchFunc, options *goatquery.QueryOptions) (*gorm.DB, *int64, error) {
	if options != nil && query.Top > options.MaxTop {
		return nil, nil, fmt.Errorf("The value supplied for the query parameter 'Top' was greater than the maximum top allowed for this resource")
	}

	// Filter
	if query.Filter != "" {
		l := lexer.NewLexer(query.Filter)
		p := parser.NewParser(l)

		statements := p.ParseFilter()

		db = EvaluateFilter(&statements.Expression, db)
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

func EvaluateFilter(exp ast.Expression, db *gorm.DB) *gorm.DB {
	switch exp := exp.(type) {
	case *ast.InfixExpression:
		identifier, ok := exp.Left.(*ast.Identifier)
		if ok {

			var value interface{}

			switch right := exp.Right.(type) {
			case *ast.StringLiteral:
				value = right.Value
			case *ast.IntegerLiteral:
				value = right.Value
			}

			switch strings.ToLower(exp.Operator) {
			case keywords.EQ:
				return db.Where(fmt.Sprintf("%s = ?", identifier.TokenLiteral()), value)
			case keywords.NE:
				return db.Where(fmt.Sprintf("%s <> ?", identifier.TokenLiteral()), value)
			case keywords.CONTAINS:
				if str, ok := exp.Right.(*ast.StringLiteral); ok {
					return db.Where(fmt.Sprintf("%s LIKE ?", identifier.TokenLiteral()), "%"+str.Value+"%")
				}
			}
		}

		switch exp.Operator {
		case keywords.AND:
			return EvaluateFilter(exp.Right, EvaluateFilter(exp.Left, db))
		case keywords.OR:
			left := EvaluateFilter(exp.Left, db)
			right := EvaluateFilter(exp.Right, db)

			return left.Or(right)
		}
	}

	return db
}
