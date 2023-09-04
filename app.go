package goatquery

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Apply(db *gorm.DB, query Query, maxTop *int, searchFunc func(db *gorm.DB, searchTerm string) *gorm.DB, model interface{}) (*gorm.DB, *int64, error) {
	if maxTop != nil && query.Top > *maxTop {
		return nil, nil, fmt.Errorf("The value supplied for the query parameter 'Top' was greater than the maximum top allowed for this resource")
	}

	if maxTop != nil && query.Top == 0 {
		// If no top query was provided, set to max top.
		query.Top = *maxTop
	}

	// Filter
	if query.Filter != "" {
		filters := splitString(query.Filter)
		where := strings.Builder{}

		namer := db.Statement.NamingStrategy

		v := reflect.ValueOf(model)
		modelType := reflect.Indirect(v).Type().Elem()

		tableName := namer.TableName(modelType.Name())

		for i, filter := range filters {
			opts := splitStringByWhitespace(strings.TrimSpace(filter))

			if len(opts) != 3 {
				continue
			}

			if i > 0 {
				prev := filters[i-1]
				where.WriteString(fmt.Sprintf(" %s ", strings.TrimSpace(prev)))
			}

			property := opts[0]
			operand := opts[1]
			value := opts[2]

			property = GetGormColumnNameByJsonTag(namer, tableName, modelType, property)

			if strings.EqualFold(operand, "contains") {
				valueWithoutQuotes := getValueBetweenQuotes(value)
				where.WriteString(fmt.Sprintf("%s %s '%%%s%%'", property, filterOperations[operand], valueWithoutQuotes))
			} else {
				where.WriteString(fmt.Sprintf("%s %s %s", property, filterOperations[operand], value))
			}
		}

		db = db.Where(where.String())
	}

	// Search
	if searchFunc != nil && query.Search != "" {
		db = searchFunc(db, query.Search)
	}

	// Count
	var count int64
	if query.Count {
		db.Count(&count)
	}

	// Order by
	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	}

	// Select
	if query.Select != "" {
		db = db.Select(query.Select)
	}

	// Skip
	if query.Skip > 0 {
		db = db.Offset(query.Skip)
	}

	// Top
	if query.Top > 0 {
		db = db.Limit(query.Top)
	}

	if query.Count {
		return db, &count, nil
	}

	return db, nil, nil
}

func GetGormColumnNameByJsonTag(namer schema.Namer, tableName string, t reflect.Type, property string) string {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		v := strings.Split(f.Tag.Get("json"), ",")[0] // use split to ignore tag "options" like omitempty, etc.
		if v == property {
			settings := schema.ParseTagSetting(f.Tag.Get("gorm"), ";")
			if settings["COLUMN"] != "" {
				return settings["COLUMN"]
			}

			return namer.ColumnName(tableName, f.Name)
		}
	}

	return property
}
