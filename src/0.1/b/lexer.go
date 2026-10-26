package b

import (
	"fmt"
)

type TokenType string

const (
	TOKEN_NUMBER  TokenType = "NUMBER"
	TOKEN_ID      TokenType = "ID"
	TOKEN_KEYWORD TokenType = "KEYWORD"
	TOKEN_OP      TokenType = "OP"
	TOKEN_SYMBOL  TokenType = "SYMBOL"
	TOKEN_EOF     TokenType = "EOF"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	src    string
	pos    int
	ch     byte
	line   int
	col    int
	tokens []Token
}

func NewLexer(src string) *Lexer {
	if len(src) >= 3 && src[0] == 0xEF && src[1] == 0xBB && src[2] == 0xBF {
		src = src[3:]
	}
	l := &Lexer{src: src, pos: 0, line: 1, col: 0, tokens: []Token{}}
	l.nextChar()
	return l
}

func (l *Lexer) nextChar() {
	if l.pos >= len(l.src) {
		l.ch = 0
		return
	}
	l.ch = l.src[l.pos]
	l.pos++
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

func (l *Lexer) peekChar() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) Lex() []Token {
	for l.ch != 0 {
		l.skipWhitespace()
		if l.ch == 0 {
			break
		}
		switch {
		case l.isLetter():
			l.lexIdentifier()
		case l.isDigit():
			l.lexNumber()
		case l.ch == '/':
			l.lexCommentOrOp()
		default:
			l.lexSymbolOrOp()
		}
	}
	l.tokens = append(l.tokens, Token{TOKEN_EOF, "", l.line, l.col})
	return l.tokens
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.nextChar()
	}
}

func (l *Lexer) isLetter() bool {
	return (l.ch >= 'a' && l.ch <= 'z') || (l.ch >= 'A' && l.ch <= 'Z') || l.ch == '_'
}

func (l *Lexer) isDigit() bool {
	return l.ch >= '0' && l.ch <= '9'
}

func (l *Lexer) isLetterOrDigit() bool {
	return l.isLetter() || l.isDigit()
}

func (l *Lexer) lexNumber() {
	startPos := l.pos - 1
	startCol := l.col
	for l.isDigit() {
		l.nextChar()
	}
	l.tokens = append(l.tokens, Token{TOKEN_NUMBER, l.src[startPos : l.pos-1], l.line, startCol})
}

func (l *Lexer) lexIdentifier() {
	startPos := l.pos - 1
	startCol := l.col
	for l.isLetterOrDigit() {
		l.nextChar()
	}
	value := l.src[startPos : l.pos-1]
	if isKeyword(value) {
		l.tokens = append(l.tokens, Token{TOKEN_KEYWORD, value, l.line, startCol})
	} else {
		l.tokens = append(l.tokens, Token{TOKEN_ID, value, l.line, startCol})
	}
}

func (l *Lexer) lexCommentOrOp() {
	if l.peekChar() == '*' {
		l.skipComment()
		return
	}
	l.tokens = append(l.tokens, Token{TOKEN_OP, "/", l.line, l.col})
	l.nextChar()
}

func (l *Lexer) skipComment() {
	l.nextChar()
	l.nextChar()
	for {
		if l.ch == 0 {
			panic(fmt.Sprintf("未闭合的注释 at line %d", l.line))
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.nextChar()
			l.nextChar()
			return
		}
		l.nextChar()
	}
}

func (l *Lexer) lexSymbolOrOp() {
	startCol := l.col
	switch l.ch {
	case '+', '-', '*':
		l.tokens = append(l.tokens, Token{TOKEN_OP, string(l.ch), l.line, startCol})
		l.nextChar()
	case '<':
		if l.peekChar() == '=' {
			l.tokens = append(l.tokens, Token{TOKEN_OP, "<=", l.line, startCol})
			l.nextChar()
			l.nextChar()
		} else {
			l.tokens = append(l.tokens, Token{TOKEN_OP, "<", l.line, startCol})
			l.nextChar()
		}
	case '>':
		if l.peekChar() == '=' {
			l.tokens = append(l.tokens, Token{TOKEN_OP, ">=", l.line, startCol})
			l.nextChar()
			l.nextChar()
		} else {
			l.tokens = append(l.tokens, Token{TOKEN_OP, ">", l.line, startCol})
			l.nextChar()
		}
	case '=':
		if l.peekChar() == '=' {
			l.tokens = append(l.tokens, Token{TOKEN_OP, "==", l.line, startCol})
			l.nextChar()
			l.nextChar()
		} else {
			l.tokens = append(l.tokens, Token{TOKEN_OP, "=", l.line, startCol})
			l.nextChar()
		}
	case '!':
		if l.peekChar() == '=' {
			l.tokens = append(l.tokens, Token{TOKEN_OP, "!=", l.line, startCol})
			l.nextChar()
			l.nextChar()
		} else {
			panic(fmt.Sprintf("未知字符 '!' at line %d", l.line))
		}
	case '(', ')', '{', '}', ';', ',':
		l.tokens = append(l.tokens, Token{TOKEN_SYMBOL, string(l.ch), l.line, startCol})
		l.nextChar()
	case '\'':
		l.nextChar()
		if l.ch == 0 {
			panic("未闭合的字符常量")
		}
		var ascii int
		if l.ch == '\\' {
			l.nextChar()
			switch l.ch {
			case 'n':
				ascii = 10
			case 't':
				ascii = 9
			case 'r':
				ascii = 13
			case '0':
				ascii = 0
			case '\\':
				ascii = 92
			case '\'':
				ascii = 39
			case '"':
				ascii = 34
			default:
				panic(fmt.Sprintf("未知转义字符 '\\%c' at line %d", l.ch, l.line))
			}
			l.nextChar()
		} else {
			ascii = int(l.ch)
			l.nextChar()
		}
		if l.ch != '\'' {
			panic(fmt.Sprintf("期望 '，但得到 '%c' at line %d", l.ch, l.line))
		}
		l.nextChar()
		l.tokens = append(l.tokens, Token{TOKEN_NUMBER, fmt.Sprintf("%d", ascii), l.line, startCol})
	default:
		panic(fmt.Sprintf("未知字符 '%c' at line %d", l.ch, l.line))
	}
}

func isKeyword(word string) bool {
	switch word {
	case "auto", "if", "while", "return":
		return true
	default:
		return false
	}
}
