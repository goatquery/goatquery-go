package gorm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/schema"
)

type resolvedProperty struct {
	Column    string
	FieldType reflect.Type // the Go type of the resolved field
	Joins     []string     // join clauses needed for nested properties
}

func resolveProperty(namer schema.Namer, t reflect.Type, segments []string, maxDepth int) (*resolvedProperty, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("Empty property path.")
	}

	if len(segments) > maxDepth {
		return nil, fmt.Errorf("Property path exceeds maximum depth of '%d'.", maxDepth)
	}

	var joins []string
	currentType := t
	tableName := namer.TableName(currentType.Name())

	for i, segment := range segments {
		isLast := i == len(segments)-1

		field, ok := findField(currentType, segment)
		if !ok {
			return nil, fmt.Errorf("Property '%s' does not exist on type '%s'.", segment, currentType.Name())
		}

		if isLast {
			fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct && fieldType != reflect.TypeOf(time.Time{}) && fieldType != reflect.TypeOf(uuid.UUID{}) {
				s, err := schema.Parse(reflect.New(currentType).Interface(), &sync.Map{}, namer)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse schema for '%s': %w", currentType.Name(), err)
				}
				for _, r := range s.Relationships.Relations {
					if r.Field.Name == field.Name && len(r.References) > 0 {
						ref := r.References[0]
						fkCol := namer.ColumnName(tableName, ref.ForeignKey.DBName)
						if len(joins) > 0 {
							alias := strings.ReplaceAll(joins[len(joins)-1], ".", "__")
							fkCol = fmt.Sprintf("%q.%s", alias, fkCol)
						} else {
							fkCol = fmt.Sprintf("%q.%s", tableName, fkCol)
						}
						return &resolvedProperty{Column: fkCol, FieldType: fieldType, Joins: joins}, nil
					}
				}
				return nil, fmt.Errorf("Relationship '%s' not found on type '%s'.", segment, currentType.Name())
			}

			col := getColumnName(namer, tableName, field)
			// Qualify column references to avoid ambiguity in self-joins.
			if len(joins) > 0 {
				alias := strings.ReplaceAll(joins[len(joins)-1], ".", "__")
				col = fmt.Sprintf("%q.%s", alias, col)
			} else {
				col = fmt.Sprintf("%q.%s", tableName, col)
			}
			return &resolvedProperty{Column: col, FieldType: field.Type, Joins: joins}, nil
		}

		// Navigate into nested type
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
		}

		if fieldType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Property '%s' is not a navigable type.", segment)
		}

		if len(joins) == 0 {
			joins = append(joins, field.Name)
		} else {
			joins = append(joins, joins[len(joins)-1]+"."+field.Name)
		}
		tableName = namer.TableName(fieldType.Name())
		currentType = fieldType
	}

	return nil, fmt.Errorf("Unexpected end of property path.")
}

func findField(t reflect.Type, name string) (reflect.StructField, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous {
			if found, ok := findField(field.Type, name); ok {
				return found, true
			}
			continue
		}

		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = field.Name
		}

		if strings.EqualFold(jsonName, name) {
			return field, true
		}

		// Also try matching against the field name directly
		if strings.EqualFold(field.Name, name) {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

func getColumnName(namer schema.Namer, tableName string, field reflect.StructField) string {
	settings := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	if col := settings["COLUMN"]; col != "" {
		return col
	}
	return namer.ColumnName(tableName, field.Name)
}

func findRelationship(namer schema.Namer, t reflect.Type, name string) (*schema.Relationship, error) {
	s, err := schema.Parse(reflect.New(t).Interface(), &sync.Map{}, namer)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse schema for '%s': %w", t.Name(), err)
	}

	for _, r := range s.Relationships.Relations {
		jsonName := strings.Split(r.Field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = r.Field.Name
		}
		if strings.EqualFold(jsonName, name) || strings.EqualFold(r.Field.Name, name) {
			return r, nil
		}
	}

	return nil, nil
}
