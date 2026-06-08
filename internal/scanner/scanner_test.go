package scanner

import (
	"strings"
	"testing"

	"github.com/nt54hamnghi/golox/internal/scanner/token"
	"github.com/stretchr/testify/require"
)

func TestScannerLexicalErrors(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantErrs   []string
		wantTokens []token.TokenType
	}{
		{
			name:       "single unexpected character",
			source:     "@",
			wantErrs:   []string{"[line 1] Error: Unexpected character: @"},
			wantTokens: []token.TokenType{token.EOF},
		},
		{
			name:   "mixed valid and invalid characters",
			source: ",.$(#",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: $",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []token.TokenType{token.COMMA, token.DOT, token.LEFT_PAREN, token.EOF},
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
			wantTokens: []token.TokenType{token.EOF},
		},
		{
			name:   "valid tokens continue around lexical errors",
			source: "{(+.@-$;#)}",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: $",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []token.TokenType{token.LEFT_BRACE, token.LEFT_PAREN, token.PLUS, token.DOT, token.MINUS, token.SEMICOLON, token.RIGHT_PAREN, token.RIGHT_BRACE, token.EOF},
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
		wantTokens []token.TokenType
	}{
		{
			name:       "error on second line after valid first line",
			source:     "()\n\t@",
			wantErrs:   []string{"[line 2] Error: Unexpected character: @"},
			wantTokens: []token.TokenType{token.LEFT_PAREN, token.RIGHT_PAREN, token.EOF},
		},
		{
			name:   "multiple unexpected characters on same line",
			source: " @%#",
			wantErrs: []string{
				"[line 1] Error: Unexpected character: @",
				"[line 1] Error: Unexpected character: %",
				"[line 1] Error: Unexpected character: #",
			},
			wantTokens: []token.TokenType{token.EOF},
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
			wantTokens: []token.TokenType{
				token.LEFT_PAREN,
				token.RIGHT_PAREN,
				token.LEFT_BRACE,
				token.RIGHT_BRACE,
				token.PLUS,
				token.PLUS,
				token.PLUS,
				token.PLUS,
				token.PLUS,
				token.PLUS,
				token.EOF,
			},
		},
		{
			name:       "newline splits valid punctuation and error",
			source:     "({;\n$})",
			wantErrs:   []string{"[line 2] Error: Unexpected character: $"},
			wantTokens: []token.TokenType{token.LEFT_PAREN, token.LEFT_BRACE, token.SEMICOLON, token.RIGHT_BRACE, token.RIGHT_PAREN, token.EOF},
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
	r.Equal(token.EOF, tokens[0].Type)
}

func assertScanTokenTypes(t *testing.T, source string, want []token.TokenType) {
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

func assertScanTokenTypesAndErrors(t *testing.T, source string, want []token.TokenType, wantErrs []string) {
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
		want   []token.TokenType
	}{
		{"single left paren", "(", []token.TokenType{token.LEFT_PAREN, token.EOF}},
		{"double right paren", "))", []token.TokenType{token.RIGHT_PAREN, token.RIGHT_PAREN, token.EOF}},
		{"mixed right and left", ")))((", []token.TokenType{token.RIGHT_PAREN, token.RIGHT_PAREN, token.RIGHT_PAREN, token.LEFT_PAREN, token.LEFT_PAREN, token.EOF}},
		{"nested sequence", "()((())", []token.TokenType{token.LEFT_PAREN, token.RIGHT_PAREN, token.LEFT_PAREN, token.LEFT_PAREN, token.LEFT_PAREN, token.RIGHT_PAREN, token.RIGHT_PAREN, token.EOF}},
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
		want   []token.TokenType
	}{
		{"single right brace", "}", []token.TokenType{token.RIGHT_BRACE, token.EOF}},
		{"pair of braces", "{{}}", []token.TokenType{token.LEFT_BRACE, token.LEFT_BRACE, token.RIGHT_BRACE, token.RIGHT_BRACE, token.EOF}},
		{"alternating braces", "}{}{{", []token.TokenType{token.RIGHT_BRACE, token.LEFT_BRACE, token.RIGHT_BRACE, token.LEFT_BRACE, token.LEFT_BRACE, token.EOF}},
		{"mixed braces and parens", "{(){()}", []token.TokenType{token.LEFT_BRACE, token.LEFT_PAREN, token.RIGHT_PAREN, token.LEFT_BRACE, token.LEFT_PAREN, token.RIGHT_PAREN, token.RIGHT_BRACE, token.EOF}},
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
		want   []token.TokenType
	}{
		{"plus minus", "+-", []token.TokenType{token.PLUS, token.MINUS, token.EOF}},
		{"all single character punctuators", "++--**..,,;;", []token.TokenType{token.PLUS, token.PLUS, token.MINUS, token.MINUS, token.STAR, token.STAR, token.DOT, token.DOT, token.COMMA, token.COMMA, token.SEMICOLON, token.SEMICOLON, token.EOF}},
		{"mixed punctuation order", "-+*,+*;", []token.TokenType{token.MINUS, token.PLUS, token.STAR, token.COMMA, token.PLUS, token.STAR, token.SEMICOLON, token.EOF}},
		{"single character tokens in grouping", "({*,+-.})", []token.TokenType{token.LEFT_PAREN, token.LEFT_BRACE, token.STAR, token.COMMA, token.PLUS, token.MINUS, token.DOT, token.RIGHT_BRACE, token.RIGHT_PAREN, token.EOF}},
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
		want     []token.TokenType
		wantErrs []string
	}{
		{"single equal", "=", []token.TokenType{token.EQUAL, token.EOF}, nil},
		{"double equal", "==", []token.TokenType{token.EQUAL_EQUAL, token.EOF}, nil},
		{"grouped equal operators", "({=}){==}", []token.TokenType{token.LEFT_PAREN, token.LEFT_BRACE, token.EQUAL, token.RIGHT_BRACE, token.RIGHT_PAREN, token.LEFT_BRACE, token.EQUAL_EQUAL, token.RIGHT_BRACE, token.EOF}, nil},
		{
			"operators mixed with lexical errors",
			"((==#%=$))",
			[]token.TokenType{token.LEFT_PAREN, token.LEFT_PAREN, token.EQUAL_EQUAL, token.EQUAL, token.RIGHT_PAREN, token.RIGHT_PAREN, token.EOF},
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
		want     []token.TokenType
		wantErrs []string
	}{
		{"bang equal", "!=", []token.TokenType{token.BANG_EQUAL, token.EOF}, nil},
		{"bang and equality chain", "!!===", []token.TokenType{token.BANG, token.BANG_EQUAL, token.EQUAL_EQUAL, token.EOF}, nil},
		{"bang operators with grouping", "!{!}(!===)=", []token.TokenType{token.BANG, token.LEFT_BRACE, token.BANG, token.RIGHT_BRACE, token.LEFT_PAREN, token.BANG_EQUAL, token.EQUAL_EQUAL, token.RIGHT_PAREN, token.EQUAL, token.EOF}, nil},
		{
			"unexpected chars among bang tokens",
			"{(!==@%!)}",
			[]token.TokenType{token.LEFT_BRACE, token.LEFT_PAREN, token.BANG_EQUAL, token.EQUAL, token.BANG, token.RIGHT_PAREN, token.RIGHT_BRACE, token.EOF},
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
		want   []token.TokenType
	}{
		{"greater equal", ">=", []token.TokenType{token.GREATER_EQUAL, token.EOF}},
		{"mixed less and greater", "<<<=>>>=", []token.TokenType{token.LESS, token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER, token.GREATER_EQUAL, token.EOF}},
		{"alternating relational operators", ">=><><=", []token.TokenType{token.GREATER_EQUAL, token.GREATER, token.LESS, token.GREATER, token.LESS_EQUAL, token.EOF}},
		{"relational neighbors", "(){===!}", []token.TokenType{token.LEFT_PAREN, token.RIGHT_PAREN, token.LEFT_BRACE, token.EQUAL_EQUAL, token.EQUAL, token.BANG, token.RIGHT_BRACE, token.EOF}},
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
		want   []token.TokenType
	}{
		{"comment only", "//Comment", []token.TokenType{token.EOF}},
		{"comment after paren", "(///Unicode:£§᯽☺♣)", []token.TokenType{token.LEFT_PAREN, token.EOF}},
		{"single slash token", "/", []token.TokenType{token.SLASH, token.EOF}},
		{"operators before comment", "({(!=!*)})//Comment", []token.TokenType{token.LEFT_PAREN, token.LEFT_BRACE, token.LEFT_PAREN, token.BANG_EQUAL, token.BANG, token.STAR, token.RIGHT_PAREN, token.RIGHT_BRACE, token.RIGHT_PAREN, token.EOF}},
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
		want   []token.TokenType
	}{
		{"single space", " ", []token.TokenType{token.EOF}},
		{"mixed spaces tabs newline", " \t\n ", []token.TokenType{token.EOF}},
		{"whitespace around punctuation", "{\n\t}\n((-,+\n ))", []token.TokenType{token.LEFT_BRACE, token.RIGHT_BRACE, token.LEFT_PAREN, token.LEFT_PAREN, token.MINUS, token.COMMA, token.PLUS, token.RIGHT_PAREN, token.RIGHT_PAREN, token.EOF}},
		{"whitespace with relational ops", "{  \t\t\n}\n((<>.<=*))", []token.TokenType{token.LEFT_BRACE, token.RIGHT_BRACE, token.LEFT_PAREN, token.LEFT_PAREN, token.LESS, token.GREATER, token.DOT, token.LESS_EQUAL, token.STAR, token.RIGHT_PAREN, token.RIGHT_PAREN, token.EOF}},
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
		want     []token.TokenType
		wantErrs []string
	}{
		{"simple string", "\"hello\"", []token.TokenType{token.STRING, token.EOF}, nil},
		{"unterminated string", "\"hello\" , \"unterminated", []token.TokenType{token.STRING, token.COMMA, token.EOF}, []string{"[line 1] Error: Unterminated string."}},
		{"string with tab and slashes", "\"foo \tbar 123 // hello world!\"", []token.TokenType{token.STRING, token.EOF}, nil},
		{"strings in expression", "(\"foo\"+\"world\") != \"other_string\"", []token.TokenType{token.LEFT_PAREN, token.STRING, token.PLUS, token.STRING, token.RIGHT_PAREN, token.BANG_EQUAL, token.STRING, token.EOF}, nil},
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
		want   []token.TokenType
	}{
		{"integer", "16", []token.TokenType{token.NUMBER, token.EOF}},
		{"fractional", "1752.8717", []token.TokenType{token.NUMBER, token.EOF}},
		{"fractional with trailing zeros", "65.0000", []token.TokenType{token.NUMBER, token.EOF}},
		{"numbers in complex expression", "(25+11) > 36 != (\"Success\" != \"Failure\") != (36 >= 70)", []token.TokenType{token.LEFT_PAREN, token.NUMBER, token.PLUS, token.NUMBER, token.RIGHT_PAREN, token.GREATER, token.NUMBER, token.BANG_EQUAL, token.LEFT_PAREN, token.STRING, token.BANG_EQUAL, token.STRING, token.RIGHT_PAREN, token.BANG_EQUAL, token.LEFT_PAREN, token.NUMBER, token.GREATER_EQUAL, token.NUMBER, token.RIGHT_PAREN, token.EOF}},
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
		want   []token.TokenType
	}{
		{"simple identifiers", "baz bar", []token.TokenType{token.IDENTIFIER, token.IDENTIFIER, token.EOF}},
		{"underscore and digits", "_1236ar foo world_ baz f00", []token.TokenType{token.IDENTIFIER, token.IDENTIFIER, token.IDENTIFIER, token.IDENTIFIER, token.IDENTIFIER, token.EOF}},
		{"identifiers with assignments", "message = \"Hello, World!\"\nnumber = 123", []token.TokenType{token.IDENTIFIER, token.EQUAL, token.STRING, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.EOF}},
		{
			"complex identifiers in expression",
			"{\n// This is a complex test case\nstr1 = \"Test\"\nstr2 = \"Case\"\nnum1 = 100\nnum2 = 200.00\nresult = (str1 == str2) != ((num1 + num2) >= 300)\n}",
			[]token.TokenType{
				token.LEFT_BRACE, token.IDENTIFIER, token.EQUAL, token.STRING, token.IDENTIFIER, token.EQUAL, token.STRING, token.IDENTIFIER, token.EQUAL, token.NUMBER,
				token.IDENTIFIER, token.EQUAL, token.NUMBER, token.IDENTIFIER, token.EQUAL, token.LEFT_PAREN, token.IDENTIFIER, token.EQUAL_EQUAL, token.IDENTIFIER,
				token.RIGHT_PAREN, token.BANG_EQUAL, token.LEFT_PAREN, token.LEFT_PAREN, token.IDENTIFIER, token.PLUS, token.IDENTIFIER, token.RIGHT_PAREN,
				token.GREATER_EQUAL, token.NUMBER, token.RIGHT_PAREN, token.RIGHT_BRACE, token.EOF,
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
		want   []token.TokenType
	}{
		{"single reserved word", "else", []token.TokenType{token.ELSE, token.EOF}},
		{
			"reserved and uppercase identifiers",
			"nil true print class this ELSE AND WHILE FALSE while or CLASS VAR var NIL if FOR super IF FUN and OR TRUE SUPER for fun PRINT RETURN false else return THIS",
			[]token.TokenType{
				token.NIL, token.TRUE, token.PRINT, token.CLASS, token.THIS, token.IDENTIFIER, token.IDENTIFIER, token.IDENTIFIER, token.IDENTIFIER, token.WHILE, token.OR, token.IDENTIFIER,
				token.IDENTIFIER, token.VAR, token.IDENTIFIER, token.IF, token.IDENTIFIER, token.SUPER, token.IDENTIFIER, token.IDENTIFIER, token.AND, token.IDENTIFIER, token.IDENTIFIER,
				token.IDENTIFIER, token.FOR, token.FUN, token.IDENTIFIER, token.IDENTIFIER, token.FALSE, token.ELSE, token.RETURN, token.IDENTIFIER, token.EOF,
			},
		},
		{
			"reserved words in if else",
			"var greeting = \"Hello\"\nif (greeting == \"Hello\") {\n    return true\n} else {\n    return false\n}",
			[]token.TokenType{token.VAR, token.IDENTIFIER, token.EQUAL, token.STRING, token.IF, token.LEFT_PAREN, token.IDENTIFIER, token.EQUAL_EQUAL, token.STRING, token.RIGHT_PAREN, token.LEFT_BRACE, token.RETURN, token.TRUE, token.RIGHT_BRACE, token.ELSE, token.LEFT_BRACE, token.RETURN, token.FALSE, token.RIGHT_BRACE, token.EOF},
		},
		{
			"reserved words in loop and condition",
			"var result = (a + b) > 7 or \"Success\" != \"Failure\" or x >= 5\nwhile (result) {\n    var counter = 0\n    counter = counter + 1\n    if (counter == 10) {\n        return nil\n    }\n}",
			[]token.TokenType{
				token.VAR, token.IDENTIFIER, token.EQUAL, token.LEFT_PAREN, token.IDENTIFIER, token.PLUS, token.IDENTIFIER, token.RIGHT_PAREN, token.GREATER, token.NUMBER, token.OR,
				token.STRING, token.BANG_EQUAL, token.STRING, token.OR, token.IDENTIFIER, token.GREATER_EQUAL, token.NUMBER, token.WHILE, token.LEFT_PAREN, token.IDENTIFIER,
				token.RIGHT_PAREN, token.LEFT_BRACE, token.VAR, token.IDENTIFIER, token.EQUAL, token.NUMBER, token.IDENTIFIER, token.EQUAL, token.IDENTIFIER, token.PLUS, token.NUMBER,
				token.IF, token.LEFT_PAREN, token.IDENTIFIER, token.EQUAL_EQUAL, token.NUMBER, token.RIGHT_PAREN, token.LEFT_BRACE, token.RETURN, token.NIL, token.RIGHT_BRACE,
				token.RIGHT_BRACE, token.EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScanTokenTypes(t, tt.source, tt.want)
		})
	}
}
