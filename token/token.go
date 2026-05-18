// Package token defines token types used internally by the goatquery lexer and parser.
// Most consumers should use the goatquery root package or a module instead of importing
// this package directly.
package token

// TokenType represents the type of a lexer token.
type TokenType string

const (
	EOF     TokenType = "EOF"
	ILLEGAL TokenType = "ILLEGAL"

	// Identifiers and literals
	IDENT    TokenType = "IDENT"
	STRING   TokenType = "STRING"
	INT      TokenType = "INT"
	DOUBLE   TokenType = "DOUBLE"
	UUID     TokenType = "UUID"
	DATETIME TokenType = "DATETIME"
	DATE     TokenType = "DATE"
	NULL     TokenType = "NULL"
	BOOLEAN  TokenType = "BOOLEAN"

	// Delimiters
	LPAREN TokenType = "LPAREN"
	RPAREN TokenType = "RPAREN"
	SLASH  TokenType = "SLASH"
	COLON  TokenType = "COLON"
	COMMA  TokenType = "COMMA"
)

// Keywords used in filter and orderby expressions.
const (
	Asc  = "asc"
	Desc = "desc"

	Eq       = "eq"
	Ne       = "ne"
	Contains = "contains"

	Lt  = "lt"
	Lte = "lte"
	Gt  = "gt"
	Gte = "gte"

	And = "and"
	Or  = "or"

	Null  = "null"
	True  = "true"
	False = "false"

	Any = "any"
	All = "all"
)

// Token represents a lexer token with its type and literal value.
type Token struct {
	Type    TokenType
	Literal string
}

// NewToken creates a token from a single byte character.
func NewToken(tokenType TokenType, character byte) Token {
	return Token{Type: tokenType, Literal: string(character)}
}
