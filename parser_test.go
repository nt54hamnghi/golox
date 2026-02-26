package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func parseExpression(source string) (Expr, error) {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

func TestParsingExpressionSyntacticErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name:    "unterminated string in parse mode",
			source:  "\"quz",
			wantErr: "[line 1] Error: Unterminated string.",
		},
		{
			name:    "grouping with identifier without closing paren",
			source:  "(foo",
			wantErr: "[line 1] Error at 'foo': Expect expression.",
		},
		{
			name:    "operator without right operand",
			source:  "(67 +)",
			wantErr: "[line 1] Error at ')': Expect expression.",
		},
		{
			name:    "plus token alone",
			source:  "+",
			wantErr: "[line 1] Error at '+': Expect expression.",
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseExpression(tt.source)

			r.Error(err)
			r.Nil(expr)
			r.Equal(tt.wantErr, err.Error())
		})
	}
}

func assertParseOutput(t *testing.T, source string, want string) {
	t.Helper()
	r := require.New(t)

	expr, err := parseExpression(source)

	r.NoError(err)
	r.NotNil(expr)

	printer := AstPrinter{}
	r.Equal(want, printer.String(expr))
}

func TestParsingExpressionsBooleansAndNil(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"true literal", "true", "true"},
		{"false literal", "false", "false"},
		{"nil literal", "nil", "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsNumberLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"integer", "30", "30"},
		{"zero with decimal", "0.0", "0"},
		{"fractional", "10.12", "10.12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsStringLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"plain words", "\"quz bar\"", "quz bar"},
		{"quoted text inside string", "\"'quz'\"", "'quz'"},
		{"comment-like text in string", "\"// hello\"", "// hello"},
		{"numeric text in string", "\"95\"", "95"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsParentheses(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"grouped string", "(\"foo\")", "(group foo)"},
		{"nested grouping", "((true))", "(group (group true))"},
		{"grouped nil", "(nil)", "(group nil)"},
		{"grouped number", "(13.82)", "(group 13.82)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsUnaryOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"logical not", "!true", "(! true)"},
		{"numeric negation", "-87", "(- 87)"},
		{"double not", "!!true", "(! (! true))"},
		{"nested unary in grouping", "(!!(true))", "(group (! (! (group true))))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsArithmeticOperatorsOne(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"mixed multiply divide", "83 * 34 / 46", "(/ (* 83 34) 46)"},
		{"left associative division", "87 / 43 / 39", "(/ (/ 87 43) 39)"},
		{"multiple multiplies before divide", "26 * 57 * 49 / 30", "(/ (* (* 26 57) 49) 30)"},
		{"grouped expression with unary", "(65 * -60 / (35 * 42))", "(group (/ (* 65 (- 60)) (group (* 35 42))))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsArithmeticOperatorsTwo(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"string concatenation shape", "\"hello\" + \"world\"", "(+ hello world)"},
		{"mix subtraction and multiplication", "26 - 44 * 41 - 56", "(- (- 26 (* 44 41)) 56)"},
		{"addition subtraction and division", "76 + 32 - 88 / 84", "(- (+ 76 32) (/ 88 84))"},
		{"complex grouped arithmetic", "(-47 + 13) * (45 * 15) / (65 + 95)", "(/ (* (group (+ (- 47) 13)) (group (* 45 15))) (group (+ 65 95)))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsComparisonOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"greater than", "83 > 70", "(> 83 70)"},
		{"less equal", "13 <= 96", "(<= 13 96)"},
		{"chained less", "83 < 96 < 109", "(< (< 83 96) 109)"},
		{"grouped comparison with unary", "(94 - 24) >= -(60 / 56 + 99)", "(>= (group (- 94 24)) (- (group (+ (/ 60 56) 99))))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}

func TestParsingExpressionsEqualityOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"string inequality", "\"hello\"!=\"bar\"", "(!= hello bar)"},
		{"string equality", "\"foo\" == \"foo\"", "(== foo foo)"},
		{"number equality", "74 == 20", "(== 74 20)"},
		{"nested equality comparison", "(84 != 12) == ((-84 + 80) >= (24 * 84))", "(== (group (!= 84 12)) (group (>= (group (+ (- 84) 80)) (group (* 24 84)))))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseOutput(t, tt.source, tt.want)
		})
	}
}
