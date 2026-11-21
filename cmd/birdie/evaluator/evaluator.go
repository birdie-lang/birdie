package evaluator

import (
	"fmt"
	"math/big"

	"github.com/birdie-lang/birdie/cmd/birdie/ast"
	"github.com/birdie-lang/birdie/cmd/birdie/diagnostic"
	"github.com/birdie-lang/birdie/cmd/birdie/object"
	"github.com/birdie-lang/birdie/cmd/birdie/source"
	"github.com/birdie-lang/birdie/cmd/birdie/token"
)

var (
	NONE  = &object.None{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(
	node ast.Node,
	env *object.Environment,
) (object.Object, *diagnostic.Diagnostic) {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, nil
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value), nil
	case *ast.PrefixExpression:
		right, err := Eval(node.Right, env)
		if err != nil {
			return nil, err
		}
		return evalPrefixExpression(node.Token, right)
	case *ast.InfixExpression:
		left, err := Eval(node.Left, env)
		if err != nil {
			return nil, err
		}
		right, err := Eval(node.Right, env)
		if err != nil {
			return nil, err
		}
		return evalInfixExpression(node.Token, left, right)
	case *ast.AssignStatement:
		val, err := Eval(node.Value, env)
		if err != nil {
			return nil, err
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val, err := Eval(node.ReturnValue, env)
		if err != nil {
			return nil, err
		}
		return &object.ReturnValue{Value: val}, nil
	case *ast.FunctionLiteral:
		env.Set(node.Name.Value, &object.Function{Decl: node, Env: env})
	case *ast.CallExpression:
		function, err := Eval(node.Function, env)
		if err != nil {
			return nil, err
		}
		args, err := evalExpressions(node.Arguments, env)
		if err != nil {
			return nil, err
		}

		return applyFunction(function, args, node.Token.Span)
		// 	case *ast.StringLiteral:
		// 		return &object.String{Value: node.Value}
		// 	case *ast.ArrayLiteral:
		// 		elements := evalExpressions(node.Elements, env)
		// 		if len(elements) == 1 && isError(elements[0]) {
		// 			return elements[0]
		// 		}
		// 		return &object.Array{Elements: elements}
		// 	case *ast.IndexExpression:
		// 		left := Eval(node.Left, env)
		// 		if isError(left) {
		// 			return left
		// 		}
		// 		index := Eval(node.Index, env)
		// 		if isError(index) {
		// 			return index
		// 		}
		// 		return evalIndexExpression(left, index)
		// 	case *ast.HashLiteral:
		// 		return evalHashLiteral(node, env)
		//
	}
	return nil, nil
}

func evalExpressions(exps []ast.Expression, env *object.Environment) ([]object.Object, *diagnostic.Diagnostic) {
	var result []object.Object

	for _, e := range exps {
		evaluated, err := Eval(e, env)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) (object.Object, *diagnostic.Diagnostic) {
	if val, ok := env.Get(node.Value); ok {
		return val, nil
	}

	// if builtin, ok := builtins[node.Value]; ok {
	// 	return builtin
	// }

	span := node.Token.Span
	return nil, newError(diagnostic.UnboundVariable, span, "identifier not found: %s", node.Value)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NONE:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) (object.Object, *diagnostic.Diagnostic) {
	condition, err := Eval(ie.If.Condition, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(condition) {
		return Eval(ie.If.Consequence, env)
	}

	for _, branch := range ie.ElseIf {
		condition, err := Eval(branch.Condition, env)
		if err != nil {
			return nil, err
		}
		if isTruthy(condition) {
			return Eval(branch.Consequence, env)
		}
	}

	if ie.Else != nil {
		return Eval(ie.Else, env)
	}
	return NONE, nil
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator token.Token, right object.Object) (object.Object, *diagnostic.Diagnostic) {
	switch operator.Type {
	case token.NOT:
		return evalBangOperatorExpression(right), nil
	case token.MINUS:
		return evalMinusPrefixOperatorExpression(operator, right)
	default:
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "unknown operator: %s%s", operator.Text(), right.Type())
	}
}

func evalInfixExpression(operator token.Token, left, right object.Object) (object.Object, *diagnostic.Diagnostic) {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator.Type == token.EQ:
		return nativeBoolToBooleanObject(left == right), nil
	case operator.Type == token.NEQ:
		return nativeBoolToBooleanObject(left != right), nil
	case left.Type() != right.Type():
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "type mismatch: %s %s %s", left.Type(), operator.Text(), right.Type())
	default:
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "unknown operator: %s %s %s", left.Type(), operator.Text(), right.Type())
	}
}

func evalStringInfixExpression(operator token.Token, left, right object.Object) (object.Object, *diagnostic.Diagnostic) {
	if operator.Type != token.PLUS {
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "unknown operator: %s %s %s", left.Type(), operator.Text(), right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}, nil
}

func evalIntegerInfixExpression(operator token.Token, left, right object.Object) (object.Object, *diagnostic.Diagnostic) {
	leftVal := &left.(*object.Integer).Value
	rightVal := &right.(*object.Integer).Value

	switch operator.Type {
	case token.PLUS:
		result := &object.Integer{}
		result.Value.Add(leftVal, rightVal)
		return result, nil
	case token.MINUS:
		result := &object.Integer{}
		result.Value.Sub(leftVal, rightVal)
		return result, nil
	case token.MUL:
		result := &object.Integer{}
		result.Value.Mul(leftVal, rightVal)
		return result, nil
	case token.DIV:
		result := &object.Integer{}
		if rightVal.Sign() == 0 {
			return nil, newError(diagnostic.DivisionByZero, operator.Span, "division by zero")
		}
		result.Value.Quo(leftVal, rightVal)
		return result, nil
	case token.DDIV:
		if rightVal.Sign() == 0 {
			return nil, newError(
				diagnostic.DivisionByZero,
				operator.Span,
				"division by zero",
			)
		}

		result := &object.Integer{}
		result.Value.Divide(leftVal, rightVal, nil, big.Floor)
		return result, nil
	case token.LT:
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) < 0), nil
	case token.GT:
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) > 0), nil
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) == 0), nil
	case token.NEQ:
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) != 0), nil
	default:
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "unknown operator: %s %s %s", left.Type(), operator.Text(), right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NONE:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(operator token.Token, right object.Object) (object.Object, *diagnostic.Diagnostic) {
	if right.Type() != object.INTEGER_OBJ {
		return nil, newError(diagnostic.UnknownOperator, operator.Span, "unknown operator: -%s", right.Type())
	}

	value := &right.(*object.Integer).Value
	return &object.Integer{Value: *new(big.Int).Neg(value)}, nil
}

func evalStatements(statements []ast.Statement, env *object.Environment) (object.Object, *diagnostic.Diagnostic) {
	var result object.Object
	var err *diagnostic.Diagnostic

	for _, statement := range statements {
		result, err = Eval(statement, env)

		if err != nil {
			return nil, err
		}

		if _, ok := result.(*object.ReturnValue); ok {
			return result, nil
		}
	}

	return result, nil
}

func applyFunction(fn object.Object, args []object.Object, callSpan source.Span) (object.Object, *diagnostic.Diagnostic) {
	switch fn := fn.(type) {
	case *object.Function:
		decl := fn.Decl
		if len(decl.Parameters) != len(args) {
			return nil, newError(diagnostic.ArgumentCountMismatch, callSpan,
				"%s expects %d arguments, but got %d", decl.Name.Value, len(decl.Parameters), len(args))
		}
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated, err := Eval(fn.Decl.Body, extendedEnv)
		if err != nil {
			return nil, err
		}
		return unwrapReturnValue(evaluated), nil
	// case *object.Builtin:
	// 	if result := fn.Fn(args...); result != nil {
	// 		return result
	// 	}
	// 	return NULL
	default:
		return nil, newError(diagnostic.NoFunction, callSpan, "not a function: %s", fn.Type())
	}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewClosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Decl.Parameters {
		env.Set(param.Name.Value, args[paramIdx])
	}

	return env
}

func evalProgram(statements []ast.Statement, env *object.Environment) (object.Object, *diagnostic.Diagnostic) {
	registerFunctions(statements, env)
	return evalStatements(statements, env)
}

func registerFunctions(statements []ast.Statement, env *object.Environment) {
	for _, stmt := range statements {
		expr, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			continue
		}
		if fn, ok := expr.Expression.(*ast.FunctionLiteral); ok {
			env.Set(fn.Name.Value, &object.Function{Decl: fn, Env: env})
		}
	}
}

func newError(code diagnostic.Code, span source.Span, message string, a ...any) *diagnostic.Diagnostic {
	message = fmt.Sprintf(message, a...)
	return &diagnostic.Diagnostic{
		Code:    code,
		Message: message,
		Span:    span,
	}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
