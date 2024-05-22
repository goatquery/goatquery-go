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
