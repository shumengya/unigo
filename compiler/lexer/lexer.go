package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"unigo/compiler/diag"
	"unigo/compiler/token"
)

var keywords = map[string]token.Kind{
	"package":     token.Package,
	"import":      token.Import,
	"from":        token.From,
	"var":         token.Var,
	"const":       token.Const,
	"func":        token.Func,
	"if":          token.If,
	"else":        token.Else,
	"elif":        token.Elif,
	"while":       token.While,
	"for":         token.For,
	"switch":      token.Switch,
	"case":        token.Case,
	"default":     token.Default,
	"enum":        token.Enum,
	"struct":      token.Struct,
	"interface":   token.Interface,
	"newtype":     token.Newtype,
	"alias":       token.Alias,
	"return":      token.Return,
	"map":         token.Map,
	"true":        token.True,
	"false":       token.False,
	"nil":         token.Nil,
	"fallthrough": token.Fallthrough,
	"iota":        token.Iota,
	"class":       token.Class,
	"new":         token.New,
	"extends":     token.Extends,
	"implements":  token.Implements,
	"type":        token.TypeWord,
}

type Lexer struct {
	file string
	src  string
	pos  int
	line int
	col  int
	errs *diag.List
}

func New(file, src string, errs *diag.List) *Lexer {
	if strings.HasPrefix(src, "\uFEFF") {
		src = strings.TrimPrefix(src, "\uFEFF")
	}
	return &Lexer{file: file, src: src, line: 1, col: 1, errs: errs}
}

func (l *Lexer) Next() token.Token {
	space := l.skipGap()
	if l.pos >= len(l.src) {
		return token.Token{Kind: token.EOF, Pos: l.position(), SpaceBefore: space}
	}
	start := l.position()
	r := l.peek()
	switch {
	case r == '"':
		lit, ok := l.scanString()
		if !ok {
			return token.Token{Kind: token.Illegal, Lit: "字符串没有结束", Pos: start, SpaceBefore: space}
		}
		return token.Token{Kind: token.String, Lit: lit, Pos: start, SpaceBefore: space}
	case r == '`':
		lit, ok := l.scanRawString()
		if !ok {
			return token.Token{Kind: token.Illegal, Lit: "字符串没有结束", Pos: start, SpaceBefore: space}
		}
		return token.Token{Kind: token.String, Lit: lit, Pos: start, SpaceBefore: space}
	case r == '\'':
		l.advance()
		l.error(start, "不支持字符字面量")
		return token.Token{Kind: token.Illegal, Lit: "不支持字符字面量", Pos: start, SpaceBefore: space}
	case r >= '0' && r <= '9':
		kind, lit := l.scanNumber()
		return token.Token{Kind: kind, Lit: lit, Pos: start, SpaceBefore: space}
	case isIdentStart(r):
		lit := l.scanIdent()
		if k, ok := keywords[lit]; ok {
			return token.Token{Kind: k, Lit: lit, Pos: start, SpaceBefore: space}
		}
		return token.Token{Kind: token.Ident, Lit: lit, Pos: start, SpaceBefore: space}
	}

	if l.starts(":=") {
		l.advance()
		l.advance()
		l.error(start, "不支持 :=")
		return token.Token{Kind: token.Illegal, Lit: "不支持 :=", Pos: start, SpaceBefore: space}
	}
	if l.starts("/*") {
		l.advance()
		l.advance()
		l.error(start, "注释只有 //")
		return token.Token{Kind: token.Illegal, Lit: "注释只有 //", Pos: start, SpaceBefore: space}
	}

	two := ""
	if l.pos+1 < len(l.src) {
		two = l.src[l.pos : l.pos+2]
	}
	switch two {
	case "==":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Eql, Lit: "==", Pos: start, SpaceBefore: space}
	case "!=":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Neq, Lit: "!=", Pos: start, SpaceBefore: space}
	case "<=":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Leq, Lit: "<=", Pos: start, SpaceBefore: space}
	case ">=":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Geq, Lit: ">=", Pos: start, SpaceBefore: space}
	case "&&":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Land, Lit: "&&", Pos: start, SpaceBefore: space}
	case "||":
		l.advance()
		l.advance()
		return token.Token{Kind: token.Lor, Lit: "||", Pos: start, SpaceBefore: space}
	}

	l.advance()
	kind := token.Illegal
	lit := string(r)
	switch r {
	case '+':
		kind = token.Add
	case '-':
		kind = token.Sub
	case '*':
		kind = token.Mul
	case '/':
		kind = token.Quo
	case '%':
		kind = token.Rem
	case '=':
		kind = token.Assign
	case '<':
		kind = token.Lss
	case '>':
		kind = token.Gtr
	case '!':
		kind = token.Not
	case '&':
		kind = token.And
	case '(':
		kind = token.Lparen
	case ')':
		kind = token.Rparen
	case '{':
		kind = token.Lbrace
	case '}':
		kind = token.Rbrace
	case '[':
		kind = token.Lbrack
	case ']':
		kind = token.Rbrack
	case ',':
		kind = token.Comma
	case ';':
		kind = token.Semi
	case ':':
		kind = token.Colon
	case '.':
		kind = token.Period
	default:
		l.error(start, "非法字符 "+lit)
		return token.Token{Kind: token.Illegal, Lit: "非法字符 " + lit, Pos: start, SpaceBefore: space}
	}
	return token.Token{Kind: kind, Lit: lit, Pos: start, SpaceBefore: space}
}

func (l *Lexer) skipGap() int {
	sawOther := false
	spaces := 0
	for l.pos < len(l.src) {
		if l.starts("//") {
			sawOther = true
			l.advance()
			l.advance()
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		r := l.peek()
		switch r {
		case ' ':
			spaces++
			l.advance()
		case '\t', '\n', '\r':
			sawOther = true
			spaces = 0
			l.advance()
		default:
			if sawOther {
				return -1
			}
			return spaces
		}
	}
	if sawOther {
		return -1
	}
	return spaces
}

func (l *Lexer) scanRawString() (string, bool) {
	l.advance()
	var b strings.Builder
	for l.pos < len(l.src) {
		r := l.peek()
		if r == '`' {
			l.advance()
			return b.String(), true
		}
		b.WriteRune(r)
		l.advance()
	}
	l.error(l.position(), "字符串没有结束")
	return b.String(), false
}

func (l *Lexer) scanString() (string, bool) {
	l.advance()
	var b strings.Builder
	for l.pos < len(l.src) {
		r := l.peek()
		if r == '"' {
			l.advance()
			return b.String(), true
		}
		if r == '\n' || r == '\r' {
			l.error(l.position(), "字符串没有结束")
			return b.String(), false
		}
		if r == '\\' {
			l.advance()
			if l.pos >= len(l.src) {
				break
			}
			esc := l.peek()
			l.advance()
			switch esc {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				l.error(l.position(), "不支持的转义")
				b.WriteRune(esc)
			}
			continue
		}
		b.WriteRune(r)
		l.advance()
	}
	l.error(l.position(), "字符串没有结束")
	return b.String(), false
}

func (l *Lexer) scanNumber() (token.Kind, string) {
	start := l.pos
	kind := token.Int
	for l.pos < len(l.src) && isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekAhead()) {
		kind = token.Float
		l.advance()
		for l.pos < len(l.src) && isDigit(l.peek()) {
			l.advance()
		}
	}
	if l.peek() == 'e' || l.peek() == 'E' {
		kind = token.Float
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		if !isDigit(l.peek()) {
			l.error(l.position(), "小数格式不对")
		}
		for l.pos < len(l.src) && isDigit(l.peek()) {
			l.advance()
		}
	}
	return kind, l.src[start:l.pos]
}

func (l *Lexer) scanIdent() string {
	start := l.pos
	l.advance()
	for l.pos < len(l.src) && isIdentCont(l.peek()) {
		l.advance()
	}
	return l.src[start:l.pos]
}

func (l *Lexer) starts(s string) bool {
	return strings.HasPrefix(l.src[l.pos:], s)
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	return r
}

func (l *Lexer) peekAhead() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	_, size := utf8.DecodeRuneInString(l.src[l.pos:])
	if l.pos+size >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos+size:])
	return r
}

func (l *Lexer) advance() {
	if l.pos >= len(l.src) {
		return
	}
	r, size := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
		return
	}
	l.col++
}

func (l *Lexer) position() token.Position {
	return token.Position{File: l.file, Offset: l.pos, Line: l.line, Col: l.col}
}

func (l *Lexer) error(pos token.Position, msg string) {
	if l.errs != nil {
		l.errs.Add(pos, msg)
	}
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentCont(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
