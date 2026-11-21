package parser

import (
	"fmt"
	"math/big"

	"github.com/birdie-lang/birdie/cmd/birdie/ast"
	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/lexer"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

const (
	_int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(x)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.DIV:      PRODUCT,
	token.DDIV:     PRODUCT,
	token.MUL:      PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l           *lexer.Lexer
	diagnostics []diagnostic.Diagnostic

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	// p.registerPrefix(token.LPAREN, p.parseGroupExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUN, p.parseFunctionLiteral)
	// p.registerPrefix(token.STRING, p.parseStringLiteral)
	// p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	// p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.DDIV, p.parseInfixExpression)
	p.registerInfix(token.MUL, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	// p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.EOL) {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		} else {
			p.synchronize()
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	// defer untrace(trace("parsePrefixExpression"))

	exp := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Text()}

	p.nextToken()

	exp.Right = p.parseExpression(PREFIX)

	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// defer untrace(trace("parseInfixExpression"))

	exp := &ast.InfixExpression{Token: p.curToken, Operator: p.curToken.Text(), Left: left}

	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	if exp.Right == nil {
		return nil
	}

	return exp
}

func (p *Parser) Diagnostics() []diagnostic.Diagnostic {
	return p.diagnostics
}

// Recover at a statement or block boundary so the next statement can be parsed.
func (p *Parser) synchronize() {
	for !p.curTokenIs(token.EOL) && !p.curTokenIs(token.DEDENT) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) peekError(t token.TokenType) {
	expected := fmt.Sprintf("%q", t.Spelling())
	switch t {
	case token.IDENT:
		expected = "an identifier"
	case token.INT:
		expected = "an integer"
	case token.STRING:
		expected = "a string"
	case token.EOL:
		expected = "end of line"
	case token.EOF:
		expected = "end of input"
	}
	p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
		Code: diagnostic.UnexpectedToken, Message: fmt.Sprintf("expected %s, found %s", expected, describeToken(p.peekToken)),
		Span: p.peekToken.Span,
	})
}

func describeToken(t token.Token) string {
	switch t.Type {
	case token.EOF:
		return "end of input"
	case token.EOL:
		return "end of line"
	case token.INDENT:
		return "an indentation increase"
	case token.DEDENT:
		return "an indentation decrease"
	default:
		return fmt.Sprintf("%q", t.Text())
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Parse ExpressionStatement or Assignment
func (p *Parser) parseExpressionStatement() ast.Statement {
	start := p.curToken
	left := p.parseExpression(LOWEST)
	if left == nil {
		return nil
	}

	if p.peekToken.Type != token.ASSIGN {
		return &ast.ExpressionStatement{
			Token:      start,
			Expression: left,
		}
	}

	// Initially, support assigning to identifiers only.
	name, ok := left.(*ast.Identifier)
	if !ok {
		p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
			Code: diagnostic.InvalidAssignmentTarget, Message: "cannot assign to this expression; expected a variable name",
			Span: source.Span{Start: start.Span.Start, End: p.curToken.Span.End},
		})
		return nil
	}

	p.nextToken() // Current token becomes "=".
	p.nextToken() // Current token becomes the first RHS token.

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	return &ast.AssignStatement{
		Token: start,
		Name:  name,
		Value: value,
	}
}

func (p *Parser) parseExpression(precedance int) ast.Expression {
	// defer untrace(trace("parseExpression"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// The lexer already diagnosed ILLEGAL tokens.
		if !p.curTokenIs(token.ILLEGAL) {
			p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
				Code: diagnostic.ExpectedExpression, Message: fmt.Sprintf("expected an expression, found %s", describeToken(p.curToken)),
				Span: p.curToken.Span,
			})
		}
		return nil
	}
	leftExp := prefix()
	if leftExp == nil {
		return nil
	}

	for !p.peekTokenIs(token.EOL) && precedance < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
		if leftExp == nil {
			return nil
		}
	}
	return leftExp
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)
	if stmt.ReturnValue == nil {
		return nil
	}
	return stmt
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	// defer untrace(trace("parseIntegerLiteral"))

	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, ok := new(big.Int).SetString(p.curToken.Literal, 0)
	if !ok {
		p.diagnostics = append(p.diagnostics, diagnostic.Diagnostic{
			Code: diagnostic.InvalidInteger, Message: fmt.Sprintf("invalid integer literal %q", p.curToken.Literal),
			Span: p.curToken.Span,
		})
		return nil
	}
	lit.Value = *value

	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	// defer untrace(trace("parseBoolean"))

	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) expectPeek(tokens ...token.TokenType) bool {
	for _, t := range tokens {
		if p.peekTokenIs(t) {
			p.nextToken()
		} else {
			p.peekError(t)
			return false
		}
	}
	return true
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}
	first, ok := p.parseIfBranch()
	if !ok {
		return nil
	}
	exp.If = first

	for p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.peekTokenIs(token.IF) {
			exp.Else = p.parseBlockStatement()
			if exp.Else == nil {
				return nil
			}
			break
		}

		p.nextToken()
		branch, ok := p.parseIfBranch()
		if !ok {
			return nil
		}
		exp.ElseIf = append(exp.ElseIf, branch)
	}

	return exp
}

// parseIfBranch starts with if in curToken.
func (p *Parser) parseIfBranch() (ast.IfBranch, bool) {
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return ast.IfBranch{}, false
	}
	body := p.parseBlockStatement()
	if body == nil {
		return ast.IfBranch{}, false
	}
	return ast.IfBranch{Condition: condition, Consequence: body}, true
}

// parseBlockStatement starts with the opening colon in peekToken.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if !p.expectPeek(token.COLON, token.EOL, token.INDENT) {
		return nil
	}

	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.DEDENT) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.EOL) {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	lit.Name = ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()
	if lit.Parameters == nil {
		return nil
	}

	if p.peekToken.Type == token.IDENT {
		p.nextToken()
		lit.ReturnType = ast.NamedType{Token: p.curToken, Name: p.curToken.Literal}
	}

	lit.Body = p.parseBlockStatement()
	if lit.Body == nil {
		return nil
	}

	return lit
}

func (p *Parser) parseFunctionParameters() []ast.Parameter {
	params := []ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	for {
		param, ok := p.parseFunctionParameter()
		if !ok {
			return nil
		}
		params = append(params, param)
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// Enter with curToken on '(' or ',' and peekToken on the parameter name.
func (p *Parser) parseFunctionParameter() (ast.Parameter, bool) {
	if !p.expectPeek(token.IDENT) {
		return ast.Parameter{}, false
	}
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.COLON) || !p.expectPeek(token.IDENT) {
		return ast.Parameter{}, false
	}

	return ast.Parameter{
		Name: name,
		Type: &ast.NamedType{Token: p.curToken, Name: p.curToken.Literal},
	}, true
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	ce := &ast.CallExpression{Token: p.curToken, Function: left}
	ce.Arguments = p.parseExpressionList(token.RPAREN)
	if ce.Arguments == nil {
		return nil
	}
	return ce
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	for {
		p.nextToken()
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}
		list = append(list, expr)
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
