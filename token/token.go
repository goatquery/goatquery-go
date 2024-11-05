package token

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT    = "IDENT"
	INT      = "INT"
	STRING   = "STRING"
	GUID     = "GUID"
	FLOAT    = "FLOAT"
	DATETIME = "DATETIME"

	LPAREN = "LPAREN"
	RPAREN = "RPAREN"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

func NewToken(tokenType TokenType, character byte) Token {
	return Token{Type: tokenType, Literal: string(character)}
}
