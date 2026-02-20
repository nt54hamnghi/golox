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

func TestScannerEmptyFile(t *testing.T) {
	scanner := NewScanner("")
	tokens, err := scanner.ScanTokens()

	r := require.New(t)
	r.NoError(err)
	r.Len(tokens, 1)
	r.Equal(EOF, tokens[0].Type)
}

func assertScanTokenTypes(t *testing.T, source string, want []TokenType) {
	// Helper marks the calling function as a test helper function.
	t.Helper()
	r := require.New(t)

	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()

	r.NoError(err)
	r.Len(tokens, len(want))
	for i := range want {
		r.Equal(want[i], tokens[i].Type, "token[%d] type mismatch", i)
	}
}

func assertScanTokenTypesAndErrors(t *testing.T, source string, want []TokenType, wantErrs []string) {
	// Helper marks the calling function as a test helper function.
	t.Helper()
	r := require.New(t)

	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()

	r.Error(err)
	r.Equal(wantErrs, strings.Split(err.Error(), "\n"))
	r.Len(tokens, len(want))
	for i := range want {
		r.Equal(want[i], tokens[i].Type, "token[%d] type mismatch", i)
	}
}

func TestScannerParentheses(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"single left paren", "(", []TokenType{LEFT_PAREN, EOF}},
		{"double right paren", "))", []TokenType{RIGHT_PAREN, RIGHT_PAREN, EOF}},
		{"mixed right and left", ")))((", []TokenType{RIGHT_PAREN, RIGHT_PAREN, RIGHT_PAREN, LEFT_PAREN, LEFT_PAREN, EOF}},
		{"nested sequence", "()((())", []TokenType{LEFT_PAREN, RIGHT_PAREN, LEFT_PAREN, LEFT_PAREN, LEFT_PAREN, RIGHT_PAREN, RIGHT_PAREN, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerBraces(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"single right brace", "}", []TokenType{RIGHT_BRACE, EOF}},
		{"pair of braces", "{{}}", []TokenType{LEFT_BRACE, LEFT_BRACE, RIGHT_BRACE, RIGHT_BRACE, EOF}},
		{"alternating braces", "}{}{{", []TokenType{RIGHT_BRACE, LEFT_BRACE, RIGHT_BRACE, LEFT_BRACE, LEFT_BRACE, EOF}},
		{"mixed braces and parens", "{(){()}", []TokenType{LEFT_BRACE, LEFT_PAREN, RIGHT_PAREN, LEFT_BRACE, LEFT_PAREN, RIGHT_PAREN, RIGHT_BRACE, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerOtherSingleCharacterTokens(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"plus minus", "+-", []TokenType{PLUS, MINUS, EOF}},
		{"all single character punctuators", "++--**..,,;;", []TokenType{PLUS, PLUS, MINUS, MINUS, STAR, STAR, DOT, DOT, COMMA, COMMA, SEMICOLON, SEMICOLON, EOF}},
		{"mixed punctuation order", "-+*,+*;", []TokenType{MINUS, PLUS, STAR, COMMA, PLUS, STAR, SEMICOLON, EOF}},
		{"single character tokens in grouping", "({*,+-.})", []TokenType{LEFT_PAREN, LEFT_BRACE, STAR, COMMA, PLUS, MINUS, DOT, RIGHT_BRACE, RIGHT_PAREN, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerAssignmentAndEqualityOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		want     []TokenType
		wantErrs []string
	}{
		{"single equal", "=", []TokenType{EQUAL, EOF}, nil},
		{"double equal", "==", []TokenType{EQUAL_EQUAL, EOF}, nil},
		{"grouped equal operators", "({=}){==}", []TokenType{LEFT_PAREN, LEFT_BRACE, EQUAL, RIGHT_BRACE, RIGHT_PAREN, LEFT_BRACE, EQUAL_EQUAL, RIGHT_BRACE, EOF}, nil},
		{
			"operators mixed with lexical errors",
			"((==#%=$))",
			[]TokenType{LEFT_PAREN, LEFT_PAREN, EQUAL_EQUAL, EQUAL, RIGHT_PAREN, RIGHT_PAREN, EOF},
			[]string{
				"[line 1] Error: Unexpected character: #",
				"[line 1] Error: Unexpected character: %",
				"[line 1] Error: Unexpected character: $",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantErrs) == 0 {
				assertScanTokenTypes(t, tt.source, tt.want)
				return
			}
			assertScanTokenTypesAndErrors(t, tt.source, tt.want, tt.wantErrs)
		})
	}
}

func TestScannerNegationAndInequalityOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		want     []TokenType
		wantErrs []string
	}{
		{"bang equal", "!=", []TokenType{BANG_EQUAL, EOF}, nil},
		{"bang and equality chain", "!!===", []TokenType{BANG, BANG_EQUAL, EQUAL_EQUAL, EOF}, nil},
		{"bang operators with grouping", "!{!}(!===)=", []TokenType{BANG, LEFT_BRACE, BANG, RIGHT_BRACE, LEFT_PAREN, BANG_EQUAL, EQUAL_EQUAL, RIGHT_PAREN, EQUAL, EOF}, nil},
		{
			"unexpected chars among bang tokens",
			"{(!==@%!)}",
			[]TokenType{LEFT_BRACE, LEFT_PAREN, BANG_EQUAL, EQUAL, BANG, RIGHT_PAREN, RIGHT_BRACE, EOF},
			[]string{
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: %",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantErrs) == 0 {
				assertScanTokenTypes(t, tt.source, tt.want)
				return
			}
			assertScanTokenTypesAndErrors(t, tt.source, tt.want, tt.wantErrs)
		})
	}
}

func TestScannerRelationalOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"greater equal", ">=", []TokenType{GREATER_EQUAL, EOF}},
		{"mixed less and greater", "<<<=>>>=", []TokenType{LESS, LESS, LESS_EQUAL, GREATER, GREATER, GREATER_EQUAL, EOF}},
		{"alternating relational operators", ">=><><=", []TokenType{GREATER_EQUAL, GREATER, LESS, GREATER, LESS_EQUAL, EOF}},
		{"relational neighbors", "(){===!}", []TokenType{LEFT_PAREN, RIGHT_PAREN, LEFT_BRACE, EQUAL_EQUAL, EQUAL, BANG, RIGHT_BRACE, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerDivisionOperatorAndComments(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"comment only", "//Comment", []TokenType{EOF}},
		{"comment after paren", "(///Unicode:£§᯽☺♣)", []TokenType{LEFT_PAREN, EOF}},
		{"single slash token", "/", []TokenType{SLASH, EOF}},
		{"operators before comment", "({(!=!*)})//Comment", []TokenType{LEFT_PAREN, LEFT_BRACE, LEFT_PAREN, BANG_EQUAL, BANG, STAR, RIGHT_PAREN, RIGHT_BRACE, RIGHT_PAREN, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerWhitespace(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"single space", " ", []TokenType{EOF}},
		{"mixed spaces tabs newline", " \t\n ", []TokenType{EOF}},
		{"whitespace around punctuation", "{\n\t}\n((-,+\n ))", []TokenType{LEFT_BRACE, RIGHT_BRACE, LEFT_PAREN, LEFT_PAREN, MINUS, COMMA, PLUS, RIGHT_PAREN, RIGHT_PAREN, EOF}},
		{"whitespace with relational ops", "{  \t\t\n}\n((<>.<=*))", []TokenType{LEFT_BRACE, RIGHT_BRACE, LEFT_PAREN, LEFT_PAREN, LESS, GREATER, DOT, LESS_EQUAL, STAR, RIGHT_PAREN, RIGHT_PAREN, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerStringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		want     []TokenType
		wantErrs []string
	}{
		{"simple string", "\"hello\"", []TokenType{STRING, EOF}, nil},
		{"unterminated string", "\"hello\" , \"unterminated", []TokenType{STRING, COMMA, EOF}, []string{"[line 1] Error: Unterminated string."}},
		{"string with tab and slashes", "\"foo \tbar 123 // hello world!\"", []TokenType{STRING, EOF}, nil},
		{"strings in expression", "(\"foo\"+\"world\") != \"other_string\"", []TokenType{LEFT_PAREN, STRING, PLUS, STRING, RIGHT_PAREN, BANG_EQUAL, STRING, EOF}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantErrs) == 0 {
				assertScanTokenTypes(t, tt.source, tt.want)
				return
			}
			assertScanTokenTypesAndErrors(t, tt.source, tt.want, tt.wantErrs)
		})
	}
}

func TestScannerNumberLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"integer", "16", []TokenType{NUMBER, EOF}},
		{"fractional", "1752.8717", []TokenType{NUMBER, EOF}},
		{"fractional with trailing zeros", "65.0000", []TokenType{NUMBER, EOF}},
		{"numbers in complex expression", "(25+11) > 36 != (\"Success\" != \"Failure\") != (36 >= 70)", []TokenType{LEFT_PAREN, NUMBER, PLUS, NUMBER, RIGHT_PAREN, GREATER, NUMBER, BANG_EQUAL, LEFT_PAREN, STRING, BANG_EQUAL, STRING, RIGHT_PAREN, BANG_EQUAL, LEFT_PAREN, NUMBER, GREATER_EQUAL, NUMBER, RIGHT_PAREN, EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerIdentifiers(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"simple identifiers", "baz bar", []TokenType{IDENTIFIER, IDENTIFIER, EOF}},
		{"underscore and digits", "_1236ar foo world_ baz f00", []TokenType{IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, EOF}},
		{"identifiers with assignments", "message = \"Hello, World!\"\nnumber = 123", []TokenType{IDENTIFIER, EQUAL, STRING, IDENTIFIER, EQUAL, NUMBER, EOF}},
		{
			"complex identifiers in expression",
			"{\n// This is a complex test case\nstr1 = \"Test\"\nstr2 = \"Case\"\nnum1 = 100\nnum2 = 200.00\nresult = (str1 == str2) != ((num1 + num2) >= 300)\n}",
			[]TokenType{
				LEFT_BRACE, IDENTIFIER, EQUAL, STRING, IDENTIFIER, EQUAL, STRING, IDENTIFIER, EQUAL, NUMBER,
				IDENTIFIER, EQUAL, NUMBER, IDENTIFIER, EQUAL, LEFT_PAREN, IDENTIFIER, EQUAL_EQUAL, IDENTIFIER,
				RIGHT_PAREN, BANG_EQUAL, LEFT_PAREN, LEFT_PAREN, IDENTIFIER, PLUS, IDENTIFIER, RIGHT_PAREN,
				GREATER_EQUAL, NUMBER, RIGHT_PAREN, RIGHT_BRACE, EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}

func TestScannerReservedWords(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []TokenType
	}{
		{"single reserved word", "else", []TokenType{ELSE, EOF}},
		{
			"reserved and uppercase identifiers",
			"nil true print class this ELSE AND WHILE FALSE while or CLASS VAR var NIL if FOR super IF FUN and OR TRUE SUPER for fun PRINT RETURN false else return THIS",
			[]TokenType{
				NIL, TRUE, PRINT, CLASS, THIS, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, WHILE, OR, IDENTIFIER,
				IDENTIFIER, VAR, IDENTIFIER, IF, IDENTIFIER, SUPER, IDENTIFIER, IDENTIFIER, AND, IDENTIFIER, IDENTIFIER,
				IDENTIFIER, FOR, FUN, IDENTIFIER, IDENTIFIER, FALSE, ELSE, RETURN, IDENTIFIER, EOF,
			},
		},
		{
			"reserved words in if else",
			"var greeting = \"Hello\"\nif (greeting == \"Hello\") {\n    return true\n} else {\n    return false\n}",
			[]TokenType{VAR, IDENTIFIER, EQUAL, STRING, IF, LEFT_PAREN, IDENTIFIER, EQUAL_EQUAL, STRING, RIGHT_PAREN, LEFT_BRACE, RETURN, TRUE, RIGHT_BRACE, ELSE, LEFT_BRACE, RETURN, FALSE, RIGHT_BRACE, EOF},
		},
		{
			"reserved words in loop and condition",
			"var result = (a + b) > 7 or \"Success\" != \"Failure\" or x >= 5\nwhile (result) {\n    var counter = 0\n    counter = counter + 1\n    if (counter == 10) {\n        return nil\n    }\n}",
			[]TokenType{
				VAR, IDENTIFIER, EQUAL, LEFT_PAREN, IDENTIFIER, PLUS, IDENTIFIER, RIGHT_PAREN, GREATER, NUMBER, OR,
				STRING, BANG_EQUAL, STRING, OR, IDENTIFIER, GREATER_EQUAL, NUMBER, WHILE, LEFT_PAREN, IDENTIFIER,
				RIGHT_PAREN, LEFT_BRACE, VAR, IDENTIFIER, EQUAL, NUMBER, IDENTIFIER, EQUAL, IDENTIFIER, PLUS, NUMBER,
				IF, LEFT_PAREN, IDENTIFIER, EQUAL_EQUAL, NUMBER, RIGHT_PAREN, LEFT_BRACE, RETURN, NIL, RIGHT_BRACE,
				RIGHT_BRACE, EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}
