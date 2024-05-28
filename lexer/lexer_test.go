package lexer

import (
	"testing"

	"github.com/goatquery/goatquery-go/token"
	"github.com/stretchr/testify/assert"
)

func Test_OrderByNextToken(t *testing.T) {
	input := `id asc
	iD desc
	id aSc
	id DeSc
	id AsC
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "id"},
		{token.IDENT, "asc"},

		{token.IDENT, "iD"},
		{token.IDENT, "desc"},

		{token.IDENT, "id"},
		{token.IDENT, "aSc"},

		{token.IDENT, "id"},
		{token.IDENT, "DeSc"},

		{token.IDENT, "id"},
		{token.IDENT, "AsC"},
	}

	lexer := NewLexer(input)

	for _, test := range tests {
		token := lexer.NextToken()

		assert.Equal(t, test.expectedType, token.Type)
		assert.Equal(t, test.expectedLiteral, token.Literal)
	}
}

func Test_FilterNextToken(t *testing.T) {
	input := `Name eq 'John'
	Id eq 1
	Name eq 'John' and Id eq 1
	eq eq 'John'
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},

		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},

		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},
		{token.IDENT, "and"},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},

		{token.IDENT, "eq"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},
	}

	lexer := NewLexer(input)

	for _, test := range tests {
		token := lexer.NextToken()

		assert.Equal(t, test.expectedType, token.Type)
		assert.Equal(t, test.expectedLiteral, token.Literal)
	}
}
