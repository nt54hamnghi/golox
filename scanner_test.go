package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScannerLexicalErrors(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantErrs   []string
		wantTokens []TokenType
	}{
		{
			name:       "single unexpected character",
			source:     "@",
			wantErrs:   []string{"[line 1] Error: Unexpected character: @"},
			wantTokens: []TokenType{EOF},
		},
		{
			name:   "mixed valid and invalid characters",
			source: ",.$(#",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: $",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []TokenType{COMMA, DOT, LEFT_PAREN, EOF},
		},
		{
			name:   "all invalid characters",
			source: "$@%#@",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: $",
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: %",
				"[line 1] Error: Unexpected character: #",
				"[line 1] Error: Unexpected character: @",
			},
			wantTokens: []TokenType{EOF},
		},
		{
			name:   "valid tokens continue around lexical errors",
			source: "{(+.@-$;#)}",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: $",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []TokenType{LEFT_BRACE, LEFT_PAREN, PLUS, DOT, MINUS, SEMICOLON, RIGHT_PAREN, RIGHT_BRACE, EOF},
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.source)
			tokens, err := scanner.ScanTokens()

			r.Error(err)

			gotErrs := strings.Split(err.Error(), "\n")
			r.Len(gotErrs, len(tt.wantErrs))
			for i := range tt.wantErrs {
				r.Equal(tt.wantErrs[i], gotErrs[i], "error[%d] mismatch", i)
			}

			r.Len(tokens, len(tt.wantTokens))
			for i := range tt.wantTokens {
				r.Equal(tt.wantTokens[i], tokens[i].Type, "token[%d] type mismatch", i)
			}
		})
	}
}

func TestScannerMultilineErrors(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantErrs   []string
		wantTokens []TokenType
	}{
		{
			name:       "error on second line after valid first line",
			source:     "()\n\t@",
			wantErrs:   []string{"[line 2] Error: Unexpected character: @"},
			wantTokens: []TokenType{LEFT_PAREN, RIGHT_PAREN, EOF},
		},
		{
			name:   "multiple unexpected characters on same line",
			source: " @%#",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: %",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []TokenType{EOF},
		},
		{
			name:   "errors across lines with comments and valid tokens",
			source: "()  #\t{}\n@\n$\n+++\n// Let's Go!\n+++\n#",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: #",
				"[line 2] Error: Unexpected character: @",
				"[line 3] Error: Unexpected character: $",
				"[line 7] Error: Unexpected character: #",
			},
			wantTokens: []TokenType{
				LEFT_PAREN,
				RIGHT_PAREN,
				LEFT_BRACE,
				RIGHT_BRACE,
				PLUS,
				PLUS,
				PLUS,
				PLUS,
				PLUS,
				PLUS,
				EOF,
			},
		},
		{
			name:       "newline splits valid punctuation and error",
			source:     "({;\n$})",
			wantErrs:   []string{"[line 2] Error: Unexpected character: $"},
			wantTokens: []TokenType{LEFT_PAREN, LEFT_BRACE, SEMICOLON, RIGHT_BRACE, RIGHT_PAREN, EOF},
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.source)
			tokens, err := scanner.ScanTokens()

			r.Error(err)

			gotErrs := strings.Split(err.Error(), "\n")
			r.Len(gotErrs, len(tt.wantErrs))
			for i := range tt.wantErrs {
				r.Equal(tt.wantErrs[i], gotErrs[i], "error[%d] mismatch", i)
			}

			r.Len(tokens, len(tt.wantTokens))
			for i := range tt.wantTokens {
				r.Equal(tt.wantTokens[i], tokens[i].Type, "token[%d] type mismatch", i)
			}
		})
	}
}
