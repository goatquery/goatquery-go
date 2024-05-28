package parser

import (
	"testing"

	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/stretchr/testify/assert"
)

func Test_ParsingOrderByStatement(t *testing.T) {
	tests := []struct {
		input             string
		expectedLiteral   string
		expectedDirection ast.OrderByDirection
	}{
		{"id", "id", ast.Ascending},
		{"id asc", "id", ast.Ascending},
		{"id desc", "id", ast.Descending},
	}

	for _, test := range tests {
		l := lexer.NewLexer(test.input)
		p := NewParser(l)

		statements := p.ParseOrderBy()
		assert.Len(t, statements, 1)

		stmt := statements[0]

		assert.Equal(t, test.expectedLiteral, stmt.TokenLiteral())
		assert.Equal(t, test.expectedDirection, stmt.Direction)
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

func Test_ParsingComplexFilterStatement(t *testing.T) {
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
