package parser

import (
	"fmt"

	"unigo/compiler/ast"
	"unigo/compiler/diag"
	"unigo/compiler/lexer"
	"unigo/compiler/token"
)

type Parser struct {
	lex       *lexer.Lexer
	cur       token.Token
	nxt       token.Token
	has       bool
	errs      *diag.List
	stopBrace bool
}

func Parse(filename string, src []byte) (*ast.File, diag.List) {
	var errs diag.List
	p := &Parser{
		lex:  lexer.New(filename, string(src), &errs),
		errs: &errs,
	}
	p.next()
	file := p.parseFile()
	return file, errs
}

func (p *Parser) next() {
	if p.has {
		p.cur = p.nxt
		p.has = false
		return
	}
	p.cur = p.lex.Next()
}

func (p *Parser) peek() token.Token {
	if !p.has {
		p.nxt = p.lex.Next()
		p.has = true
	}
	return p.nxt
}

func (p *Parser) error(pos token.Position, msg string) {
	p.errs.Add(pos, msg)
}

func (p *Parser) errorf(pos token.Position, format string, args ...any) {
	p.errs.Add(pos, fmt.Sprintf(format, args...))
}

func (p *Parser) expect(k token.Kind) {
	if p.cur.Kind == k {
		p.next()
		return
	}
	if k == token.Semi {
		p.error(p.cur.Pos, "缺少分号")
		return
	}
	p.errorf(p.cur.Pos, "期望 %s", k.String())
}

func (p *Parser) expectIdent() (string, token.Position) {
	if p.cur.Kind == token.Ident {
		name, pos := p.cur.Lit, p.cur.Pos
		p.next()
		return name, pos
	}
	if token.IsForbidden(p.cur.Kind) {
		p.errorf(p.cur.Pos, "禁止使用 %s", p.cur.Lit)
		pos := p.cur.Pos
		p.next()
		return "_bad", pos
	}
	pos := p.cur.Pos
	p.error(pos, "期望名字")
	return "_bad", pos
}

func (p *Parser) parseFile() *ast.File {
	file := &ast.File{}
	if p.cur.Kind != token.Package {
		p.error(p.cur.Pos, "文件必须以 package 开头")
	} else {
		at := p.cur.Pos
		p.next()
		name, _ := p.expectIdent()
		p.expect(token.Semi)
		file.Package = &ast.PackageDecl{At: at, Name: name}
	}
	if file.Package == nil {
		file.Package = &ast.PackageDecl{Name: "_bad", At: p.cur.Pos}
	}
	for p.cur.Kind == token.Import {
		file.Imports = append(file.Imports, p.parseImport())
	}
	for p.cur.Kind != token.EOF {
		if p.cur.Kind == token.Illegal {
			p.next()
			continue
		}
		before := p.cur.Pos.Offset
		d := p.parseDecl()
		if d != nil {
			file.Decls = append(file.Decls, d)
		}
		if p.cur.Kind != token.EOF && p.cur.Pos.Offset == before {
			p.next()
		}
	}
	return file
}

func (p *Parser) parseImport() *ast.ImportDecl {
	at := p.cur.Pos
	p.next()
	imp := &ast.ImportDecl{At: at, Name: "_bad"}
	switch p.cur.Kind {
	case token.String:
		p.error(p.cur.Pos, "导入必须写成 import 名字 from \"路径\"")
		p.next()
		p.expect(token.Semi)
		return imp
	case token.Lparen:
		p.error(p.cur.Pos, "禁止 import ()")
		p.next()
		for p.cur.Kind != token.Rparen && p.cur.Kind != token.EOF && p.cur.Kind != token.Semi {
			p.next()
		}
		if p.cur.Kind == token.Rparen {
			p.next()
		}
		p.expect(token.Semi)
		return imp
	}
	name, _ := p.expectIdent()
	imp.Name = name
	if name == "_" {
		p.error(at, "禁止 import _")
	}
	if p.cur.Kind != token.From {
		p.error(p.cur.Pos, "导入必须写成 import 名字 from \"路径\"")
	} else {
		p.next()
	}
	if p.cur.Kind != token.String {
		p.error(p.cur.Pos, "导入路径必须是字符串")
	} else {
		imp.Path = p.cur.Lit
		p.next()
	}
	p.expect(token.Semi)
	return imp
}

func (p *Parser) parseDecl() ast.Decl {
	switch p.cur.Kind {
	case token.Const:
		return p.parseConst()
	case token.Var:
		return p.parseVar()
	case token.Func:
		return p.parseFunc()
	case token.Enum:
		return p.parseEnum()
	case token.Struct:
		return p.parseStruct()
	case token.Interface:
		return p.parseInterface()
	case token.Newtype:
		return p.parseNewtype(true)
	case token.Alias:
		return p.parseNewtype(false)
	case token.TypeWord:
		p.error(p.cur.Pos, "禁止 type，请使用 newtype 或 alias")
		p.next()
		return nil
	case token.Class, token.New, token.Extends, token.Implements:
		p.errorf(p.cur.Pos, "禁止使用 %s", p.cur.Lit)
		p.next()
		return nil
	default:
		p.error(p.cur.Pos, "不支持的声明")
		p.next()
		return nil
	}
}

func (p *Parser) parseConst() *ast.ConstDecl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	typ := p.parseAnnoType()
	p.expect(token.Assign)
	val := p.parseExpr()
	p.expect(token.Semi)
	return &ast.ConstDecl{At: at, Name: name, NamePos: namePos, Type: typ, Value: val}
}

func (p *Parser) parseVar() *ast.VarDecl {
	at := p.cur.Pos
	p.next()
	d := &ast.VarDecl{At: at}
	if p.cur.Kind == token.Lparen {
		d.Paren = true
		p.next()
		for p.cur.Kind != token.Rparen && p.cur.Kind != token.EOF && p.cur.Kind != token.Semi {
			d.Specs = append(d.Specs, p.parseVarSpec())
			if p.cur.Kind == token.Comma {
				p.next()
				continue
			}
			break
		}
		p.expect(token.Rparen)
		if len(d.Specs) < 2 {
			p.error(at, "只接收一个值时不能加括号")
		}
	} else {
		d.Specs = append(d.Specs, p.parseVarSpec())
		if p.cur.Kind == token.Comma {
			p.error(p.cur.Pos, "多个值必须放进圆括号")
			for p.cur.Kind != token.Assign && p.cur.Kind != token.Semi && p.cur.Kind != token.EOF && p.cur.Kind != token.Rbrace {
				p.next()
			}
		}
	}
	p.expect(token.Assign)
	d.Value = p.parseExpr()
	if p.cur.Kind == token.Else {
		epos := p.cur.Pos
		p.next()
		p.expect(token.Return)
		er := &ast.ElseReturn{At: epos}
		if p.cur.Kind == token.Semi {
			er.Bare = true
		} else {
			er.Value = p.parseExpr()
		}
		d.Else = er
	}
	p.expect(token.Semi)
	if d.Value == nil {
		d.Value = p.badExpr()
	}
	return d
}

func (p *Parser) parseVarSpec() *ast.VarSpec {
	name, pos := p.expectIdent()
	typ := p.parseAnnoType()
	return &ast.VarSpec{At: pos, Name: name, Type: typ}
}

func (p *Parser) parseAnnoType() ast.TypeExpr {
	if p.cur.Kind != token.Colon {
		p.error(p.cur.Pos, "类型注解必须写成 名字 :类型")
		return &ast.IdentType{At: p.cur.Pos, Name: "_bad"}
	}
	if p.cur.SpaceBefore != 1 {
		p.error(p.cur.Pos, "类型注解必须写成 名字 :类型，冒号前恰好一个空格")
	}
	p.next()
	if p.cur.SpaceBefore != 0 {
		p.error(p.cur.Pos, "类型注解必须写成 名字 :类型，冒号必须紧贴类型")
	}
	return p.parseType()
}

func (p *Parser) parseResult() ast.TypeExpr {
	if p.cur.Kind != token.Colon {
		return nil
	}
	if p.cur.SpaceBefore != 1 {
		p.error(p.cur.Pos, "类型注解必须写成 名字 :类型，冒号前恰好一个空格")
	}
	p.next()
	if p.cur.Kind == token.Lparen {
		p.error(p.cur.Pos, "函数只能有一个返回值")
		for p.cur.Kind != token.Rparen && p.cur.Kind != token.Lbrace && p.cur.Kind != token.EOF && p.cur.Kind != token.Semi {
			p.next()
		}
		if p.cur.Kind == token.Rparen {
			p.next()
		}
		return nil
	}
	if p.cur.SpaceBefore != 0 {
		p.error(p.cur.Pos, "类型注解必须写成 名字 :类型，冒号必须紧贴类型")
	}
	return p.parseType()
}

func (p *Parser) parseType() ast.TypeExpr {
	var base ast.TypeExpr
	switch p.cur.Kind {
	case token.Mul:
		at := p.cur.Pos
		p.next()
		base = &ast.PointerType{At: at, Base: p.parseQualType()}
	case token.Map:
		at := p.cur.Pos
		p.next()
		p.expect(token.Lbrack)
		key := p.parseType()
		p.expect(token.Comma)
		val := p.parseType()
		p.expect(token.Rbrack)
		base = &ast.MapType{At: at, Key: key, Value: val}
	default:
		base = p.parseQualType()
	}
	for p.cur.Kind == token.Lbrack && p.peek().Kind == token.Rbrack {
		at := p.cur.Pos
		p.next()
		p.next()
		base = &ast.SliceType{At: at, Elem: base}
	}
	return base
}

func (p *Parser) parseQualType() ast.TypeExpr {
	name, pos := p.expectIdent()
	if p.cur.Kind == token.Period {
		p.next()
		sel, _ := p.expectIdent()
		return &ast.SelectorType{At: pos, Pkg: name, Name: sel}
	}
	return &ast.IdentType{At: pos, Name: name}
}

func (p *Parser) parseFunc() *ast.FuncDecl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	p.expect(token.Lparen)
	params := p.parseParams()
	p.expect(token.Rparen)
	result := p.parseResult()
	body := p.parseBlock()
	return &ast.FuncDecl{At: at, Name: name, NamePos: namePos, Params: params, Result: result, Body: body}
}

func (p *Parser) parseParams() []*ast.VarSpec {
	if p.cur.Kind == token.Rparen {
		return nil
	}
	var ps []*ast.VarSpec
	for p.cur.Kind != token.Rparen && p.cur.Kind != token.EOF && p.cur.Kind != token.Lbrace {
		ps = append(ps, p.parseVarSpec())
		if p.cur.Kind == token.Comma {
			p.next()
			if p.cur.Kind == token.Rparen {
				break
			}
			continue
		}
		break
	}
	return ps
}

func (p *Parser) parseEnum() *ast.EnumDecl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	p.expect(token.Lbrace)
	var members []ast.EnumMember
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		if p.cur.Kind == token.Iota {
			p.error(p.cur.Pos, "禁止 iota")
			p.next()
			continue
		}
		m, mpos := p.expectIdent()
		p.expect(token.Semi)
		if m != "_bad" {
			members = append(members, ast.EnumMember{At: mpos, Name: m})
		}
	}
	p.expect(token.Rbrace)
	return &ast.EnumDecl{At: at, Name: name, NamePos: namePos, Members: members}
}

func (p *Parser) parseStruct() *ast.StructDecl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	p.expect(token.Lbrace)
	var fields []*ast.VarSpec
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		if p.cur.Kind == token.Func {
			p.error(p.cur.Pos, "结构体里不能写函数")
			p.next()
			continue
		}
		fields = append(fields, p.parseVarSpec())
		p.expect(token.Semi)
	}
	p.expect(token.Rbrace)
	return &ast.StructDecl{At: at, Name: name, NamePos: namePos, Fields: fields}
}

func (p *Parser) parseInterface() *ast.InterfaceDecl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	p.expect(token.Lbrace)
	var methods []*ast.FuncSig
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		methods = append(methods, p.parseFuncSig())
		p.expect(token.Semi)
	}
	p.expect(token.Rbrace)
	return &ast.InterfaceDecl{At: at, Name: name, NamePos: namePos, Methods: methods}
}

func (p *Parser) parseFuncSig() *ast.FuncSig {
	at := p.cur.Pos
	p.expect(token.Func)
	name, namePos := p.expectIdent()
	p.expect(token.Lparen)
	params := p.parseParams()
	p.expect(token.Rparen)
	result := p.parseResult()
	return &ast.FuncSig{At: at, Name: name, NamePos: namePos, Params: params, Result: result}
}

func (p *Parser) parseNewtype(isNew bool) ast.Decl {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	if p.cur.Kind == token.Colon {
		if isNew {
			p.error(p.cur.Pos, "新类型要写成 newtype 名字 = 类型")
		} else {
			p.error(p.cur.Pos, "别名要写成 alias 名字 = 类型")
		}
	}
	p.expect(token.Assign)
	typ := p.parseType()
	p.expect(token.Semi)
	if isNew {
		return &ast.NewtypeDecl{At: at, Name: name, NamePos: namePos, Type: typ}
	}
	return &ast.AliasDecl{At: at, Name: name, NamePos: namePos, Type: typ}
}

func (p *Parser) parseBlock() *ast.Block {
	at := p.cur.Pos
	if p.cur.Kind != token.Lbrace {
		p.error(p.cur.Pos, "分支必须使用花括号")
		return &ast.Block{At: at}
	}
	p.next()
	b := &ast.Block{At: at}
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		if p.cur.Kind == token.Illegal {
			p.next()
			continue
		}
		before := p.cur.Pos.Offset
		st := p.parseStmt()
		if st != nil {
			b.List = append(b.List, st)
		}
		if p.cur.Kind != token.EOF && p.cur.Kind != token.Rbrace && p.cur.Pos.Offset == before {
			p.next()
		}
	}
	p.expect(token.Rbrace)
	return b
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.cur.Kind {
	case token.Const:
		return p.parseConst()
	case token.Var:
		return p.parseVar()
	case token.If:
		return p.parseIf()
	case token.While:
		return p.parseWhile()
	case token.For:
		return p.parseFor()
	case token.Switch:
		return p.parseSwitch()
	case token.Return:
		return p.parseReturn()
	case token.Fallthrough:
		p.error(p.cur.Pos, "禁止 fallthrough")
		p.next()
		return nil
	case token.Semi:
		p.error(p.cur.Pos, "不支持空语句")
		p.next()
		return nil
	case token.Rbrace, token.EOF:
		return nil
	case token.Class, token.New, token.Extends, token.Implements, token.TypeWord, token.Iota:
		p.errorf(p.cur.Pos, "禁止使用 %s", p.cur.Lit)
		p.next()
		return nil
	default:
		expr := p.parseExpr()
		if p.cur.Kind == token.Assign {
			at := expr.Pos()
			p.next()
			val := p.parseExpr()
			p.expect(token.Semi)
			return &ast.AssignStmt{At: at, Target: expr, Value: val}
		}
		p.expect(token.Semi)
		return &ast.ExprStmt{At: expr.Pos(), X: expr}
	}
}

func (p *Parser) parseIf() *ast.IfStmt {
	at := p.cur.Pos
	p.next()
	cond := p.parseCond()
	body := p.parseBlock()
	s := &ast.IfStmt{At: at, Cond: cond, Body: body}
	for p.cur.Kind == token.Elif {
		epos := p.cur.Pos
		p.next()
		c := p.parseCond()
		b := p.parseBlock()
		s.Elifs = append(s.Elifs, ast.Elif{At: epos, Cond: c, Body: b})
	}
	if p.cur.Kind == token.Else {
		p.next()
		s.Else = p.parseBlock()
	}
	return s
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	at := p.cur.Pos
	p.next()
	return &ast.WhileStmt{At: at, Cond: p.parseCond(), Body: p.parseBlock()}
}

func (p *Parser) parseFor() *ast.ForStmt {
	at := p.cur.Pos
	p.next()
	return &ast.ForStmt{At: at, Cond: p.parseCond(), Body: p.parseBlock()}
}

func (p *Parser) parseCond() ast.Expr {
	if p.cur.Kind != token.Lparen {
		p.error(p.cur.Pos, "条件必须放在括号里")
		return p.badExpr()
	}
	p.next()
	e := p.parseExpr()
	p.expect(token.Rparen)
	return e
}

func (p *Parser) parseSwitch() *ast.SwitchStmt {
	at := p.cur.Pos
	p.next()
	tag := p.parseCond()
	if p.cur.Kind != token.Lbrace {
		p.error(p.cur.Pos, "分支必须使用花括号")
		return &ast.SwitchStmt{At: at, Tag: tag}
	}
	p.next()
	s := &ast.SwitchStmt{At: at, Tag: tag}
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		switch p.cur.Kind {
		case token.Case:
			cpos := p.cur.Pos
			p.next()
			prev := p.stopBrace
			p.stopBrace = true
			val := p.parseExpr()
			p.stopBrace = prev
			body := p.parseBlock()
			s.Cases = append(s.Cases, &ast.CaseClause{At: cpos, Value: val, Body: body})
		case token.Default:
			cpos := p.cur.Pos
			p.next()
			body := p.parseBlock()
			s.Cases = append(s.Cases, &ast.CaseClause{At: cpos, IsDefault: true, Body: body})
		case token.Fallthrough:
			p.error(p.cur.Pos, "禁止 fallthrough")
			p.next()
		default:
			p.error(p.cur.Pos, "switch 里只能写 case 或 default")
			p.next()
		}
	}
	p.expect(token.Rbrace)
	return s
}

func (p *Parser) parseReturn() *ast.ReturnStmt {
	at := p.cur.Pos
	p.next()
	s := &ast.ReturnStmt{At: at}
	if p.cur.Kind != token.Semi {
		s.Value = p.parseExpr()
	}
	p.expect(token.Semi)
	return s
}

func (p *Parser) parseExpr() ast.Expr {
	return p.parseBinary(1)
}

func (p *Parser) parseBinary(min int) ast.Expr {
	left := p.parseUnary()
	for {
		prec := prec(p.cur.Kind)
		if prec < min {
			break
		}
		op := p.cur.Kind
		at := p.cur.Pos
		p.next()
		right := p.parseBinary(prec + 1)
		left = &ast.BinaryExpr{At: at, Op: op, X: left, Y: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	switch p.cur.Kind {
	case token.And, token.Not, token.Sub:
		op := p.cur.Kind
		at := p.cur.Pos
		p.next()
		return &ast.UnaryExpr{At: at, Op: op, X: p.parseUnary()}
	default:
		return p.parsePrimary()
	}
}

func (p *Parser) parsePrimary() ast.Expr {
	switch p.cur.Kind {
	case token.String, token.Int, token.Float, token.True, token.False, token.Nil:
		lit := &ast.BasicLit{At: p.cur.Pos, Kind: p.cur.Kind, Value: p.cur.Lit}
		p.next()
		return p.parsePostfix(lit)
	case token.Ident:
		id := &ast.Ident{At: p.cur.Pos, Name: p.cur.Lit}
		p.next()
		return p.parsePostfix(id)
	case token.Lbrack:
		return p.parsePostfix(p.parseSliceLit())
	case token.Lbrace:
		return p.parsePostfix(p.parseMapLit())
	case token.Lparen:
		p.next()
		e := p.parseExpr()
		p.expect(token.Rparen)
		return p.parsePostfix(e)
	default:
		if token.IsForbidden(p.cur.Kind) {
			p.errorf(p.cur.Pos, "禁止使用 %s", p.cur.Lit)
			p.next()
			return p.badExpr()
		}
		p.error(p.cur.Pos, "期望表达式")
		return p.badExpr()
	}
}

func (p *Parser) parsePostfix(x ast.Expr) ast.Expr {
	for {
		switch p.cur.Kind {
		case token.Period:
			p.next()
			sel, pos := p.expectIdent()
			x = &ast.SelectorExpr{At: pos, X: x, Sel: sel}
		case token.Lparen:
			x = p.parseCall(x)
		case token.Lbrack:
			at := p.cur.Pos
			p.next()
			if p.cur.Kind == token.Colon {
				p.error(p.cur.Pos, "不支持切片区间")
			}
			idx := p.parseExpr()
			p.expect(token.Rbrack)
			x = &ast.IndexExpr{At: at, X: x, Index: idx}
		case token.Lbrace:
			if p.stopBrace {
				return x
			}
			if _, ok := x.(*ast.Ident); !ok {
				if _, ok := x.(*ast.SelectorExpr); !ok {
					return x
				}
			}
			x = p.parseComposite(x)
		default:
			return x
		}
	}
}

func (p *Parser) parseCall(fn ast.Expr) ast.Expr {
	at := p.cur.Pos
	p.next()
	var args []ast.Expr
	if p.cur.Kind != token.Rparen {
		for {
			args = append(args, p.parseExpr())
			if p.cur.Kind == token.Comma {
				p.next()
				if p.cur.Kind == token.Rparen {
					break
				}
				continue
			}
			break
		}
	}
	p.expect(token.Rparen)
	return &ast.CallExpr{At: at, Fun: fn, Args: args}
}

func (p *Parser) parseSliceLit() ast.Expr {
	at := p.cur.Pos
	p.next()
	var elts []ast.Expr
	if p.cur.Kind != token.Rbrack {
		for {
			elts = append(elts, p.parseExpr())
			if p.cur.Kind == token.Comma {
				p.next()
				if p.cur.Kind == token.Rbrack {
					break
				}
				continue
			}
			break
		}
	}
	p.expect(token.Rbrack)
	return &ast.SliceLit{At: at, Elts: elts}
}

func (p *Parser) parseMapLit() ast.Expr {
	at := p.cur.Pos
	p.next()
	lit := &ast.MapLit{At: at}
	if p.cur.Kind != token.Rbrace {
		for {
			key := p.parseExpr()
			if p.cur.Kind == token.Assign {
				p.error(p.cur.Pos, "映射字面量要用冒号")
			}
			p.expect(token.Colon)
			val := p.parseExpr()
			lit.Entries = append(lit.Entries, ast.MapEntry{Key: key, Value: val})
			if p.cur.Kind == token.Comma {
				p.next()
				if p.cur.Kind == token.Rbrace {
					break
				}
				continue
			}
			break
		}
	}
	p.expect(token.Rbrace)
	return lit
}

func (p *Parser) parseComposite(typ ast.Expr) ast.Expr {
	at := p.cur.Pos
	p.next()
	if p.cur.Kind == token.Rbrace {
		p.next()
		return &ast.StructLit{At: at, Type: typ}
	}
	if p.cur.Kind == token.Func {
		lit := &ast.InterfaceLit{At: at, Type: typ}
		for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
			lit.Methods = append(lit.Methods, p.parseFuncLit())
			p.expect(token.Semi)
		}
		p.expect(token.Rbrace)
		return lit
	}
	lit := &ast.StructLit{At: at, Type: typ}
	for p.cur.Kind != token.Rbrace && p.cur.Kind != token.EOF {
		name, pos := p.expectIdent()
		if p.cur.Kind == token.Colon {
			p.error(p.cur.Pos, "结构体字面量的字段要用等号")
		}
		p.expect(token.Assign)
		val := p.parseExpr()
		lit.Fields = append(lit.Fields, ast.FieldInit{At: pos, Name: name, Value: val})
		if p.cur.Kind == token.Comma {
			p.next()
			if p.cur.Kind == token.Rbrace {
				break
			}
			continue
		}
		break
	}
	p.expect(token.Rbrace)
	return lit
}

func (p *Parser) parseFuncLit() *ast.FuncLit {
	at := p.cur.Pos
	p.next()
	name, namePos := p.expectIdent()
	p.expect(token.Lparen)
	params := p.parseParams()
	p.expect(token.Rparen)
	result := p.parseResult()
	body := p.parseBlock()
	return &ast.FuncLit{At: at, Name: name, NamePos: namePos, Params: params, Result: result, Body: body}
}

func (p *Parser) badExpr() ast.Expr {
	return &ast.Ident{At: p.cur.Pos, Name: "_bad"}
}

func prec(k token.Kind) int {
	switch k {
	case token.Lor:
		return 1
	case token.Land:
		return 2
	case token.Eql, token.Neq:
		return 3
	case token.Lss, token.Gtr, token.Leq, token.Geq:
		return 4
	case token.Add, token.Sub:
		return 5
	case token.Mul, token.Quo, token.Rem:
		return 6
	default:
		return 0
	}
}
