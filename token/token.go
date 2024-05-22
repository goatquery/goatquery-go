package token

import "strings"

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	// Keywords
	ASC  = "ASC"
	DESC = "DESC"
)

var keywords = map[string]TokenType{
	"asc":  ASC,
	"desc": DESC,
}

func LookupIdent(identifier string) TokenType {
	if tok, ok := keywords[strings.ToLower(identifier)]; ok {
		return tok
	}

	return IDENT
}

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

func NewToken(tokenType TokenType, character byte) Token {
	return Token{Type: tokenType, Literal: string(character)}
}
