package parser

import (
	"strconv"
	"strings"

	"github.com/goatquery/goatquery-go"
	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/keywords"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
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

		if p.peekIdentiferIs(keywords.DESC) {
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
	statement.Expression = p.ParseExpression(0)

	return statement
}

func (p *Parser) ParseExpression(precedence int) *ast.InfixExpression {
	var left *ast.InfixExpression

	if p.currentTokenIs(token.LPAREN) {
		left = p.ParseGroupedExpression()
	} else {
		left = p.ParseFilterStatement()
	}

	p.NextToken()

	for !p.currentTokenIs(token.EOF) && precedence < p.GetPrecedence(p.currentToken.Type) {
		if p.currentIdentiferIs(keywords.AND) || p.currentIdentiferIs(keywords.OR) {
			left = &ast.InfixExpression{Token: p.currentToken, Left: left, Operator: p.currentToken.Literal}
			currentPrecedence := p.GetPrecedence(p.currentToken.Type)

			p.NextToken()

			right := p.ParseExpression(currentPrecedence)
			left.Right = right
		} else {
			break
		}
	}

	return left
}

func (p *Parser) ParseGroupedExpression() *ast.InfixExpression {
	p.NextToken()

	exp := p.ParseExpression(0)

	if !p.currentTokenIs(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) ParseFilterStatement() *ast.InfixExpression {
	identifer := ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.peekIdentiferIn(keywords.EQ, keywords.NE, keywords.CONTAINS, keywords.LT, keywords.LTE, keywords.GT, keywords.GTE) {
		return nil
	}

	p.NextToken()

	statement := ast.InfixExpression{Token: p.currentToken, Left: &identifer, Operator: p.currentToken.Literal}

	if !p.peekTokenIn(token.STRING, token.INT, token.GUID, token.DATETIME, token.FLOAT) {
		return nil
	}

	p.NextToken()

	if strings.EqualFold(statement.Operator, keywords.CONTAINS) && p.currentToken.Type != token.STRING {
		return nil
	}

	switch p.currentToken.Type {
	case token.GUID:
		val, err := uuid.Parse(p.currentToken.Literal)
		if err == nil {
			statement.Right = &ast.GuidLiteral{Token: p.currentToken, Value: val}
		}
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
	case token.FLOAT:
		literal := &ast.FloatLiteral{Token: p.currentToken}

		literalWithoutSuffix := strings.TrimSuffix(p.currentToken.Literal, "f")
		value, err := strconv.ParseFloat(literalWithoutSuffix, 64)
		if err != nil {
			return nil
		}

		literal.Value = value

		statement.Right = literal
	case token.DATETIME:
		literal := &ast.DateTimeLiteral{Token: p.currentToken}

		value, err := goatquery.ParseDateTime(p.currentToken.Literal)
		if err != nil {
			return nil
		}

		literal.Value = *value

		statement.Right = literal
	}

	return &statement
}

func (p *Parser) GetPrecedence(t token.TokenType) int {
	switch t {
	case token.IDENT:
		if p.currentIdentiferIs(keywords.AND) {
			return 2
		}

		if p.currentIdentiferIs(keywords.OR) {
			return 1
		}
	}

	return 0
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekTokenIn(tokens ...token.TokenType) bool {
	for _, token := range tokens {
		if p.peekToken.Type == token {
			return true
		}
	}

	return false
}

func (p *Parser) peekIdentiferIs(identifier string) bool {
	return p.peekToken.Type == token.IDENT && strings.EqualFold(p.peekToken.Literal, identifier)
}

func (p *Parser) peekIdentiferIn(identifiers ...string) bool {
	if p.peekToken.Type != token.IDENT {
		return false
	}

	for _, identifier := range identifiers {
		if strings.EqualFold(p.peekToken.Literal, identifier) {
			return true
		}
	}

	return false
}

func (p *Parser) currentIdentiferIs(identifier string) bool {
	return p.currentToken.Type == token.IDENT && strings.EqualFold(p.currentToken.Literal, identifier)
}
