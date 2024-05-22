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
