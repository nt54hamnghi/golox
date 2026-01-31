package main

type Scanner struct {
	// Raw source code
	source []rune
	// Slice of tokens to be filled as we scan the source code
	tokens []Token
	// Offset into the source code, pointing at the first character of the lexeme being scanned
	start int
	// Offset into the source code, pointing at the character being considered for the current lexeme
	current int
	// Line where the lexeme is located.
	line int
}

func NewScanner(src string) Scanner {
	return Scanner{
		source:  []rune(src),
		tokens:  []Token{},
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) scanTokens() []Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}
	s.tokens = append(s.tokens, NewToken(EOF, "", nil, s.line))
	return s.tokens
}

func (s *Scanner) scanToken() {
	switch char := s.advanced(); char {
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '.':
		s.addToken(DOT, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case ';':
		s.addToken(SEMICOLON, nil)
	case '*':
		s.addToken(STAR, nil)
	case '!':
		var typ TokenType
		if s.match('=') {
			typ = BANG_EQUAL
		} else {
			typ = BANG
		}
		s.addToken(typ, nil)
	case '=':
		var typ TokenType
		if s.match('=') {
			typ = EQUAL_EQUAL
		} else {
			typ = EQUAL
		}
		s.addToken(typ, nil)
	case '<':
		var typ TokenType
		if s.match('=') {
			typ = LESS_EQUAL
		} else {
			typ = LESS
		}
		s.addToken(typ, nil)
	case '>':
		var typ TokenType
		if s.match('=') {
			typ = GREATER_EQUAL
		} else {
			typ = GREATER
		}
		s.addToken(typ, nil)
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advanced()
			}
		} else {
			s.addToken(SLASH, nil)
		}
	case ' ', '\r', '\t':
		// ignore whitespace
		return
	case '\n':
		s.line++
		return
	case '"':
		s.stringLiteral()
	default:
		err(s.line, "Unexpected characters.")
	}
}

func (s *Scanner) advanced() rune {
	char := s.source[s.current]
	s.current++
	return char
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	char := s.source[s.current]
	if char != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0 // null character
	}
	return s.source[s.current]
}

func (s *Scanner) stringLiteral() {
	for s.peek() != '"' && !s.isAtEnd() {
		// support for multi-line string, updating line
		// counter when encountering a newline
		if s.peek() == '\n' {
			s.line += 1
		}
		s.advanced()
	}

	if s.isAtEnd() {
		err(s.line, "Unterminated string")
		return
	}

	// consume the closing "
	s.advanced()

	// trim the surrounding quotes
	value := s.source[s.start+1 : s.current-1]
	s.addToken(STRING, string(value))
}

func (s *Scanner) addToken(typ TokenType, literal any) {
	text := string(s.source[s.start:s.current])
	token := NewToken(typ, text, literal, s.line)
	s.tokens = append(s.tokens, token)
}

func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}
