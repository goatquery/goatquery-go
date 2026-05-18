// Package lexer tokenizes query expressions into tokens. Most consumers should use
// the goatquery root package or a module instead of importing this package directly.
package lexer

import (
	"strings"
	"time"

	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
)

// Lexer tokenizes a query expression string.
type Lexer struct {
	input        string
	position     int
	readPosition int
	character    byte
}

// NewLexer returns a Lexer for the given input.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readCharacter()
	return l
}

// NextToken returns the next token from the input.
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
	case '/':
		tok = token.NewToken(token.SLASH, l.character)
	case ':':
		tok = token.NewToken(token.COLON, l.character)
	case ',':
		tok = token.NewToken(token.COMMA, l.character)
	case '\'':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	default:
		if l.character == '-' && isDigit(l.peekCharacter()) {
			return l.readNegativeNumericToken()
		}

		if isDigit(l.character) {
			return l.readNumericOrDateToken()
		}

		if isLetter(l.character) {
			literal := l.readIdentifier()

			// Check if this identifier could be the start of a UUID (e.g., "e4c7772b-...")
			// UUIDs starting with hex letters: all chars are hex and next char is '-'
			if l.character == '-' && isAllHex(literal) {
				if guid := l.tryReadUUID(literal); guid != "" {
					tok.Literal = guid
					tok.Type = token.UUID
					return tok
				}
			}

			tok.Literal = literal
			tok.Type = classifyIdentifier(literal)
			return tok
		}

		tok = token.NewToken(token.ILLEGAL, l.character)
	}

	l.readCharacter()
	return tok
}

func (l *Lexer) readCharacter() {
	if l.readPosition >= len(l.input) {
		l.character = 0
	} else {
		l.character = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekCharacter() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.character == ' ' || l.character == '\t' || l.character == '\n' || l.character == '\r' {
		l.readCharacter()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.character) || isDigit(l.character) {
		l.readCharacter()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	l.readCharacter()

	var b strings.Builder
	for l.character != 0 {
		if l.character == '\\' {
			next := l.peekCharacter()
			if next == '\'' || next == '\\' {
				l.readCharacter() // skip backslash
				b.WriteByte(l.character)
				l.readCharacter() // advance past escaped char
				continue
			}
		}
		if l.character == '\'' {
			l.readCharacter() // skip closing quote
			break
		}
		b.WriteByte(l.character)
		l.readCharacter()
	}
	return b.String()
}

func (l *Lexer) readNumericOrDateToken() token.Token {
	start := l.position

	for isNumericComponent(l.character) {
		l.readCharacter()
	}

	literal := l.input[start:l.position]
	tokenType := classifyNumericLiteral(literal)

	return token.Token{Type: tokenType, Literal: literal}
}

func (l *Lexer) readNegativeNumericToken() token.Token {
	start := l.position
	l.readCharacter()

	hasDot := false
	for isDigit(l.character) || l.character == '.' {
		if l.character == '.' {
			if hasDot {
				break
			}
			hasDot = true
		}
		l.readCharacter()
	}

	literal := l.input[start:l.position]
	tokenType := token.INT
	if strings.Contains(literal, ".") {
		tokenType = token.DOUBLE
	}

	return token.Token{Type: tokenType, Literal: literal}
}

func classifyNumericLiteral(literal string) token.TokenType {
	if _, err := uuid.Parse(literal); err == nil {
		return token.UUID
	}

	if len(literal) == 10 {
		if _, err := time.Parse(time.DateOnly, literal); err == nil {
			return token.DATE
		}
	}

	if strings.ContainsAny(literal, "Tt") {
		if _, err := time.Parse(time.RFC3339, literal); err == nil {
			return token.DATETIME
		}
		if _, err := time.Parse(time.RFC3339Nano, literal); err == nil {
			return token.DATETIME
		}
	}

	if strings.Contains(literal, ".") {
		return token.DOUBLE
	}

	return token.INT
}

func classifyIdentifier(literal string) token.TokenType {
	lower := strings.ToLower(literal)

	if lower == token.Null {
		return token.NULL
	}

	if lower == token.True || lower == token.False {
		return token.BOOLEAN
	}

	return token.IDENT
}

func isNumericComponent(ch byte) bool {
	return isDigit(ch) || isHexLetter(ch) || ch == '.' || ch == '-' || ch == ':' || ch == 'T' || ch == 'Z' || ch == '+'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func isHexChar(ch byte) bool {
	return isDigit(ch) || isHexLetter(ch)
}

func isAllHex(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isHexChar(s[i]) {
			return false
		}
	}
	return len(s) > 0
}

func (l *Lexer) tryReadUUID(prefix string) string {
	if len(prefix) != 8 {
		return ""
	}

	remaining := l.input[l.position:]
	if len(remaining) < 28 {
		return ""
	}

	candidate := prefix + remaining[:28]
	if _, err := uuid.Parse(candidate); err != nil {
		return ""
	}

	for i := 0; i < 28; i++ {
		l.readCharacter()
	}

	return candidate
}
