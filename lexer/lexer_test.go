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
	asc asc
	address1Line asc
	addASCress1Line desc
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

		{token.IDENT, "asc"},
		{token.IDENT, "asc"},

		{token.IDENT, "address1Line"},
		{token.IDENT, "asc"},

		{token.IDENT, "addASCress1Line"},
		{token.IDENT, "desc"},
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
	Name eq 'john' or Id eq 1
	Id eq 1 and Name eq 'John' or Id eq 2
	Id eq 1 or Name eq 'John' or Id eq 2
	Id ne 1
	Name contains 'John'
	(Id eq 1 or Id eq 2) and Name eq 'John'
	address1Line eq '1 Main Street'
	addASCress1Line contains '10 Test Av'
	id eq e4c7772b-8947-4e46-98ed-644b417d2a08
	id eq 10f
	id ne 0.1121563052701180f
	id ne 0.1121563052701180F
	age lt 50
	age lte 50
	age gt 50
	age gte 50
	dateOfBirth eq 2000-01-01
	dateOfBirth lt 2000-01-01
	dateOfBirth lte 2000-01-01
	dateOfBirth gt 2000-01-01
	dateOfBirth gte 2000-01-01
	dateOfBirth eq 2023-01-01T15:30:00Z
	dateOfBirth eq 2023-01-30T09:29:55.1750906Z
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

		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "john"},
		{token.IDENT, "or"},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},

		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},
		{token.IDENT, "and"},
		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},
		{token.IDENT, "or"},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "2"},

		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},
		{token.IDENT, "or"},
		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},
		{token.IDENT, "or"},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "2"},

		{token.IDENT, "Id"},
		{token.IDENT, "ne"},
		{token.INT, "1"},

		{token.IDENT, "Name"},
		{token.IDENT, "contains"},
		{token.STRING, "John"},

		{token.LPAREN, "("},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "1"},
		{token.IDENT, "or"},
		{token.IDENT, "Id"},
		{token.IDENT, "eq"},
		{token.INT, "2"},
		{token.RPAREN, ")"},
		{token.IDENT, "and"},
		{token.IDENT, "Name"},
		{token.IDENT, "eq"},
		{token.STRING, "John"},

		{token.IDENT, "address1Line"},
		{token.IDENT, "eq"},
		{token.STRING, "1 Main Street"},

		{token.IDENT, "addASCress1Line"},
		{token.IDENT, "contains"},
		{token.STRING, "10 Test Av"},

		{token.IDENT, "id"},
		{token.IDENT, "eq"},
		{token.GUID, "e4c7772b-8947-4e46-98ed-644b417d2a08"},

		{token.IDENT, "id"},
		{token.IDENT, "eq"},
		{token.FLOAT, "10f"},

		{token.IDENT, "id"},
		{token.IDENT, "ne"},
		{token.FLOAT, "0.1121563052701180f"},

		{token.IDENT, "id"},
		{token.IDENT, "ne"},
		{token.FLOAT, "0.1121563052701180F"},

		{token.IDENT, "age"},
		{token.IDENT, "lt"},
		{token.INT, "50"},

		{token.IDENT, "age"},
		{token.IDENT, "lte"},
		{token.INT, "50"},

		{token.IDENT, "age"},
		{token.IDENT, "gt"},
		{token.INT, "50"},

		{token.IDENT, "age"},
		{token.IDENT, "gte"},
		{token.INT, "50"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "eq"},
		{token.DATETIME, "2000-01-01"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "lt"},
		{token.DATETIME, "2000-01-01"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "lte"},
		{token.DATETIME, "2000-01-01"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "gt"},
		{token.DATETIME, "2000-01-01"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "gte"},
		{token.DATETIME, "2000-01-01"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "eq"},
		{token.DATETIME, "2023-01-01T15:30:00Z"},

		{token.IDENT, "dateOfBirth"},
		{token.IDENT, "eq"},
		{token.DATETIME, "2023-01-30T09:29:55.1750906Z"},
	}

	lexer := NewLexer(input)

	for _, test := range tests {
		token := lexer.NextToken()

		assert.Equal(t, test.expectedType, token.Type)
		assert.Equal(t, test.expectedLiteral, token.Literal)
	}
}
