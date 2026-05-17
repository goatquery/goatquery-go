package gorm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func evaluateFilter(expr ast.Expression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int, joined map[string]bool, rootType *reflect.Type) (*gorm.DB, error) {
	return evaluateFilterWith(expr, db, namer, t, maxDepth, rootType, func(db *gorm.DB, join string) (*gorm.DB, error) {
		if !joined[join] {
			rawJoin, err := buildRawJoin(namer, t, join)
			if err != nil {
				return nil, err
			}
			db = db.Joins(rawJoin)
			joined[join] = true
		}
		return db, nil
	})
}

func evaluateFilterWithRawJoins(expr ast.Expression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int) (*gorm.DB, error) {
	applied := make(map[string]bool)
	return evaluateFilterWith(expr, db, namer, t, maxDepth, nil, func(db *gorm.DB, join string) (*gorm.DB, error) {
		if !applied[join] {
			rawJoin, err := buildRawJoin(namer, t, join)
			if err != nil {
				return nil, err
			}
			db = db.Joins(rawJoin)
			applied[join] = true
		}
		return db, nil
	})
}

func evaluateFilterWith(expr ast.Expression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int, rootType *reflect.Type, applyJoin func(*gorm.DB, string) (*gorm.DB, error)) (*gorm.DB, error) {
	// Collect all joins needed by the expression tree
	allJoins, err := collectJoins(expr, namer, t, maxDepth, rootType)
	if err != nil {
		return nil, err
	}

	for _, join := range allJoins {
		db, err = applyJoin(db, join)
		if err != nil {
			return nil, err
		}
	}

	return evaluateWhere(expr, db, namer, t, maxDepth, rootType)
}

func evaluateWhere(expr ast.Expression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int, rootType *reflect.Type) (*gorm.DB, error) {
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return evaluateInfixWhere(e, db, namer, t, maxDepth, rootType)
	case *ast.LambdaExpression:
		return evaluateLambda(e, db, namer, t, maxDepth)
	default:
		return nil, fmt.Errorf("Unsupported expression type '%T'.", expr)
	}
}

func buildRawJoin(namer schema.Namer, t reflect.Type, joinPath string) (string, error) {
	segments := strings.Split(joinPath, ".")
	currentType := t
	parentAlias := namer.TableName(currentType.Name())

	for i, seg := range segments {
		s, err := schema.Parse(reflect.New(currentType).Interface(), &sync.Map{}, namer)
		if err != nil {
			return "", fmt.Errorf("Failed to parse schema: %w", err)
		}

		var rel *schema.Relationship
		for _, r := range s.Relationships.Relations {
			if r.Field.Name == seg {
				rel = r
				break
			}
		}
		if rel == nil {
			return "", fmt.Errorf("Relationship '%s' not found on type '%s'.", seg, currentType.Name())
		}
		if len(rel.References) == 0 {
			return "", fmt.Errorf("No references found for relationship '%s'.", seg)
		}

		// Build the alias: for depth 0 it's just the field name, for deeper it uses "__"
		var alias string
		if i == 0 {
			alias = seg
		} else {
			alias = strings.Join(segments[:i+1], "__")
		}

		// Only generate the JOIN SQL for the last segment
		if i == len(segments)-1 {
			childTable := namer.TableName(rel.FieldSchema.ModelType.Name())
			parentTableForCol := namer.TableName(currentType.Name())

			ref := rel.References[0]
			var fkCol, pkCol string
			if ref.OwnPrimaryKey {
				fkCol = fmt.Sprintf("%q.%s", alias, namer.ColumnName(childTable, ref.ForeignKey.DBName))
				pkCol = fmt.Sprintf("%q.%s", parentAlias, namer.ColumnName(parentTableForCol, ref.PrimaryKey.DBName))
			} else {
				fkCol = fmt.Sprintf("%q.%s", parentAlias, namer.ColumnName(parentTableForCol, ref.ForeignKey.DBName))
				pkCol = fmt.Sprintf("%q.%s", alias, namer.ColumnName(childTable, ref.PrimaryKey.DBName))
			}

			return fmt.Sprintf("LEFT JOIN %q %q ON %s = %s", childTable, alias, fkCol, pkCol), nil
		}

		// Not the last segment: advance to the next type and update parent alias
		parentAlias = alias
		currentType = rel.FieldSchema.ModelType
	}

	return "", fmt.Errorf("Unexpected end of join path.")
}

func collectJoins(expr ast.Expression, namer schema.Namer, t reflect.Type, maxDepth int, rootType *reflect.Type) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	var walk func(expr ast.Expression) error
	walk = func(expr ast.Expression) error {
		switch e := expr.(type) {
		case *ast.InfixExpression:
			op := strings.ToLower(e.Operator)
			if op == token.And || op == token.Or {
				if err := walk(e.Left); err != nil {
					return err
				}
				return walk(e.Right)
			}
			// Comparison: resolve property to find joins
			segments := getPropertySegments(e.Left)
			if segments == nil {
				return nil
			}
			prop, err := resolveProperty(namer, t, segments, maxDepth)
			if err != nil {
				// If inside a lambda body, try resolving against the root type
				if rootType != nil {
					_, err = resolveProperty(namer, *rootType, segments, maxDepth)
					if err != nil {
						return err
					}
					// Root property joins are on the outer query, not the EXISTS subquery.
					// Don't collect them here — they're handled by the outer evaluateFilter.
					return nil
				}
				return err
			}
			for _, join := range prop.Joins {
				if !seen[join] {
					seen[join] = true
					result = append(result, join)
				}
			}
		case *ast.LambdaExpression:
			// Lambda handles its own joins internally; no outer joins needed
		}
		return nil
	}

	if err := walk(expr); err != nil {
		return nil, err
	}
	return result, nil
}

func evaluateInfixWhere(e *ast.InfixExpression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int, rootType *reflect.Type) (*gorm.DB, error) {
	op := strings.ToLower(e.Operator)

	if op == token.And || op == token.Or {
		freshDB := db.Session(&gorm.Session{NewDB: true})
		left, err := evaluateWhere(e.Left, freshDB, namer, t, maxDepth, rootType)
		if err != nil {
			return nil, err
		}
		right, err := evaluateWhere(e.Right, freshDB, namer, t, maxDepth, rootType)
		if err != nil {
			return nil, err
		}

		if op == token.And {
			return db.Where(left).Where(right), nil
		}
		return db.Where(freshDB.Where(left).Or(right)), nil
	}

	segments := getPropertySegments(e.Left)
	if segments == nil {
		return nil, fmt.Errorf("Expected property on left side of comparison, got '%T'.", e.Left)
	}

	prop, err := resolveProperty(namer, t, segments, maxDepth)
	if err != nil {
		// If inside a lambda body, fall back to the root entity type
		if rootType != nil {
			prop, err = resolveProperty(namer, *rootType, segments, maxDepth)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	value, err := extractLiteralValue(e.Right)
	if err != nil {
		return nil, err
	}

	return applyComparison(db, prop.Column, op, value, e.Right, prop.FieldType)
}

func applyComparison(db *gorm.DB, column, op string, value interface{}, rightExpr ast.Expression, fieldType reflect.Type) (*gorm.DB, error) {
	underlyingType := fieldType
	if underlyingType.Kind() == reflect.Pointer {
		underlyingType = underlyingType.Elem()
	}

	if op == token.Contains {
		if underlyingType.Kind() != reflect.String {
			return nil, fmt.Errorf("The 'contains' operator can only be used with string properties, not '%s'.", underlyingType.Name())
		}
	}

	if op == token.Lt || op == token.Lte || op == token.Gt || op == token.Gte {
		switch {
		case underlyingType.Kind() == reflect.String:
			return nil, fmt.Errorf("Operator '%s' is not supported for type '%s'.", op, underlyingType.Name())
		case underlyingType.Kind() == reflect.Bool:
			return nil, fmt.Errorf("Operator '%s' is not supported for type '%s'.", op, underlyingType.Name())
		case underlyingType == reflect.TypeOf(uuid.UUID{}):
			return nil, fmt.Errorf("Operator '%s' is not supported for type '%s'.", op, underlyingType.Name())
		}
	}

	if dateLit, ok := rightExpr.(*ast.DateLiteral); ok {
		return applyDateRangeComparison(db, column, op, dateLit.Value)
	}

	// UUID string fallback: parse string as UUID when target field is uuid.UUID
	isString := isStringLiteral(rightExpr)
	if isString && underlyingType == reflect.TypeOf(uuid.UUID{}) {
		if strVal, ok := value.(string); ok {
			if parsed, err := uuid.Parse(strVal); err == nil {
				value = parsed
				isString = false
			}
		}
	}

	isNull := isNullLiteral(rightExpr)

	switch op {
	case token.Eq:
		if isNull {
			return db.Where(fmt.Sprintf("%s IS NULL", column)), nil
		}
		if isString {
			return db.Where(fmt.Sprintf("%s IS NOT NULL AND LOWER(%s) = LOWER(?)", column, column), value), nil
		}
		return db.Where(fmt.Sprintf("%s = ?", column), value), nil

	case token.Ne:
		if isNull {
			return db.Where(fmt.Sprintf("%s IS NOT NULL", column)), nil
		}
		if isString {
			return db.Where(fmt.Sprintf("%s IS NULL OR LOWER(%s) != LOWER(?)", column, column), value), nil
		}
		return db.Where(fmt.Sprintf("%s != ?", column), value), nil

	case token.Contains:
		escaped := escapeLike(value.(string))
		pattern := "%" + escaped + "%"
		return db.Where(fmt.Sprintf("%s IS NOT NULL AND LOWER(%s) LIKE LOWER(?) ESCAPE '\\'", column, column), pattern), nil

	case token.Lt:
		return db.Where(fmt.Sprintf("%s < ?", column), value), nil
	case token.Lte:
		return db.Where(fmt.Sprintf("%s <= ?", column), value), nil
	case token.Gt:
		return db.Where(fmt.Sprintf("%s > ?", column), value), nil
	case token.Gte:
		return db.Where(fmt.Sprintf("%s >= ?", column), value), nil

	default:
		return nil, fmt.Errorf("Unsupported operator '%s'.", op)
	}
}

func evaluateLambda(e *ast.LambdaExpression, db *gorm.DB, namer schema.Namer, t reflect.Type, maxDepth int) (*gorm.DB, error) {
	collSegments := getPropertySegments(e.Property)
	if collSegments == nil {
		return nil, fmt.Errorf("Expected property for lambda collection, got '%T'.", e.Property)
	}

	// Walk intermediate navigation segments to find the final collection's parent type.
	parentType := t
	var lastIntermediateFieldName string
	for i := 0; i < len(collSegments)-1; i++ {
		seg := collSegments[i]
		r, err := findRelationship(namer, parentType, seg)
		if err != nil {
			return nil, err
		}
		if r == nil {
			return nil, fmt.Errorf("Relationship '%s' not found on type '%s'.", seg, parentType.Name())
		}
		db = db.Joins(r.Field.Name)
		lastIntermediateFieldName = r.Field.Name
		parentType = r.FieldSchema.ModelType
	}

	collName := collSegments[len(collSegments)-1]

	rel, err := findRelationship(namer, parentType, collName)
	if err != nil {
		return nil, err
	}

	if rel == nil {
		return evaluatePrimitiveLambda(e, db, namer, parentType, collName)
	}

	elemType := rel.FieldSchema.ModelType
	elemTableName := namer.TableName(elemType.Name())

	var fkConditions []string
	var parentTableRef string
	if lastIntermediateFieldName != "" {
		parentTableRef = fmt.Sprintf("%q", lastIntermediateFieldName)
	} else {
		parentTableRef = namer.TableName(parentType.Name())
	}
	parentTableNameForCol := namer.TableName(parentType.Name())
	for _, ref := range rel.References {
		parentCol := namer.ColumnName(parentTableNameForCol, ref.PrimaryKey.DBName)
		childCol := namer.ColumnName(elemTableName, ref.ForeignKey.DBName)
		fkConditions = append(fkConditions, fmt.Sprintf("%s.%s = %s.%s", elemTableName, childCol, parentTableRef, parentCol))
	}
	fkClause := strings.Join(fkConditions, " AND ")

	bodyDB := db.Session(&gorm.Session{NewDB: true}).Model(reflect.New(elemType).Interface())

	rewrittenBody := rewriteLambdaBody(e.Body, e.Parameter)

	bodyDB, err = evaluateFilter(rewrittenBody, bodyDB, namer, elemType, maxDepth, make(map[string]bool), &t)
	if err != nil {
		return nil, fmt.Errorf("Failed to evaluate lambda body: %w", err)
	}

	bodySQL := bodyDB.Where(fkClause)

	switch e.Function {
	case "any":
		return db.Where("EXISTS (?)", bodySQL.Select("1")), nil
	case "all":
		existsAny := db.Session(&gorm.Session{NewDB: true}).Table(elemTableName).Where(fkClause).Select("1")

		pkCol := namer.ColumnName(elemTableName, rel.FieldSchema.PrimaryFields[0].DBName)
		qualifiedPK := fmt.Sprintf("%s.%s", elemTableName, pkCol)

		bodyPKs := db.Session(&gorm.Session{NewDB: true}).Table(elemTableName).Where(fkClause)
		rewrittenBody2 := rewriteLambdaBody(e.Body, e.Parameter)

		bodyPKs, err = evaluateFilterWithRawJoins(rewrittenBody2, bodyPKs, namer, elemType, maxDepth)
		if err != nil {
			return nil, fmt.Errorf("Failed to evaluate lambda body: %w", err)
		}
		bodyPKs = bodyPKs.Select(qualifiedPK)

		violators := db.Session(&gorm.Session{NewDB: true}).
			Table(elemTableName).
			Where(fkClause).
			Where(fmt.Sprintf("%s NOT IN (?)", qualifiedPK), bodyPKs).
			Select("1")

		return db.Where("EXISTS (?) AND NOT EXISTS (?)", existsAny, violators), nil
	default:
		return nil, fmt.Errorf("Unsupported lambda function '%s'.", e.Function)
	}
}

func evaluatePrimitiveLambda(e *ast.LambdaExpression, db *gorm.DB, namer schema.Namer, parentType reflect.Type, collName string) (*gorm.DB, error) {
	field, ok := findField(parentType, collName)
	if !ok {
		return nil, fmt.Errorf("Property '%s' does not exist on type '%s'.", collName, parentType.Name())
	}

	ft := field.Type
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	if ft.Kind() != reflect.Slice {
		return nil, fmt.Errorf("Property '%s' is not a collection.", collName)
	}

	if e.Function != "any" {
		return nil, fmt.Errorf("Lambda function '%s' is not supported for primitive collections.", e.Function)
	}

	rewrittenBody := rewriteLambdaBody(e.Body, e.Parameter)
	infix, ok := rewrittenBody.(*ast.InfixExpression)
	if !ok {
		return nil, fmt.Errorf("Primitive collection lambda body must be a simple comparison.")
	}

	op := strings.ToLower(infix.Operator)
	if op != token.Eq {
		return nil, fmt.Errorf("Primitive collection lambda only supports 'eq' operator, got '%s'.", op)
	}

	value, err := extractLiteralValue(infix.Right)
	if err != nil {
		return nil, err
	}

	col := getColumnName(namer, namer.TableName(parentType.Name()), field)
	jsonValue, err := json.Marshal([]interface{}{value})
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal value for jsonb containment: %w", err)
	}

	return db.Where(fmt.Sprintf("%s @> ?::jsonb", col), string(jsonValue)), nil
}

func rewriteLambdaBody(expr ast.Expression, param string) ast.Expression {
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Token:    e.Token,
			Left:     rewriteLambdaBody(e.Left, param),
			Operator: e.Operator,
			Right:    rewriteLambdaBody(e.Right, param),
		}
	case *ast.PropertyPath:
		if len(e.Segments) > 1 && strings.EqualFold(e.Segments[0], param) {
			newSegments := e.Segments[1:]
			if len(newSegments) == 1 {
				return &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: newSegments[0]}}
			}
			return &ast.PropertyPath{Token: e.Token, Segments: newSegments}
		}
		return e
	case *ast.Identifier:
		return e
	case *ast.LambdaExpression:
		return &ast.LambdaExpression{
			Token:     e.Token,
			Property:  rewriteLambdaBody(e.Property, param),
			Function:  e.Function,
			Parameter: e.Parameter,
			Body:      rewriteLambdaBody(e.Body, param),
		}
	default:
		return e
	}
}

func getPropertySegments(expr ast.Expression) []string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return []string{e.Token.Literal}
	case *ast.PropertyPath:
		return e.Segments
	default:
		return nil
	}
}

func applyDateRangeComparison(db *gorm.DB, column, op string, date time.Time) (*gorm.DB, error) {
	startOfDay := date
	startOfNextDay := date.AddDate(0, 0, 1)

	switch op {
	case token.Eq:
		return db.Where(fmt.Sprintf("%s >= ? AND %s < ?", column, column), startOfDay, startOfNextDay), nil
	case token.Ne:
		return db.Where(fmt.Sprintf("%s < ? OR %s >= ?", column, column), startOfDay, startOfNextDay), nil
	case token.Lt:
		return db.Where(fmt.Sprintf("%s < ?", column), startOfDay), nil
	case token.Lte:
		return db.Where(fmt.Sprintf("%s < ?", column), startOfNextDay), nil
	case token.Gt:
		return db.Where(fmt.Sprintf("%s >= ?", column), startOfNextDay), nil
	case token.Gte:
		return db.Where(fmt.Sprintf("%s >= ?", column), startOfDay), nil
	default:
		return nil, fmt.Errorf("Unsupported operator '%s'.", op)
	}
}

func extractLiteralValue(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.IntegerLiteral:
		return e.Value, nil
	case *ast.FloatLiteral:
		return e.Value, nil
	case *ast.UUIDLiteral:
		return e.Value, nil
	case *ast.DateTimeLiteral:
		return e.Value, nil
	case *ast.DateLiteral:
		return e.Value, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	default:
		return nil, fmt.Errorf("Unsupported literal type '%T'.", expr)
	}
}

func isStringLiteral(expr ast.Expression) bool {
	_, ok := expr.(*ast.StringLiteral)
	return ok
}

func isNullLiteral(expr ast.Expression) bool {
	_, ok := expr.(*ast.NullLiteral)
	return ok
}
