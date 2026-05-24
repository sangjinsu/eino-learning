package tools

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const CalculatorToolName = "calculator"

var (
	ErrBlankExpression       = errors.New("calculator tool: expression is required")
	ErrUnsupportedExpression = errors.New("calculator tool: unsupported expression")
	ErrDivisionByZero        = errors.New("calculator tool: division by zero")
)

type CalculatorInput struct {
	Expression string `json:"expression" jsonschema:"description=Arithmetic expression using +, -, *, /, and parentheses,required"`
}

type CalculatorOutput struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

func NewCalculatorTool() (einotool.InvokableTool, error) {
	return utils.InferTool[CalculatorInput, CalculatorOutput](
		CalculatorToolName,
		"Evaluate a safe arithmetic expression using +, -, *, /, and parentheses.",
		Calculate,
	)
}

func Calculate(_ context.Context, input CalculatorInput) (CalculatorOutput, error) {
	expression := strings.TrimSpace(input.Expression)
	if expression == "" {
		return CalculatorOutput{}, ErrBlankExpression
	}

	parsed, err := parser.ParseExpr(expression)
	if err != nil {
		return CalculatorOutput{}, fmt.Errorf("%w: %v", ErrUnsupportedExpression, err)
	}

	result, err := evalExpression(parsed)
	if err != nil {
		return CalculatorOutput{}, err
	}

	return CalculatorOutput{
		Expression: expression,
		Result:     result,
	}, nil
}

func evalExpression(expr ast.Expr) (float64, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.INT && e.Kind != token.FLOAT {
			return 0, fmt.Errorf("%w: literal %s", ErrUnsupportedExpression, e.Value)
		}

		value, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrUnsupportedExpression, err)
		}
		return value, nil
	case *ast.BinaryExpr:
		left, err := evalExpression(e.X)
		if err != nil {
			return 0, err
		}
		right, err := evalExpression(e.Y)
		if err != nil {
			return 0, err
		}

		switch e.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, ErrDivisionByZero
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("%w: operator %s", ErrUnsupportedExpression, e.Op.String())
		}
	case *ast.ParenExpr:
		return evalExpression(e.X)
	case *ast.UnaryExpr:
		value, err := evalExpression(e.X)
		if err != nil {
			return 0, err
		}

		switch e.Op {
		case token.ADD:
			return value, nil
		case token.SUB:
			return -value, nil
		default:
			return 0, fmt.Errorf("%w: unary operator %s", ErrUnsupportedExpression, e.Op.String())
		}
	default:
		return 0, fmt.Errorf("%w: %T", ErrUnsupportedExpression, expr)
	}
}
