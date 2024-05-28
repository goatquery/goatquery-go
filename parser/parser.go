package parser

import (
	"strconv"
	"strings"

	"github.com/goatquery/goatquery-go/ast"
	keyword "github.com/goatquery/goatquery-go/keywords"
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

		if p.peekIdentiferIs(keyword.DESC) {
			statement.Direction = ast.Descending
		}

		p.NextToken()

		statements = append(statements, *statement)

		p.NextToken()
	}

	return statements
}

func (p *Parser) ParseFilter() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}
	statement.Expression = p.ParseExpression()

	return statement
}

func (p *Parser) ParseExpression() ast.InfixExpression {
	left := p.ParseFilterStatement()

	p.NextToken()

	for p.currentIdentiferIs(keyword.AND) || p.currentIdentiferIs(keyword.OR) {
		left = &ast.InfixExpression{Token: p.currentToken, Left: left, Operator: p.currentToken.Literal}

		p.NextToken()

		right := p.ParseFilterStatement()
		left.Right = right

		p.NextToken()
	}

	return *left
}

func (p *Parser) ParseFilterStatement() *ast.InfixExpression {
	identifer := ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.peekIdentiferIs(keyword.EQ) {
		return nil
	}

	p.NextToken()

	statement := ast.InfixExpression{Token: p.currentToken, Left: &identifer, Operator: p.currentToken.Literal}

	if !p.peekTokenIs(token.STRING) && !p.peekTokenIs(token.INT) {
		return nil
	}

	p.NextToken()

	switch p.currentToken.Type {
	case token.STRING:
		statement.Right = &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
	case token.INT:
		literal := &ast.IntegerLiteral{Token: p.currentToken}

		value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
		if err != nil {
			return nil
		}

		literal.Value = value

		statement.Right = literal
	}

	return &statement
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekIdentiferIs(identifier string) bool {
	return p.peekToken.Type == token.IDENT && strings.EqualFold(p.peekToken.Literal, identifier)
}

func (p *Parser) currentIdentiferIs(identifier string) bool {
	return p.currentToken.Type == token.IDENT && strings.EqualFold(p.currentToken.Literal, identifier)
}
