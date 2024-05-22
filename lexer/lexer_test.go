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
		{token.ASC, "asc"},

		{token.IDENT, "iD"},
		{token.DESC, "desc"},

		{token.IDENT, "id"},
		{token.ASC, "aSc"},

		{token.IDENT, "id"},
		{token.DESC, "DeSc"},

		{token.IDENT, "id"},
		{token.ASC, "AsC"},
	}

	lexer := NewLexer(input)

	for _, test := range tests {
		token := lexer.NextToken()

		assert.Equal(t, test.expectedType, token.Type)
		assert.Equal(t, test.expectedLiteral, token.Literal)
	}
}
