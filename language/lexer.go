// Reference implementation in JS:
// https://github.com/graphql/graphql-js/blob/master/src/language/lexer.js
package language

import (
	"fmt"
	"unicode/utf8"
)

const (
	SOF = iota + 1
	EOF
	BANG
	DOLLAR
	PAREN_L
	PAREN_R
	SPREAD
	COLON
	EQUALS
	AT
	BRACKET_L
	BRACKET_R
	BRACE_L
	PIPE
	BRACE_R
	NAME
	INT
	FLOAT
	STRING
	COMMENT
)

var TokenDisplay = map[int]string{
	SOF: "<SOF>",
	EOF: "<EOF>",
	BANG: "!",
	DOLLAR: "$",
	PAREN_L: "(",
	PAREN_R: ")",
	SPREAD: "...",
	COLON: ":",
	EQUALS: "=",
	AT: "@",
	BRACKET_L: "[",
	BRACKET_R: "]",
	BRACE_L: "{",
	PIPE: "|",
	BRACE_R: "}",
	NAME: "Name",
	INT: "Int",
	FLOAT: "Float",
	STRING: "String",
	COMMENT: "Comment",
}

type Body []byte

type Lexer struct {
	Body Body
	LastToken *Token
	Token *Token
	Line int
	LineStart int
}

func NewLexer(body Body) *Lexer {
	sofToken := &Token{Kind: SOF}
	return &Lexer{
		Body: body,
		LastToken: sofToken,
		Token: sofToken,
		Line: 1,
		LineStart: 0,
	}
}

func (l *Lexer) Advance() *Token {
	l.LastToken = l.Token
	token := l.Token
	if token.Kind != EOF {
		for {
			token, _ := readToken(l, token) // TODO handle error
			token.Next = token
			if token.Kind != COMMENT {
				break
			}
		}
	}
	l.Token = token
	return token
}

type Token struct {
	Kind int
	Start int
	End int
	Line int
	Column int
	Prev *Token
	Next *Token
	Value string
}

func readToken(lexer *Lexer, prev *Token) (*Token, error) {
	body := lexer.Body
	length := len(body)

	position := skipWhitespace(lexer, prev.End)
	line := lexer.Line
	col := 1 + position - lexer.LineStart

	code, _ := runeAt(body, position)
	if code == RUNE_EOF {
		return &Token{EOF, length, length, line, col, prev, nil, ""}, nil
	}

	if isSource(code) {
		return nil, fmt.Errorf(`Invalid character %c`, code)
	}

	switch code {
	case '!':
		return &Token{BANG, position, position+1, line, col, prev, nil, ""}, nil

	case '$':
		return &Token{DOLLAR, position, position+1, line, col, prev, nil, ""}, nil

	case '(':
		return &Token{PAREN_L, position, position+1, line, col, prev, nil, ""}, nil

	case ')':
		return &Token{PAREN_R, position, position+1, line, col, prev, nil, ""}, nil

	case '.':
		if isSpread(body, position) {
			return &Token{SPREAD, position, position+3, line, col, prev, nil, ""}, nil
		}
		return &Token{}, fmt.Errorf(`Invalid character %c`, code)

	case ':':
		return &Token{COLON, position, position+1, line, col, prev, nil, ""}, nil

	case '=':
		return &Token{EQUALS, position, position+1, line, col, prev, nil, ""}, nil

	case '@':
		return &Token{AT, position, position+1, line, col, prev, nil, ""}, nil

	case '[':
		return &Token{BRACKET_L, position, position+1, line, col, prev, nil, ""}, nil

	case ']':
		return &Token{BRACKET_R, position, position+1, line, col, prev, nil, ""}, nil

	case '{':
		return &Token{BRACE_L, position, position+1, line, col, prev, nil, ""}, nil

	case '|':
		return &Token{PIPE, position, position+1, line, col, prev, nil, ""}, nil

	case '}':
		return &Token{BRACE_R, position, position+1, line, col, prev, nil, ""}, nil

	case '"':
		return readString(body, position)
	}

	// TODO readName, readNumber, readString
	return &Token{}, fmt.Errorf(`Invalid character %c`, code)
}

const (
	RUNE_EOF     = utf8.RuneError
	RUNE_BOM     = 0xFEFF
	RUNE_TAB     = 0x0009
	RUNE_SPACE   = 0x0020
	RUNE_NEWLINE = 0x000A
	RUNE_CR      = 0x000D
	RUNE_COMMA   = 0x002C
)

func readString(body Body, start, line, col int, prev *Token) (*Token, error) {
	return &Token{}, nil  // TODO
}

func readComment(body Body, start, line, col int, prev *Token) *Token {
	position := start
	for {
		code, n := runeAt(body, position)
		if !isComment(code) {
			break
		}
		position += n
	}
	return &Token{
		Kind: COMMENT,
		Start: start,
		End: position,
		Line: line,
		Column: col,
		Prev: prev,
		Next: nil,
		Value: string(body[start + 1:position]),
	}
}

// Utils
// ---

func skipWhitespace(lexer *Lexer, start int) (position int) {
	body := lexer.Body
	position = start
	for position < len(body) {
		code, _ := runeAt(body, position)
		if isSingleIgnored(code) {
			position++
		} else if code == RUNE_NEWLINE {
			position++
			lexer.Line++
			lexer.LineStart = position
		} else if code == RUNE_CR {
			if r, _ := runeAt(body, position); r == RUNE_NEWLINE {
				position += 2  // NL after CR
			} else {
				position++
			}
			lexer.Line++
			lexer.LineStart = position
		} else {
			// End of whitespace
			break
		}
	}
	return position
}

func runeAt(body Body, pos int) (rune, int) {
	if len(body) <= pos {
		return RUNE_EOF, 0
	}
	return utf8.DecodeRune(body[pos:])
}


func isSource(r rune) bool {
	return r < RUNE_SPACE && r != RUNE_TAB && r != RUNE_NEWLINE && r != RUNE_CR
}

func isComment(r rune) bool {
	return r != RUNE_EOF && (r >= RUNE_SPACE || r == RUNE_TAB)
}

func isSingleIgnored(r rune) bool {
	return r == RUNE_TAB || r == RUNE_SPACE || r == RUNE_COMMA || r == RUNE_BOM
}

func isSpread(body Body, pos int) bool {
	next1, _ := runeAt(body, pos+1)
	next2, _ := runeAt(body, pos+2)
	return next1 == '.' && next2 == '.'
}
