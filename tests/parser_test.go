package test_test

import (
	"fmt"
	"katan/src"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PrefixTuple struct {
	Operator string
	Value    interface{}
}

type InfixTuple struct {
	Left     interface{}
	Operator string
	Rigth    interface{}
}

type ParserTests struct {
	suite.Suite
}

func (p *ParserTests) InitParserTests(source string) (*src.Parser, *src.Program) {
	lexer := src.NewLexer(source)
	parser := src.NewParser(lexer)
	program := parser.ParseProgam()

	return parser, &program
}

func (p *ParserTests) TestParseProgram() {
	source := "var x = 5;"
	_, program := p.InitParserTests(source)

	p.Assert().NotNil(program)

	p.Assert().IsType(&src.Program{}, program)

	p.Assert().Implements((*src.ASTNode)(nil), program)
}

func (p *ParserTests) TestLetStatements() {
	source := `
		var x = 5;
		var y = 10;
		var foo = 20;
		var bar = verdadero;
	`
	_, program := p.InitParserTests(source)

	p.Assert().Equal(4, len(program.Staments))
	expected := []PrefixTuple{
		{Operator: "x", Value: 5},
		{Operator: "y", Value: 10},
		{Operator: "foo", Value: 20},
		{Operator: "bar", Value: true},
	}

	for i, statement := range program.Staments {
		p.Assert().Equal("var", statement.TokenLiteral())
		p.Assert().IsType(&src.LetStatement{}, statement.(*src.LetStatement))

		letStatement := statement.(*src.LetStatement)
		p.Assert().NotNil(letStatement.Name)
		p.testIdentifier(letStatement.Name, expected[i].Operator)

		p.Assert().NotNil(letStatement.Value)
		p.testLiteralExpression(letStatement.Value, expected[i].Value)
	}
}

func (p *ParserTests) TestNamesInLetStatements() {
	source := `
		var x = 5;
		var y = 10;
		var foo = 20;
	`
	_, program := p.InitParserTests(source)
	p.Assert().Equal(3, len(program.Staments))

	var names []string
	for _, stament := range program.Staments {
		stament := stament.(*src.LetStatement)
		if !p.Assert().NotNil(stament.Name) {
			p.T().Fail()
		}

		if !p.Assert().Implements((*src.Stmt)(nil), stament) {
			p.T().Fail()
		}

		names = append(names, stament.Name.Str())
	}

	expectedNames := []string{"x", "y", "foo"}
	if !p.Assert().Equal(expectedNames, names) {
		p.T().Fail()
	}

}

func (p *ParserTests) TestParseErrors() {
	source := "var x 5;"
	parser, _ := p.InitParserTests(source)
	if !p.Assert().Equal(1, len(parser.Errors())) {
		p.T().Fail()
	}
}

func (p *ParserTests) TestReturnStatement() {
	source := `
		regresa 5;
		regresa foo;
		regresa verdadero;
		regresa falso;
	`

	_, program := p.InitParserTests(source)

	if !p.Assert().Equal(4, len(program.Staments)) {
		p.T().Log("len of program statements are not 2")
		p.T().Fail()
	}

	expected := []PrefixTuple{
		{Operator: "regresa", Value: 5},
		{Operator: "regresa", Value: "foo"},
		{Operator: "regresa", Value: true},
		{Operator: "regresa", Value: false},
	}

	for i, statement := range program.Staments {
		p.Assert().Equal("regresa", statement.TokenLiteral())
		p.Assert().IsType(&src.ReturnStament{}, statement.(*src.ReturnStament))

		returnStament := statement.(*src.ReturnStament)
		p.Assert().NotNil(returnStament.ReturnValue)
		p.testLiteralExpression(returnStament.ReturnValue, expected[i].Value)
	}
}

func (p *ParserTests) TestIdentifierExpression() {
	source := "foobar;"
	parser, program := p.InitParserTests(source)

	p.testProgramStatements(parser, program, 1)

	expressionStament := program.Staments[0].(*src.ExpressionStament)
	if !p.Assert().NotNil(expressionStament.Expression) {
		p.T().Fail()
	}

	p.testLiteralExpression(expressionStament.Expression, "foobar")
}

func (p *ParserTests) TestIntegerExpressions() {
	source := "5;"
	parser, program := p.InitParserTests(source)

	p.testProgramStatements(parser, program, 1)
	expressionStament := program.Staments[0].(*src.ExpressionStament)
	p.Assert().NotNil(expressionStament.Expression)
	p.testLiteralExpression(expressionStament.Expression, 5)
}

func (p *ParserTests) TestPrefixExpressions() {
	source := "!5; -15; !verdadero; !falso;"
	parser, program := p.InitParserTests(source)

	p.testProgramStatements(parser, program, 4)
	expectedExpressions := []PrefixTuple{
		{Operator: "!", Value: 5},
		{Operator: "-", Value: 15},
		{Operator: "!", Value: true},
		{Operator: "!", Value: false},
	}

	if len(program.Staments) == len(expectedExpressions) {
		for i, stament := range program.Staments {
			stament := stament.(*src.ExpressionStament)
			p.Assert().IsType(&src.Prefix{}, stament.Expression.(*src.Prefix))

			prefix := stament.Expression.(*src.Prefix)
			p.Assert().Equal(prefix.Operator, expectedExpressions[i].Operator)

			p.Assert().NotNil(prefix.Rigth)
			p.testLiteralExpression(prefix.Rigth, expectedExpressions[i].Value)
		}
	} else {
		p.T().Log("len of staments and expected expected expressions are not equal")
		p.T().Fail()
	}
}

func (p *ParserTests) TestCallExpression() {
	source := "suma(1, 2 * 3, 4 + 5);"
	parser, program := p.InitParserTests(source)
	p.testProgramStatements(parser, program, 1)

	call := (program.Staments[0].(*src.ExpressionStament)).Expression.(*src.Call)
	p.Assert().IsType(&src.Call{}, call)
	p.testIdentifier(call.Function, "suma")
	p.Assert().NotNil(call.Arguments)

	p.Assert().Equal(3, len(call.Arguments))
	p.testLiteralExpression(call.Arguments[0], 1)
	p.testInfixExpression(call.Arguments[1], 2, "*", 3)
	p.testInfixExpression(call.Arguments[2], 4, "+", 5)
}

func (p *ParserTests) TestFunctionLiteral() {
	source := "funcion(x, y) { x + y }"
	parser, program := p.InitParserTests(source)
	p.testProgramStatements(parser, program, 1)

	functionLiteral := (program.Staments[0].(*src.ExpressionStament)).Expression.(*src.Function)
	p.Assert().IsType(&src.Function{}, functionLiteral)
	p.Assert().Equal(2, len(functionLiteral.Parameters))

	p.testLiteralExpression(functionLiteral.Parameters[0], "x")
	p.testLiteralExpression(functionLiteral.Parameters[1], "y")
	p.Assert().NotNil(functionLiteral.Body)

	p.Assert().Equal(1, len(functionLiteral.Body.Staments))
	body := functionLiteral.Body.Staments[0].(*src.ExpressionStament)
	p.Assert().NotNil(body.Expression)
	p.testInfixExpression(body.Expression, "x", "+", "y")
}

func (p *ParserTests) TestFunctionParameter() {
	tests := []map[string]interface{}{
		{
			"input":    "funcion() {};",
			"expected": []string{},
		},
		{
			"input":    "funcion(x) {};",
			"expected": []string{"x"},
		},
		{
			"input":    "funcion(x, y, z) {};",
			"expected": []string{"x", "y", "z"},
		},
	}

	for _, test := range tests {
		_, program := p.InitParserTests(test["input"].(string))
		function := (program.Staments[0].(*src.ExpressionStament)).Expression.(*src.Function)
		p.Assert().Equal(len(test["expected"].([]string)), len(function.Parameters))

		for idx, param := range test["expected"].([]string) {
			p.testLiteralExpression(function.Parameters[idx], param)
		}
	}
}

func (p *ParserTests) TestInfixExpressions() {
	source := `
		5 + 5;
		5 - 5;
		5 * 5;
		5 / 5;
		5 > 5;
		5 < 5;
		5 == 5;
		5 != 5;
		verdadero == verdadero;
		verdadero != verdadero;
		falso == verdadero;
		falso != verdadero;
		
	`
	parser, program := p.InitParserTests(source)

	p.testProgramStatements(parser, program, 12)
	expectedOperators := []InfixTuple{
		{Left: 5, Operator: "+", Rigth: 5},
		{Left: 5, Operator: "-", Rigth: 5},
		{Left: 5, Operator: "*", Rigth: 5},
		{Left: 5, Operator: "/", Rigth: 5},
		{Left: 5, Operator: ">", Rigth: 5},
		{Left: 5, Operator: "<", Rigth: 5},
		{Left: 5, Operator: "==", Rigth: 5},
		{Left: 5, Operator: "!=", Rigth: 5},
		{Left: true, Operator: "==", Rigth: true},
		{Left: true, Operator: "!=", Rigth: true},
		{Left: false, Operator: "==", Rigth: true},
		{Left: false, Operator: "!=", Rigth: true},
	}

	for i, stamment := range program.Staments {
		stament := stamment.(*src.ExpressionStament)
		p.Assert().NotNil(stament.Expression)
		p.Assert().IsType(&src.Infix{}, stament.Expression.(*src.Infix))
		p.testInfixExpression(
			stament.Expression,
			expectedOperators[i].Left,
			expectedOperators[i].Operator,
			expectedOperators[i].Rigth,
		)
	}
}

func (p *ParserTests) TestIfExpression() {
	source := "si (x > y) { z } si_no { w }"
	parser, program := p.InitParserTests(source)
	p.testProgramStatements(parser, program, 1)

	ifExpression := (program.Staments[0].(*src.ExpressionStament)).Expression.(*src.If)
	p.Assert().IsType(&src.If{}, ifExpression)
	p.Assert().NotNil(ifExpression.Condition)

	p.testInfixExpression(ifExpression.Condition, "x", ">", "y")
	p.Assert().NotNil(ifExpression.Consequence)
	p.Assert().IsType(&src.Block{}, ifExpression.Consequence)
	p.Assert().Equal(1, len(ifExpression.Consequence.Staments))

	consequenceStament := ifExpression.Consequence.Staments[0].(*src.ExpressionStament)
	p.Assert().NotNil(consequenceStament.Expression)
	p.testIdentifier(consequenceStament.Expression, "z")

	p.Assert().NotNil(ifExpression.Alternative)
	p.Assert().IsType(&src.Block{}, ifExpression.Alternative)
	p.Assert().Equal(1, len(ifExpression.Alternative.Staments))

	alternativeStament := ifExpression.Alternative.Staments[0].(*src.ExpressionStament)
	p.Assert().NotNil(alternativeStament.Expression)
	p.testIdentifier(alternativeStament.Expression, "w")
}

func (p *ParserTests) TestBooleanExpressions() {
	source := "verdadero; falso;"
	parser, program := p.InitParserTests(source)

	p.testProgramStatements(parser, program, 2)
	expectedValues := []bool{true, false}

	for i, stament := range program.Staments {
		expressionStament := stament.(*src.ExpressionStament)
		p.Assert().NotNil(expressionStament.Expression)
		p.testLiteralExpression(expressionStament.Expression, expectedValues[i])
	}
}

func (p *ParserTests) TestOperatorPrecedence() {
	type TupleToTest struct {
		source        string
		expected      string
		expectedCount int
	}
	test_source := []TupleToTest{
		{source: "-a * b;", expected: "((- a) * b)", expectedCount: 1},
		{source: "!-a;", expected: "(! (- a))", expectedCount: 1},
		{source: "a + b / c;", expected: "(a + (b / c))", expectedCount: 1},
		{source: "3 + 4; -5 * 5;", expected: "(3 + 4) ((- 5) * 5)", expectedCount: 2},
		{source: "a + b * c + d / e - f;", expected: "(((a + (b * c)) + (d / e)) - f)", expectedCount: 1},
		{source: "1 + (2 + 3) + 4;", expected: "((1 + (2 + 3)) + 4)", expectedCount: 1},
		{source: "(5 + 5) * 2;", expected: "((5 + 5) * 2)", expectedCount: 1},
		{source: "2 / (5 + 5);", expected: "(2 / (5 + 5))", expectedCount: 1},
		{source: "-(5 + 5);", expected: "(- (5 + 5))", expectedCount: 1},
		{source: "-(5 + 5);", expected: "(- (5 + 5))", expectedCount: 1},
		{source: "a + suma(b * c) + d;", expected: "((a + suma((b * c))) + d)", expectedCount: 1},
		{source: "a + suma(b * c) + d;", expected: "((a + suma((b * c))) + d)", expectedCount: 1},
		{
			source:        "suma(a, b, 1, 2 * 3, 4 + 5, suma(6, 7 * 8))",
			expected:      "suma(a, b, 1, (2 * 3), (4 + 5), suma(6, (7 * 8)))",
			expectedCount: 1,
		},
	}

	for _, source := range test_source {
		parser, program := p.InitParserTests(source.source)

		p.testProgramStatements(parser, program, source.expectedCount)
		p.Assert().Equal(source.expected, program.Str())
	}
}

func (p *ParserTests) TestStringLiteral() {
	source := `"hello world!";`

	_, program := p.InitParserTests(source)
	expressionStatement := program.Staments[0].(*src.ExpressionStament)
	stringLiteral := expressionStatement.Expression.(*src.StringLiteral)

	p.Assert().IsType(&src.StringLiteral{}, stringLiteral)
	p.Assert().Equal("hello world!", stringLiteral.Value)
}

func (p *ParserTests) testBoolean(expression src.Expression, expectedValue bool) {
	boolean := expression.(*src.Boolean)
	p.Assert().Equal(*boolean.Value, expectedValue)
	if expectedValue {
		p.Assert().Equal("verdadero", boolean.Token.Literal)
	} else {
		p.Assert().Equal("falso", boolean.Token.Literal)
	}
}

func (p *ParserTests) testInfixExpression(ex src.Expression, expectedLeft interface{}, operator string, expectedRigth interface{}) {
	infix := ex.(*src.Infix)
	p.Assert().NotNil(infix.Left)
	p.testLiteralExpression(infix.Left, expectedLeft)
	p.Assert().Equal(operator, infix.Operator)
	p.Assert().NotNil(infix.Rigth)
	p.testLiteralExpression(infix.Rigth, expectedRigth)
}

func (p *ParserTests) testProgramStatements(parser *src.Parser, program *src.Program, expectedStamenetCount int) {
	p.Assert().Equal(0, len(parser.Errors()))
	p.Assert().Equal(expectedStamenetCount, len(program.Staments))
	p.Assert().IsType(&src.ExpressionStament{}, program.Staments[0].(*src.ExpressionStament))
}

func (p *ParserTests) testLiteralExpression(expression src.Expression, expectedValue interface{}) {
	switch expectedValue := expectedValue.(type) {
	case string:
		p.testIdentifier(expression, expectedValue)
	case int:
		p.testInteger(expression, expectedValue)
	case bool:
		p.testBoolean(expression, expectedValue)
	default:
		p.T().Log(fmt.Sprintf("unhandled type of expression, Got=%s", reflect.TypeOf(expectedValue).String()))
		p.T().Fail()
	}
}

func (p *ParserTests) testIdentifier(expression src.Expression, expectedValue string) {
	p.Assert().IsType(&src.Identifier{}, expression.(*src.Identifier))

	identifier := expression.(*src.Identifier)
	p.Assert().Equal(expectedValue, identifier.Str())
	p.Assert().Equal(expectedValue, identifier.TokenLiteral())
}

func (p *ParserTests) testInteger(expression src.Expression, expectedValue int) {
	p.Assert().IsType(&src.Integer{}, expression.(*src.Integer))
	integer := expression.(*src.Integer)
	p.Assert().Equal(expectedValue, *integer.Value)
	p.Assert().Equal(fmt.Sprint(expectedValue), integer.Token.Literal)
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserTests))
}
