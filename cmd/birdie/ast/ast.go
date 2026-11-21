package ast

import (
	"math/big"
	"strings"

	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out strings.Builder

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *AssignStatement) statementNode()       {}
func (ls *AssignStatement) TokenLiteral() string { return ls.Token.Text() }
func (ls *AssignStatement) String() string {
	var out strings.Builder
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	return out.String()
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Text() }
func (rs *ReturnStatement) String() string {
	var out strings.Builder

	out.WriteString(rs.TokenLiteral())
	out.WriteString(" ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Text() }
func (es *ExpressionStatement) String() string {
	var out strings.Builder

	if es.Expression != nil {
		out.WriteString(es.Expression.String())
	}

	return out.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Text() }
func (bs *BlockStatement) String() string {
	lines := make([]string, len(bs.Statements))
	for i, stmt := range bs.Statements {
		lines[i] = "    " + strings.ReplaceAll(stmt.String(), "\n", "\n    ")
	}
	return strings.Join(lines, "\n")
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Text() }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value big.Int
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Text() }
func (il *IntegerLiteral) String() string       { return il.Value.String() }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Text() }
func (b *Boolean) String() string       { return b.Token.Text() }

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Text() }
func (pe *PrefixExpression) String() string {
	var out strings.Builder

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Text() }
func (ie *InfixExpression) String() string {
	var out strings.Builder

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type IfExpression struct {
	Token  token.Token
	If     IfBranch
	ElseIf []IfBranch
	Else   *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Text() }
func (ie *IfExpression) String() string {
	var out strings.Builder
	out.WriteString("if ")
	out.WriteString(ie.If.String())
	for _, elseIf := range ie.ElseIf {
		out.WriteString("\nelse if ")
		out.WriteString(elseIf.String())
	}
	if ie.Else != nil {
		out.WriteString("\nelse:\n")
		out.WriteString(ie.Else.String())
	}

	return out.String()
}

type IfBranch struct {
	Condition   Expression
	Consequence *BlockStatement
}

func (ib IfBranch) String() string {
	return ib.Condition.String() + ":\n" + ib.Consequence.String()
}
