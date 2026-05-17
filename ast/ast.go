// Package ast defines the abstract syntax tree nodes used by goatquery. Most consumers
// should use the goatquery root package or a module instead of importing this package directly.
package ast

import (
	"time"

	"github.com/goatquery/goatquery-go/token"
	"github.com/google/uuid"
)

// Node is the interface implemented by all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Statement is a node that represents a statement.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that represents an expression.
type Expression interface {
	Node
	expressionNode()
}

// OrderByDirection represents the sort direction.
type OrderByDirection string

const (
	Ascending  OrderByDirection = "asc"
	Descending OrderByDirection = "desc"
)

// OrderByStatement represents a single orderby clause with optional nested path segments.
type OrderByStatement struct {
	Token     token.Token
	Segments  []string
	Direction OrderByDirection
}

var _ Statement = (*OrderByStatement)(nil)

func (s *OrderByStatement) statementNode()       {}
func (s *OrderByStatement) TokenLiteral() string { return s.Token.Literal }

// Identifier represents a single identifier (property name, keyword, etc.).
type Identifier struct {
	Token token.Token
}

var _ Expression = (*Identifier)(nil)

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// PropertyPath represents a slash-separated property path (e.g., "company/name").
type PropertyPath struct {
	Token    token.Token
	Segments []string
}

var _ Expression = (*PropertyPath)(nil)

func (p *PropertyPath) expressionNode()      {}
func (p *PropertyPath) TokenLiteral() string { return p.Token.Literal }

// InfixExpression represents a binary expression (e.g., "name eq 'John'" or "a and b").
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

var _ Expression = (*InfixExpression)(nil)

func (e *InfixExpression) expressionNode()      {}
func (e *InfixExpression) TokenLiteral() string { return e.Token.Literal }

// LambdaExpression represents a lambda filter (e.g., "tags/any(t: t/name eq 'test')").
type LambdaExpression struct {
	Token     token.Token
	Property  Expression // The collection property (Identifier or PropertyPath)
	Function  string     // "any" or "all"
	Parameter string     // Lambda parameter name (e.g., "t")
	Body      Expression // The lambda body expression
}

var _ Expression = (*LambdaExpression)(nil)

func (l *LambdaExpression) expressionNode()      {}
func (l *LambdaExpression) TokenLiteral() string { return l.Token.Literal }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token
	Value string
}

var _ Expression = (*StringLiteral)(nil)

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }

// IntegerLiteral represents an integer or long value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

var _ Expression = (*IntegerLiteral)(nil)

func (i *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

// FloatLiteral represents a float, double, or decimal value.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

var _ Expression = (*FloatLiteral)(nil)

func (f *FloatLiteral) expressionNode()      {}
func (f *FloatLiteral) TokenLiteral() string { return f.Token.Literal }

// UUIDLiteral represents a UUID value.
type UUIDLiteral struct {
	Token token.Token
	Value uuid.UUID
}

var _ Expression = (*UUIDLiteral)(nil)

func (g *UUIDLiteral) expressionNode()      {}
func (g *UUIDLiteral) TokenLiteral() string { return g.Token.Literal }

// DateTimeLiteral represents a datetime value (RFC3339).
type DateTimeLiteral struct {
	Token token.Token
	Value time.Time
}

var _ Expression = (*DateTimeLiteral)(nil)

func (d *DateTimeLiteral) expressionNode()      {}
func (d *DateTimeLiteral) TokenLiteral() string { return d.Token.Literal }

// DateLiteral represents a date-only value (YYYY-MM-DD).
type DateLiteral struct {
	Token token.Token
	Value time.Time
}

var _ Expression = (*DateLiteral)(nil)

func (d *DateLiteral) expressionNode()      {}
func (d *DateLiteral) TokenLiteral() string { return d.Token.Literal }

// NullLiteral represents a null value.
type NullLiteral struct {
	Token token.Token
}

var _ Expression = (*NullLiteral)(nil)

func (n *NullLiteral) expressionNode()      {}
func (n *NullLiteral) TokenLiteral() string { return n.Token.Literal }

// BooleanLiteral represents a boolean value (true/false).
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

var _ Expression = (*BooleanLiteral)(nil)

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
