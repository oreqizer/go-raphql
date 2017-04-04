package language

import (
	"fmt"
	"unicode/utf8"
)

const (
	EOF = iota + 1
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
)

type Token struct {
	Kind int
	Start int
	End int
	Value string
}

type Body []byte

func readToken(body Body, position int) (Token, error) {
	bytepos, runepos := skipIgnored(body, position)
	code, n := runeAt(body, bytepos)
	if code == EOF_ERR {
		return Token{EOF, bytepos, bytepos, ""}, nil
	}

	if isSource(code) {
		return Token{}, fmt.Errorf(`Invalid character %c`, code)
	}

	switch code {
	case '!':
		return Token{BANG, bytepos, bytepos+1, ""}, nil

	case '$':
		return Token{DOLLAR, bytepos, bytepos+1, ""}, nil

	case '(':
		return Token{PAREN_L, bytepos, bytepos+1, ""}, nil

	case ')':
		return Token{PAREN_R, bytepos, bytepos+1, ""}, nil

	case '.':
		if isSpread(body, bytepos) {
			return Token{SPREAD, bytepos, bytepos+3, ""}, nil
		}
		return Token{}, fmt.Errorf(`Invalid character %c`, code)

	case ':':
		return Token{COLON, bytepos, bytepos+1, ""}, nil

	case '=':
		return Token{EQUALS, bytepos, bytepos+1, ""}, nil

	case '@':
		return Token{AT, bytepos, bytepos+1, ""}, nil

	case '[':
		return Token{BRACKET_L, bytepos, bytepos+1, ""}, nil

	case ']':
		return Token{BRACKET_R, bytepos, bytepos+1, ""}, nil

	case '{':
		return Token{BRACE_L, bytepos, bytepos+1, ""}, nil

	case '|':
		return Token{PIPE, bytepos, bytepos+1, ""}, nil

	case '}':
		return Token{BRACE_R, bytepos, bytepos+1, ""}, nil

	case '"':
		return readString(body, bytepos)
	}

	// TODO readName, readNumber, readString
	return Token{}, fmt.Errorf(`Invalid character %c`, code)
}

const (
	EOF_ERR = utf8.RuneError
	COMMENT = 35
	BOM     = 0xFEFF
	TAB     = 0x0009
	SPACE   = 0x0020
	NEWLINE = 0x000A
	CR      = 0x000D
	COMMA   = 0x002C
)

var ignoredRunes = map[rune]struct{}{
	BOM:     {},
	TAB:     {},
	SPACE:   {},
	NEWLINE: {},
	CR:      {},
	COMMA:   {},
}

func skipIgnored(body Body, start int) (bytepos, runepos int) {
	bytepos = start
	runepos = start
	for code, n := runeAt(body, bytepos); code != EOF_ERR; {
		if _, ok := ignoredRunes[code]; ok {
			bytepos += n
			runepos++
			continue
		}

		if code == COMMENT {
			bytepos += n
			runepos++
			// Ignore comment
			for code, n := runeAt(body, bytepos); code != EOF_ERR && isCommented(code); {
				bytepos += n
				runepos++
			}
			continue
		}

		break
	}

	return bytepos, runepos
}

func readString(body Body, pos int) (Token, error) {
	return Token{}, nil  // TODO
}

// Utils
// ---

func runeAt(body Body, pos int) (rune, int) {
	if len(body) <= pos {
		return EOF_ERR, 0
	}
	return utf8.DecodeRune(body[pos:])
}


func isSource(r rune) bool {
	return r < SPACE && r != TAB && r != NEWLINE && r != CR
}

func isCommented(r rune) bool {
	return r != 0 && r != NEWLINE && r != CR && (r >= SPACE || r == TAB)
}

func isSpread(body Body, pos int) bool {
	next1, _ := runeAt(body, pos+1)
	next2, _ := runeAt(body, pos+2)
	return next1 == '.' && next2 == '.'
}
