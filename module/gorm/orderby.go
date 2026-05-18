package gorm

import (
	"fmt"
	"reflect"

	"github.com/goatquery/goatquery-go/ast"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func evaluateOrderBy(statements []ast.OrderByStatement, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int, joined map[string]bool) (*gorm.DB, error) {
	for _, stmt := range statements {
		prop, err := resolveProperty(namer, t, stmt.Segments, maxDepth)
		if err != nil {
			return nil, err
		}

		// Apply joins for nested properties
		for _, join := range prop.Joins {
			if !joined[join] {
				rawJoin, err := buildRawJoin(namer, t, join)
				if err != nil {
					return nil, err
				}
				db = db.Joins(rawJoin)
				joined[join] = true
			}
		}

		clause := fmt.Sprintf("%s %s", prop.Column, stmt.Direction)
		db = db.Order(clause)
	}

	return db, nil
}
