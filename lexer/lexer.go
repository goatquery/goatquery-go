package lexer

import (
	"strings"
	"time"

	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	character    byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readCharacter()

	return l
}

func (l *Lexer) readCharacter() {
	if l.readPosition >= len(l.input) {
		l.character = 0 // ASCII code for "NUL"
	} else {
		l.character = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.character {
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	case '(':
		tok = token.NewToken(token.LPAREN, l.character)
	case ')':
		tok = token.NewToken(token.RPAREN, l.character)
	case '\'':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	default:
		if isLetter(l.character) || isDigit(l.character) {
			tok.Literal = l.readIdentifier()

			if isGuid(tok.Literal) {
				tok.Type = token.GUID
				return tok
			}

			if isDigit(tok.Literal[0]) {

				if isDateTime(tok.Literal) {
					tok.Type = token.DATETIME
					return tok
				}

				if strings.HasSuffix(strings.ToLower(tok.Literal), "f") {
					tok.Type = token.FLOAT
					return tok
				}

				tok.Type = token.INT
				return tok
			}

			tok.Type = token.IDENT
			return tok
		}

		tok = token.NewToken(token.ILLEGAL, l.character)
	}

	l.readCharacter()

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.character == ' ' || l.character == '\t' || l.character == '\n' || l.character == '\r' {
		l.readCharacter()
	}
}

func (l *Lexer) readIdentifier() string {
	currentPosition := l.position
	for isLetter(l.character) || isDigit(l.character) || l.character == '-' || l.character == ':' || l.character == '.' {
		l.readCharacter()
	}

	return l.input[currentPosition:l.position]
}

func (l *Lexer) readString() string {
	currentPosition := l.position + 1

	for {
		l.readCharacter()
		if l.character == '\'' || l.character == 0 {
			break
		}
	}

	return l.input[currentPosition:l.position]
}

func isGuid(value string) bool {
	_, err := uuid.Parse(value)

	return err == nil
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isDateTime(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return true
	}

	_, err = time.Parse(time.DateOnly, value)
	if err == nil {
		return true
	}

	return false
}
