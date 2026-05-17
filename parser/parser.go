// Package parser produces an AST from the token stream. Most consumers should use
// the goatquery root package or a module instead of importing this package directly.
package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goatquery/goatquery-go/ast"
	"github.com/goatquery/goatquery-go/lexer"
	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
)

const (
	precedenceLowest = 0
	precedenceOr     = 1
	precedenceAnd    = 2
)

// Parser produces an AST from a sequence of tokens.
type Parser struct {
	lexer *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token
}

// NewParser returns a Parser that reads from the given Lexer.
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{lexer: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// ParseOrderBy parses an orderby clause into a list of statements.
func (p *Parser) ParseOrderBy() ([]ast.OrderByStatement, error) {
	statements := []ast.OrderByStatement{}

	for !p.currentTokenIs(token.EOF) {
		if !p.currentTokenIs(token.IDENT) {
			return nil, fmt.Errorf("Expected property name in orderby, got '%s'.", p.currentToken.Literal)
		}

		statement := ast.OrderByStatement{
			Token:     p.currentToken,
			Segments:  []string{p.currentToken.Literal},
			Direction: ast.Ascending,
		}

		for p.peekTokenIs(token.SLASH) {
			p.nextToken()
			p.nextToken()
			if !p.currentTokenIs(token.IDENT) {
				return nil, fmt.Errorf("Expected identifier after '/' in orderby path.")
			}
			statement.Segments = append(statement.Segments, p.currentToken.Literal)
		}

		// Check for direction
		if p.peekIdentifierIs(token.Desc) {
			statement.Direction = ast.Descending
			p.nextToken()
		} else if p.peekIdentifierIs(token.Asc) {
			p.nextToken()
		}

		statements = append(statements, statement)
		p.nextToken()

		if p.currentTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	return statements, nil
}

// ParseFilter parses a filter expression.
func (p *Parser) ParseFilter() (ast.Expression, error) {
	if p.currentTokenIs(token.EOF) {
		return nil, fmt.Errorf("Empty filter expression.")
	}

	return p.parseExpression(precedenceLowest)
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	var left ast.Expression
	var err error

	if p.currentTokenIs(token.LPAREN) {
		left, err = p.parseGroupedExpression()
	} else {
		left, err = p.parseFilterStatement()
	}
	if err != nil {
		return nil, err
	}

	p.nextToken()

	for !p.currentTokenIs(token.EOF) && precedence < p.getPrecedence() {
		if !p.currentIdentifierIs(token.And) && !p.currentIdentifierIs(token.Or) {
			break
		}

		operator := p.currentToken
		currentPrecedence := p.getPrecedence()
		p.nextToken()

		right, err := p.parseExpression(currentPrecedence)
		if err != nil {
			return nil, err
		}

		left = &ast.InfixExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, err
	}

	if !p.currentTokenIs(token.RPAREN) {
		return nil, fmt.Errorf("Expected ')' to close grouped expression.")
	}

	return exp, nil
}

func (p *Parser) parseFilterStatement() (ast.Expression, error) {
	if !p.currentTokenIs(token.IDENT) {
		return nil, fmt.Errorf("Expected identifier, got '%s'.", p.currentToken.Type)
	}

	firstToken := p.currentToken
	segments := []string{p.currentToken.Literal}

	for p.peekTokenIs(token.SLASH) {
		p.nextToken()
		p.nextToken()

		if !p.currentTokenIs(token.IDENT) {
			return nil, fmt.Errorf("Expected identifier after '/' in property path.")
		}

		segments = append(segments, p.currentToken.Literal)

		lower := strings.ToLower(p.currentToken.Literal)
		if (lower == token.Any || lower == token.All) && p.peekTokenIs(token.LPAREN) {
			p.nextToken()

			collectionSegments := segments[:len(segments)-1]
			var prop ast.Expression
			if len(collectionSegments) == 1 {
				prop = &ast.Identifier{Token: firstToken}
			} else {
				prop = &ast.PropertyPath{Token: firstToken, Segments: collectionSegments}
			}

			return p.parseLambdaExpression(prop, lower)
		}
	}

	var property ast.Expression
	if len(segments) == 1 {
		property = &ast.Identifier{Token: firstToken}
	} else {
		property = &ast.PropertyPath{Token: firstToken, Segments: segments}
	}

	if !p.peekIdentifierIn(token.Eq, token.Ne, token.Contains, token.Lt, token.Lte, token.Gt, token.Gte) {
		return nil, fmt.Errorf("Expected comparison operator, got '%s'.", p.peekToken.Literal)
	}

	p.nextToken()
	operator := p.currentToken

	if strings.EqualFold(operator.Literal, token.Contains) {
		if !p.peekTokenIs(token.STRING) {
			return nil, fmt.Errorf("Operator 'contains' requires a string value.")
		}
	}

	opLower := strings.ToLower(operator.Literal)
	if opLower == token.Lt || opLower == token.Lte || opLower == token.Gt || opLower == token.Gte {
		if p.peekTokenIs(token.NULL) {
			return nil, fmt.Errorf("Operator '%s' does not support null values.", operator.Literal)
		}
		if !p.peekTokenIn(token.INT, token.DOUBLE, token.DATETIME, token.DATE) {
			return nil, fmt.Errorf("Value must be a numeric or date type when using '%s' operator.", operator.Literal)
		}
	}

	if !p.peekTokenIn(token.STRING, token.INT, token.DOUBLE, token.UUID, token.DATETIME, token.DATE, token.NULL, token.BOOLEAN) {
		return nil, fmt.Errorf("Expected value after operator '%s', got '%s'.", operator.Literal, p.peekToken.Literal)
	}

	p.nextToken()

	right, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}

	return &ast.InfixExpression{
		Token:    operator,
		Left:     property,
		Operator: operator.Literal,
		Right:    right,
	}, nil
}

func (p *Parser) parseLambdaExpression(collectionProp ast.Expression, functionName string) (ast.Expression, error) {
	lambdaToken := p.currentToken
	p.nextToken()

	if !p.currentTokenIs(token.IDENT) {
		return nil, fmt.Errorf("Expected lambda parameter name, got '%s'.", p.currentToken.Type)
	}

	paramName := p.currentToken.Literal

	p.nextToken()
	if !p.currentTokenIs(token.COLON) {
		return nil, fmt.Errorf("Expected ':' after lambda parameter '%s'.", paramName)
	}

	p.nextToken()

	body, err := p.parseExpression(precedenceLowest)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse lambda body: %w", err)
	}

	if !p.currentTokenIs(token.RPAREN) {
		return nil, fmt.Errorf("Expected ')' to close lambda expression.")
	}

	return &ast.LambdaExpression{
		Token:     lambdaToken,
		Property:  collectionProp,
		Function:  functionName,
		Parameter: paramName,
		Body:      body,
	}, nil
}

func (p *Parser) parseLiteral() (ast.Expression, error) {
	switch p.currentToken.Type {
	case token.STRING:
		return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}, nil

	case token.INT:
		value, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid number.", p.currentToken.Literal)
		}
		return &ast.IntegerLiteral{Token: p.currentToken, Value: value}, nil

	case token.DOUBLE:
		value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid number.", p.currentToken.Literal)
		}
		return &ast.FloatLiteral{Token: p.currentToken, Value: value}, nil

	case token.UUID:
		value, err := uuid.Parse(p.currentToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("Could not parse UUID '%s': %w.", p.currentToken.Literal, err)
		}
		return &ast.UUIDLiteral{Token: p.currentToken, Value: value}, nil

	case token.DATETIME:
		value, err := parseDateTime(p.currentToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("Could not parse datetime '%s': %w.", p.currentToken.Literal, err)
		}
		return &ast.DateTimeLiteral{Token: p.currentToken, Value: value}, nil

	case token.DATE:
		value, err := time.Parse(time.DateOnly, p.currentToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("Could not parse date '%s': %w.", p.currentToken.Literal, err)
		}
		return &ast.DateLiteral{Token: p.currentToken, Value: value}, nil

	case token.NULL:
		return &ast.NullLiteral{Token: p.currentToken}, nil

	case token.BOOLEAN:
		value := strings.EqualFold(p.currentToken.Literal, token.True)
		return &ast.BooleanLiteral{Token: p.currentToken, Value: value}, nil

	default:
		return nil, fmt.Errorf("Unexpected token type '%s' for literal value.", p.currentToken.Type)
	}
}

func parseDateTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("Could not parse datetime '%s'.", value)
}

func (p *Parser) getPrecedence() int {
	if p.currentTokenIs(token.IDENT) {
		if p.currentIdentifierIs(token.And) {
			return precedenceAnd
		}
		if p.currentIdentifierIs(token.Or) {
			return precedenceOr
		}
	}
	return precedenceLowest
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekTokenIn(types ...token.TokenType) bool {
	for _, t := range types {
		if p.peekToken.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) peekIdentifierIs(identifier string) bool {
	return p.peekToken.Type == token.IDENT && strings.EqualFold(p.peekToken.Literal, identifier)
}

func (p *Parser) peekIdentifierIn(identifiers ...string) bool {
	if p.peekToken.Type != token.IDENT {
		return false
	}
	for _, id := range identifiers {
		if strings.EqualFold(p.peekToken.Literal, id) {
			return true
		}
	}
	return false
}

func (p *Parser) currentIdentifierIs(identifier string) bool {
	return p.currentToken.Type == token.IDENT && strings.EqualFold(p.currentToken.Literal, identifier)
}
