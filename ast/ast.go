package ast

import (
	"github.com/goatquery/goatquery-go/token"
)

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type OrderByDirection string

const (
	Ascending  OrderByDirection = "asc"
	Descending OrderByDirection = "desc"
)

type OrderByStatement struct {
	Token     token.Token
	Direction OrderByDirection
}

var _ Statement = (*OrderByStatement)(nil)

func (s *OrderByStatement) statementNode()       {}
func (s *OrderByStatement) TokenLiteral() string { return s.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

var _ Expression = (*Identifier)(nil)

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type ExpressionStatement struct {
	Token      token.Token
	Expression InfixExpression
}

var _ Statement = (*ExpressionStatement)(nil)

func (s *ExpressionStatement) statementNode()       {}
func (s *ExpressionStatement) TokenLiteral() string { return s.Token.Literal }

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

var _ Expression = (*InfixExpression)(nil)

func (s *InfixExpression) expressionNode()      {}
func (s *InfixExpression) TokenLiteral() string { return s.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

var _ Expression = (*StringLiteral)(nil)

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

var _ Expression = (*IntegerLiteral)(nil)

func (s *IntegerLiteral) expressionNode()      {}
func (s *IntegerLiteral) TokenLiteral() string { return s.Token.Literal }
