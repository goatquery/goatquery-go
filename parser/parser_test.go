package parser

import (
	"testing"

	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/token"
	"github.com/stretchr/testify/assert"
)

func Test_ParsingOrderByStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected []ast.OrderByStatement
	}{
		{"ID desc", []ast.OrderByStatement{{Token: token.Token{Type: token.IDENT, Literal: "ID"}, Direction: ast.Descending}}},
		{"id asc", []ast.OrderByStatement{{Token: token.Token{Type: token.IDENT, Literal: "id"}, Direction: ast.Ascending}}},
		{"Name", []ast.OrderByStatement{{Token: token.Token{Type: token.IDENT, Literal: "Name"}, Direction: ast.Ascending}}},
		{"id asc, name desc", []ast.OrderByStatement{
			{Token: token.Token{Type: token.IDENT, Literal: "id"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "name"}, Direction: ast.Descending},
		}},
		{"id asc, name desc, age, address asc, postcode desc", []ast.OrderByStatement{
			{Token: token.Token{Type: token.IDENT, Literal: "id"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "name"}, Direction: ast.Descending},
			{Token: token.Token{Type: token.IDENT, Literal: "age"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "address"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "postcode"}, Direction: ast.Descending},
		}},
		{"address1Line10 asc, asc asc, desc desc", []ast.OrderByStatement{
			{Token: token.Token{Type: token.IDENT, Literal: "address1Line10"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "asc"}, Direction: ast.Ascending},
			{Token: token.Token{Type: token.IDENT, Literal: "desc"}, Direction: ast.Descending},
		}},
		{"", []ast.OrderByStatement{}},
	}

	for _, test := range tests {
		l := lexer.NewLexer(test.input)
		p := NewParser(l)

		statements := p.ParseOrderBy()

		for i, expected := range test.expected {
			stmt := statements[i]

			assert.Equal(t, expected.TokenLiteral(), stmt.TokenLiteral())
			assert.Equal(t, expected.Direction, stmt.Direction)
		}
	}
}

func Test_ParsingFilterStatement(t *testing.T) {
	tests := []struct {
		input            string
		expectedLeft     string
		expectedOperator string
		expectedRight    string
	}{
		{"Name eq 'John'", "Name", "eq", "John"},
		{"Firstname eq 'Jane'", "Firstname", "eq", "Jane"},
		{"Age eq 21", "Age", "eq", "21"},
		{"Age ne 10", "Age", "ne", "10"},
		{"Name contains 'John'", "Name", "contains", "John"},
		{"Id eq e4c7772b-8947-4e46-98ed-644b417d2a08", "Id", "eq", "e4c7772b-8947-4e46-98ed-644b417d2a08"},
		{"Id eq 3.14159265359f", "Id", "eq", "3.14159265359f"},
		{"Age lt 99", "Age", "lt", "99"},
		{"Age lte 99", "Age", "lte", "99"},
		{"Age gt 99", "Age", "gt", "99"},
		{"Age gte 99", "Age", "gte", "99"},
		{"dateOfBirth eq 2000-01-01", "dateOfBirth", "eq", "2000-01-01"},
		{"dateOfBirth lt 2000-01-01", "dateOfBirth", "lt", "2000-01-01"},
		{"dateOfBirth lte 2000-01-01", "dateOfBirth", "lte", "2000-01-01"},
		{"dateOfBirth gt 2000-01-01", "dateOfBirth", "gt", "2000-01-01"},
		{"dateOfBirth gte 2000-01-01", "dateOfBirth", "gte", "2000-01-01"},
		{"dateOfBirth eq 2023-01-30T09:29:55.1750906Z", "dateOfBirth", "eq", "2023-01-30T09:29:55.1750906Z"},
	}

	for _, test := range tests {
		l := lexer.NewLexer(test.input)
		p := NewParser(l)

		statement := p.ParseFilter()

		expression := statement.Expression
		assert.NotNil(t, expression)

		assert.Equal(t, test.expectedLeft, expression.Left.TokenLiteral())
		assert.Equal(t, test.expectedOperator, expression.Operator)
		assert.Equal(t, test.expectedRight, expression.Right.TokenLiteral())
	}
}

func Test_ParsingInvalidFilterReturnsError(t *testing.T) {
	inputs := []string{
		"Name",
		"",
		"eq nee",
		"name nee 10",
		"id contains 10",
		"id contaiins '10'",
		"id eq       John'",
	}

	for _, input := range inputs {
		l := lexer.NewLexer(input)
		p := NewParser(l)

		statement := p.ParseFilter()

		expression := statement.Expression
		assert.Nil(t, expression)
	}
}

func Test_ParsingFilterStatementWithAnd(t *testing.T) {
	input := `Name eq 'John' and Age eq 10`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	statement := p.ParseFilter()

	expression := statement.Expression
	assert.NotNil(t, expression)

	//Left
	left, ok := expression.Left.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Name", left.Left.TokenLiteral())
	assert.Equal(t, "eq", left.Operator)
	assert.Equal(t, "John", left.Right.TokenLiteral())

	// inner operator
	assert.Equal(t, "and", expression.Operator)

	// inner right
	right, ok := expression.Right.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Age", right.Left.TokenLiteral())
	assert.Equal(t, "eq", right.Operator)
	assert.Equal(t, "10", right.Right.TokenLiteral())
}

func Test_ParsingFilterStatementWithOr(t *testing.T) {
	input := `Name eq 'John' or Age eq 10`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	statement := p.ParseFilter()

	expression := statement.Expression
	assert.NotNil(t, expression)

	//Left
	left, ok := expression.Left.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Name", left.Left.TokenLiteral())
	assert.Equal(t, "eq", left.Operator)
	assert.Equal(t, "John", left.Right.TokenLiteral())

	// inner operator
	assert.Equal(t, "or", expression.Operator)

	// inner right
	right, ok := expression.Right.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Age", right.Left.TokenLiteral())
	assert.Equal(t, "eq", right.Operator)
	assert.Equal(t, "10", right.Right.TokenLiteral())
}

func Test_ParsingFilterStatementWithAndAndOr(t *testing.T) {
	input := `Name eq 'John' and Age eq 10 or Id eq 10`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	statement := p.ParseFilter()

	expression := statement.Expression
	assert.NotNil(t, expression)

	//Left
	left, ok := expression.Left.(*ast.InfixExpression)
	assert.True(t, ok)

	// Inner left
	innerLeft, ok := left.Left.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Name", innerLeft.Left.TokenLiteral())
	assert.Equal(t, "eq", innerLeft.Operator)
	assert.Equal(t, "John", innerLeft.Right.TokenLiteral())

	// inner operator
	assert.Equal(t, "and", left.Operator)

	// inner right
	innerRight, ok := left.Right.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Age", innerRight.Left.TokenLiteral())
	assert.Equal(t, "eq", innerRight.Operator)
	assert.Equal(t, "10", innerRight.Right.TokenLiteral())

	// operator
	assert.Equal(t, "or", expression.Operator)

	// right
	right, ok := expression.Right.(*ast.InfixExpression)
	assert.True(t, ok)

	assert.Equal(t, "Id", right.Left.TokenLiteral())
	assert.Equal(t, "eq", right.Operator)
	assert.Equal(t, "10", right.Right.TokenLiteral())
}
