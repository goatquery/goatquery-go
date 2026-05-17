package lexer

import (
	"testing"

	"github.com/goatquery/goatquery-go/token"
	"github.com/stretchr/testify/assert"
)

func Test_OrderByNextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ     token.TokenType
			literal string
		}
	}{
		{
			name:  "simple ascending",
			input: "id asc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "asc"},
			},
		},
		{
			name:  "simple descending",
			input: "iD desc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "iD"},
				{token.IDENT, "desc"},
			},
		},
		{
			name:  "mixed case asc",
			input: "id aSc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "aSc"},
			},
		},
		{
			name:  "mixed case desc",
			input: "id DeSc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "DeSc"},
			},
		},
		{
			name:  "property named asc",
			input: "asc asc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "asc"},
				{token.IDENT, "asc"},
			},
		},
		{
			name:  "property with digits",
			input: "address1Line asc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "address1Line"},
				{token.IDENT, "asc"},
			},
		},
		{
			name:  "property with uppercase substring",
			input: "addASCress1Line desc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "addASCress1Line"},
				{token.IDENT, "desc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, exp := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, exp.typ, tok.Type, "token %d type", i)
				assert.Equal(t, exp.literal, tok.Literal, "token %d literal", i)
			}
		})
	}
}

func Test_FilterNextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ     token.TokenType
			literal string
		}
	}{
		{
			name:  "string equality",
			input: "Name eq 'John'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Name"},
				{token.IDENT, "eq"},
				{token.STRING, "John"},
			},
		},
		{
			name:  "integer equality",
			input: "Id eq 1",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Id"},
				{token.IDENT, "eq"},
				{token.INT, "1"},
			},
		},
		{
			name:  "and operator",
			input: "Name eq 'John' and Id eq 1",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Name"},
				{token.IDENT, "eq"},
				{token.STRING, "John"},
				{token.IDENT, "and"},
				{token.IDENT, "Id"},
				{token.IDENT, "eq"},
				{token.INT, "1"},
			},
		},
		{
			name:  "property named eq",
			input: "eq eq 'John'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "eq"},
				{token.IDENT, "eq"},
				{token.STRING, "John"},
			},
		},
		{
			name:  "or operator",
			input: "Name eq 'john' or Id eq 1",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Name"},
				{token.IDENT, "eq"},
				{token.STRING, "john"},
				{token.IDENT, "or"},
				{token.IDENT, "Id"},
				{token.IDENT, "eq"},
				{token.INT, "1"},
			},
		},
		{
			name:  "and or combined",
			input: "Id eq 1 and Name eq 'John' or Id eq 2",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
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
			},
		},
		{
			name:  "multiple or",
			input: "Id eq 1 or Name eq 'John' or Id eq 2",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
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
			},
		},
		{
			name:  "ne operator",
			input: "Id ne 1",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Id"},
				{token.IDENT, "ne"},
				{token.INT, "1"},
			},
		},
		{
			name:  "contains operator",
			input: "Name contains 'John'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "Name"},
				{token.IDENT, "contains"},
				{token.STRING, "John"},
			},
		},
		{
			name:  "parentheses grouping",
			input: "(Id eq 1 or Id eq 2) and Name eq 'John'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
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
			},
		},
		{
			name:  "string with spaces",
			input: "address1Line eq '1 Main Street'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "address1Line"},
				{token.IDENT, "eq"},
				{token.STRING, "1 Main Street"},
			},
		},
		{
			name:  "property with uppercase substring",
			input: "addASCress1Line contains '10 Test Av'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "addASCress1Line"},
				{token.IDENT, "contains"},
				{token.STRING, "10 Test Av"},
			},
		},
		{
			name:  "UUID literal",
			input: "id eq e4c7772b-8947-4e46-98ed-644b417d2a08",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "eq"},
				{token.UUID, "e4c7772b-8947-4e46-98ed-644b417d2a08"},
			},
		},
		{
			name:  "multi-digit integer",
			input: "id eq 10",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "eq"},
				{token.INT, "10"},
			},
		},
		{
			name:  "long double",
			input: "id ne 0.1121563052701180",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "ne"},
				{token.DOUBLE, "0.1121563052701180"},
			},
		},
		{
			name:  "short double",
			input: "id ne 3.14",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "ne"},
				{token.DOUBLE, "3.14"},
			},
		},
		{
			name:  "lt operator",
			input: "age lt 50",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "lt"},
				{token.INT, "50"},
			},
		},
		{
			name:  "lte operator",
			input: "age lte 50",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "lte"},
				{token.INT, "50"},
			},
		},
		{
			name:  "gt operator",
			input: "age gt 50",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "gt"},
				{token.INT, "50"},
			},
		},
		{
			name:  "gte operator",
			input: "age gte 50",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "gte"},
				{token.INT, "50"},
			},
		},
		{
			name:  "date eq",
			input: "dateOfBirth eq 2000-01-01",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "eq"},
				{token.DATE, "2000-01-01"},
			},
		},
		{
			name:  "date lt",
			input: "dateOfBirth lt 2000-01-01",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "lt"},
				{token.DATE, "2000-01-01"},
			},
		},
		{
			name:  "date lte",
			input: "dateOfBirth lte 2000-01-01",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "lte"},
				{token.DATE, "2000-01-01"},
			},
		},
		{
			name:  "date gt",
			input: "dateOfBirth gt 2000-01-01",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "gt"},
				{token.DATE, "2000-01-01"},
			},
		},
		{
			name:  "date gte",
			input: "dateOfBirth gte 2000-01-01",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "gte"},
				{token.DATE, "2000-01-01"},
			},
		},
		{
			name:  "datetime UTC",
			input: "dateOfBirth eq 2023-01-01T15:30:00Z",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "eq"},
				{token.DATETIME, "2023-01-01T15:30:00Z"},
			},
		},
		{
			name:  "datetime with fractional seconds",
			input: "dateOfBirth eq 2023-01-30T09:29:55.1750906Z",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "eq"},
				{token.DATETIME, "2023-01-30T09:29:55.1750906Z"},
			},
		},
		{
			name:  "null literal",
			input: "id eq null",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "eq"},
				{token.NULL, "null"},
			},
		},
		{
			name:  "boolean true",
			input: "isActive eq true",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "isActive"},
				{token.IDENT, "eq"},
				{token.BOOLEAN, "true"},
			},
		},
		{
			name:  "boolean false",
			input: "isActive eq false",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "isActive"},
				{token.IDENT, "eq"},
				{token.BOOLEAN, "false"},
			},
		},
		{
			name:  "nested property path",
			input: "company/name eq 'test'",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "company"},
				{token.SLASH, "/"},
				{token.IDENT, "name"},
				{token.IDENT, "eq"},
				{token.STRING, "test"},
			},
		},
		{
			name:  "lambda expression",
			input: "tags/any(t: t/name eq 'test')",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "tags"},
				{token.SLASH, "/"},
				{token.IDENT, "any"},
				{token.LPAREN, "("},
				{token.IDENT, "t"},
				{token.COLON, ":"},
				{token.IDENT, "t"},
				{token.SLASH, "/"},
				{token.IDENT, "name"},
				{token.IDENT, "eq"},
				{token.STRING, "test"},
				{token.RPAREN, ")"},
			},
		},
		{
			name:  "negative integer",
			input: "age gt -5",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "gt"},
				{token.INT, "-5"},
			},
		},
		{
			name:  "negative double",
			input: "age eq -0.5",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "eq"},
				{token.DOUBLE, "-0.5"},
			},
		},
		{
			name:  "negative large integer",
			input: "age gt -2147483648",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "gt"},
				{token.INT, "-2147483648"},
			},
		},
		{
			name:  "negative zero",
			input: "age eq -0",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "eq"},
				{token.INT, "-0"},
			},
		},
		{
			name:  "bare minus is illegal",
			input: "age gt -",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "age"},
				{token.IDENT, "gt"},
				{token.ILLEGAL, "-"},
			},
		},
		{
			name:  "uppercase NULL",
			input: "balance ne NULL",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "balance"},
				{token.IDENT, "ne"},
				{token.NULL, "NULL"},
			},
		},
		{
			name:  "datetime with positive timezone offset",
			input: "dateOfBirth eq 2024-01-15T10:30:00+05:00",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "eq"},
				{token.DATETIME, "2024-01-15T10:30:00+05:00"},
			},
		},
		{
			name:  "datetime with negative timezone offset",
			input: "dateOfBirth eq 2024-01-15T10:30:00-08:00",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "dateOfBirth"},
				{token.IDENT, "eq"},
				{token.DATETIME, "2024-01-15T10:30:00-08:00"},
			},
		},
		{
			name:  "UUID starting with digit",
			input: "id eq 12345678-1234-1234-1234-123456789abc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "eq"},
				{token.UUID, "12345678-1234-1234-1234-123456789abc"},
			},
		},
		{
			name:  "UUID with all-hex-letter prefix",
			input: "id eq abcdefab-abcd-abcd-abcd-abcdefabcdef",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "eq"},
				{token.UUID, "abcdefab-abcd-abcd-abcd-abcdefabcdef"},
			},
		},
		{
			name:  "comma in orderby",
			input: "id asc, name desc",
			expected: []struct {
				typ     token.TokenType
				literal string
			}{
				{token.IDENT, "id"},
				{token.IDENT, "asc"},
				{token.COMMA, ","},
				{token.IDENT, "name"},
				{token.IDENT, "desc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, exp := range tt.expected {
				tok := lexer.NextToken()
				assert.Equal(t, exp.typ, tok.Type, "token %d type", i)
				assert.Equal(t, exp.literal, tok.Literal, "token %d literal", i)
			}
		})
	}
}

func Test_BackslashEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", `'hello'`, "hello"},
		{"escaped single quote", `'it\'s'`, "it's"},
		{"escaped backslash", `'back\\slash'`, "back\\slash"},
		{"escaped backslash and quote", `'escaped\\\'quote'`, "escaped\\'quote"},
		{"empty string", `''`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tok := lexer.NextToken()

			assert.Equal(t, token.STRING, tok.Type)
			assert.Equal(t, tt.expected, tok.Literal)
		})
	}
}
