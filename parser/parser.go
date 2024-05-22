package parser

import (
	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/token"
)

type Parser struct {
	lexer *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token
}

func NewParser(lexer *lexer.Lexer) *Parser {
	p := &Parser{lexer: lexer}

	p.NextToken()
	p.NextToken()

	return p
}

func (p *Parser) NextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseOrderBy() []ast.OrderByStatement {
	statements := []ast.OrderByStatement{}

	for !p.currentTokenIs(token.EOF) {
		if !p.currentTokenIs(token.IDENT) {
			p.NextToken()
			continue
		}

		statement := &ast.OrderByStatement{Token: p.currentToken, Direction: ast.Ascending}

		if p.peekTokenIs(token.DESC) {
			statement.Direction = ast.Descending

			p.NextToken()
		}

		statements = append(statements, *statement)

		p.NextToken()
	}

	return statements
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}
