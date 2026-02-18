package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func parseExpression(source string) (Expr[string], error) {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return nil, err
	}

	parser := NewParser[string](tokens)
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
