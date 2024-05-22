package lexer

import "github.com/goatquery/goatquery-go/token"

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
	default:
		if isLetter(l.character) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.character) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = token.NewToken(token.ILLEGAL, l.character)
		}
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
	for isLetter(l.character) {
		l.readCharacter()
	}

	return l.input[currentPosition:l.position]
}

func (l *Lexer) readNumber() string {
	currentPosition := l.position
	for isDigit(l.character) {
		l.readCharacter()
	}

	return l.input[currentPosition:l.position]
}

func (l *Lexer) readString() string {
	currentPosition := l.position + 1

	for {
		l.readCharacter()
		if l.character == '"' || l.character == 0 {
			break
		}
	}

	return l.input[currentPosition:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
