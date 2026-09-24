package check

import (
	"fmt"
	gotok "go/token"
	"go/importer"
	"go/types"
	"unicode/utf8"

	"unigo/compiler/ast"
	"unigo/compiler/diag"
	"unigo/compiler/names"
	"unigo/compiler/token"
)

const (
	tsNone = iota
	tsDoing
	tsDone
)

type scope struct {
	parent *scope
	objs   map[string]*Object
}

func (s *scope) lookup(name string) *Object {
	for s != nil {
		if o, ok := s.objs[name]; ok {
			return o
		}
		s = s.parent
	}
	return nil
}

func (s *scope) bind(o *Object) *Object {
	if prev, ok := s.objs[o.Name]; ok {
		return prev
	}
	s.objs[o.Name] = o
	return nil
}

type Checker struct {
	file      *ast.File
	errs      diag.List
	info      *Info
	typesPkg  *types.Package
	pkgScope  *scope
	scope     *scope
	inFunc    bool
	hasResult bool
	fnResult  types.Type
	tstate    map[string]int
	typeDecls map[string]ast.Decl
	named     map[string]*types.Named
	enums     map[*types.Named]*EnumInfo
	enumName  map[string]*EnumInfo
	imports   map[string]*ImportInfo
	reserved  map[string]bool
	errorType types.Type
	boolType  types.Type
	stringTyp types.Type
	intType   types.Type
}

func Check(file *ast.File) (*Info, diag.List) {
	c := newChecker(file)
	c.checkPackage()
	c.loadImports()
	c.collectTypes()
	for name := range c.typeDecls {
		c.completeType(name)
	}
	c.declareFuncs()
	c.checkValuesAndBodies()
	return c.info, c.errs
}

func newChecker(file *ast.File) *Checker {
	pkgName := "main"
	if file != nil && file.Package != nil && file.Package.Name != "" {
		pkgName = file.Package.Name
	}
	tpkg := types.NewPackage("unigo/out", pkgName)
	sc := &scope{objs: map[string]*Object{}}
	info := newInfo()
	info.Pkg = tpkg
	info.Fset = gotok.NewFileSet()
	return &Checker{
		file:      file,
		info:      info,
		typesPkg:  tpkg,
		pkgScope:  sc,
		scope:     sc,
		tstate:    map[string]int{},
		typeDecls: map[string]ast.Decl{},
		named:     map[string]*types.Named{},
		enums:     map[*types.Named]*EnumInfo{},
		enumName:  map[string]*EnumInfo{},
		imports:   map[string]*ImportInfo{},
		reserved:  map[string]bool{},
		errorType: types.Universe.Lookup("error").Type(),
		boolType:  types.Universe.Lookup("bool").Type(),
		stringTyp: types.Universe.Lookup("string").Type(),
		intType:   types.Universe.Lookup("int").Type(),
	}
}

func (c *Checker) error(pos token.Position, msg string) {
	c.errs.Add(pos, msg)
}

func (c *Checker) checkName(name string, pos token.Position) {
	if name == "" || name == "_bad" {
		return
	}
	if len(name) >= len("unigoBind") && name[:len("unigoBind")] == "unigoBind" {
		c.error(pos, "名字不能以 unigoBind 开头")
	}
	if utf8.RuneCountInString(name) < 2 {
		c.error(pos, "名字至少要两个字符")
	}
}

func (c *Checker) bind(o *Object, pos token.Position) {
	if o.Name == "" || o.Name == "_bad" {
		return
	}
	c.checkName(o.Name, pos)
	if c.reserved[o.Name] {
		c.error(pos, "名字与生成的枚举常量冲突")
	}
	if prev := c.scope.bind(o); prev != nil {
		c.error(pos, "名字重复定义")
	}
}

func (c *Checker) openScope() {
	c.scope = &scope{parent: c.scope, objs: map[string]*Object{}}
}

func (c *Checker) closeScope() {
	if c.scope != nil && c.scope.parent != nil {
		c.scope = c.scope.parent
	}
}

func (c *Checker) checkPackage() {
	if c.file == nil || c.file.Package == nil {
		return
	}
	c.checkName(c.file.Package.Name, c.file.Package.At)
}

func (c *Checker) loadImports() {
	imp := importer.ForCompiler(gotok.NewFileSet(), "gc", nil)
	seen := map[string]bool{}
	for _, im := range c.file.Imports {
		if im.Name == "_" {
			c.error(im.At, "禁止 import _")
			continue
		}
		c.checkName(im.Name, im.At)
		if im.Path == "" {
			continue
		}
		if seen[im.Path] {
			c.error(im.At, "重复导入")
		}
		seen[im.Path] = true
		pkg, err := imp.Import(im.Path)
		if err != nil {
			c.error(im.At, "无法导入 "+im.Path)
			continue
		}
		info := &ImportInfo{Local: im.Name, Path: im.Path, Name: pkg.Name(), Pkg: pkg}
		c.imports[im.Name] = info
		c.info.Imports = append(c.info.Imports, info)
		c.bind(&Object{Name: im.Name, Kind: ObjImport}, im.At)
	}
}

func (c *Checker) collectTypes() {
	for _, d := range c.file.Decls {
		switch d := d.(type) {
		case *ast.StructDecl:
			c.addShell(d.Name, d.NamePos, d)
		case *ast.EnumDecl:
			c.addShell(d.Name, d.NamePos, d)
		case *ast.InterfaceDecl:
			c.addShell(d.Name, d.NamePos, d)
		case *ast.NewtypeDecl:
			c.addShell(d.Name, d.NamePos, d)
		case *ast.AliasDecl:
			c.checkName(d.Name, d.NamePos)
			c.bind(&Object{Name: d.Name, Kind: ObjType}, d.NamePos)
			c.typeDecls[d.Name] = d
		}
	}
}

func (c *Checker) addShell(name string, pos token.Position, decl ast.Decl) {
	c.checkName(name, pos)
	tn := types.NewTypeName(gotok.NoPos, c.typesPkg, name, nil)
	named := types.NewNamed(tn, nil, nil)
	c.bind(&Object{Name: name, Kind: ObjType, Typ: named}, pos)
	c.typesPkg.Scope().Insert(tn)
	c.named[name] = named
	c.typeDecls[name] = decl
}

func (c *Checker) completeType(name string) {
	switch c.tstate[name] {
	case tsDone:
		return
	case tsDoing:
		return
	}
	decl, ok := c.typeDecls[name]
	if !ok {
		return
	}
	c.tstate[name] = tsDoing
	switch d := decl.(type) {
	case *ast.StructDecl:
		c.fillStruct(d)
	case *ast.EnumDecl:
		c.fillEnum(d)
	case *ast.InterfaceDecl:
		c.fillInterface(d)
	case *ast.NewtypeDecl:
		c.fillNewtype(d)
	case *ast.AliasDecl:
		c.fillAlias(d)
	}
	c.tstate[name] = tsDone
}

func (c *Checker) resolveType(t ast.TypeExpr) types.Type {
	if t == nil {
		return types.Typ[types.Invalid]
	}
	switch t := t.(type) {
	case *ast.IdentType:
		return c.resolveTypeName(t.Name, t.At)
	case *ast.SelectorType:
		return c.resolveSelectorType(t)
	case *ast.PointerType:
		return types.NewPointer(c.resolveType(t.Base))
	case *ast.SliceType:
		return types.NewSlice(c.resolveType(t.Elem))
	case *ast.MapType:
		key := c.resolveType(t.Key)
		val := c.resolveType(t.Value)
		if !isInvalid(key) && key.Underlying() != nil && !types.Comparable(key) {
			c.error(t.At, "映射的键必须能比较")
		}
		if isInvalid(key) {
			key = types.Typ[types.String]
		}
		if isInvalid(val) {
			val = types.Typ[types.Invalid]
		}
		return types.NewMap(key, val)
	default:
		c.error(t.Pos(), "不支持的类型")
		return types.Typ[types.Invalid]
	}
}

func (c *Checker) resolveTypeName(name string, pos token.Position) types.Type {
	if name == "" || name == "_bad" {
		return types.Typ[types.Invalid]
	}
	if obj := c.pkgScope.lookup(name); obj != nil && obj.Kind == ObjType {
		if c.tstate[name] == tsDoing {
			c.error(pos, "类型定义出现循环")
			return types.Typ[types.Invalid]
		}
		if c.tstate[name] == tsNone {
			c.completeType(name)
		}
		if obj.Typ == nil {
			c.error(pos, "类型定义出现循环")
			return types.Typ[types.Invalid]
		}
		return obj.Typ
	}
	if bt := builtinType(name); bt != nil {
		return bt
	}
	c.error(pos, "未知类型 "+name)
	return types.Typ[types.Invalid]
}

func (c *Checker) resolveSelectorType(t *ast.SelectorType) types.Type {
	imp := c.imports[t.Pkg]
	if imp == nil || imp.Pkg == nil {
		c.error(t.At, "未知的包 "+t.Pkg)
		return types.Typ[types.Invalid]
	}
	imp.Used = true
	obj := imp.Pkg.Scope().Lookup(t.Name)
	tn, ok := obj.(*types.TypeName)
	if !ok {
		c.error(t.At, "包里没有类型 "+t.Name)
		return types.Typ[types.Invalid]
	}
	return tn.Type()
}

func builtinType(name string) types.Type {
	obj := types.Universe.Lookup(name)
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil
	}
	return tn.Type()
}

func (c *Checker) fillStruct(d *ast.StructDecl) {
	named := c.named[d.Name]
	seen := map[string]bool{}
	seenGo := map[string]bool{}
	var fields []*types.Var
	var tags []string
	for _, f := range d.Fields {
		c.checkName(f.Name, f.At)
		if seen[f.Name] {
			c.error(f.At, "字段重复")
		}
		seen[f.Name] = true
		goName := names.Export(f.Name)
		if seenGo[goName] {
			c.error(f.At, "字段导出后重名")
		}
		seenGo[goName] = true
		ft := c.resolveType(f.Type)
		if isInvalid(ft) {
			ft = types.Typ[types.Invalid]
		}
		fields = append(fields, types.NewField(gotok.NoPos, c.typesPkg, f.Name, ft, false))
		tags = append(tags, fmt.Sprintf("json:%q", f.Name))
	}
	if named.Underlying() == nil {
		named.SetUnderlying(types.NewStruct(fields, tags))
	}
}

func (c *Checker) fillEnum(d *ast.EnumDecl) {
	named := c.named[d.Name]
	if named.Underlying() == nil {
		named.SetUnderlying(types.Typ[types.Int])
	}
	seen := map[string]bool{}
	var members []string
	for _, m := range d.Members {
		c.checkName(m.Name, m.At)
		if seen[m.Name] {
			c.error(m.At, "枚举成员重复")
		}
		seen[m.Name] = true
		members = append(members, m.Name)
		gen := d.Name + "_" + m.Name
		c.reserved[gen] = true
		if c.pkgScope.lookup(gen) != nil {
			c.error(m.At, "名字与生成的枚举常量冲突")
		}
	}
	info := &EnumInfo{Name: d.Name, Members: members, Typ: named}
	c.enums[named] = info
	c.enumName[d.Name] = info
	c.info.Enums = append(c.info.Enums, info)
}

func (c *Checker) fillInterface(d *ast.InterfaceDecl) {
	named := c.named[d.Name]
	seen := map[string]bool{}
	var methods []*types.Func
	for _, m := range d.Methods {
		c.checkName(m.Name, m.NamePos)
		if seen[m.Name] {
			c.error(m.NamePos, "接口函数重复")
		}
		seen[m.Name] = true
		for _, p := range m.Params {
			c.checkName(p.Name, p.At)
		}
		sig := c.makeSig(m.Params, m.Result)
		methods = append(methods, types.NewFunc(gotok.NoPos, c.typesPkg, m.Name, sig))
	}
	if named.Underlying() == nil {
		named.SetUnderlying(types.NewInterfaceType(methods, nil))
	}
}

func (c *Checker) fillNewtype(d *ast.NewtypeDecl) {
	named := c.named[d.Name]
	rhs := c.resolveType(d.Type)
	u := underlyingOf(rhs)
	if u == nil || isInvalid(u) {
		u = types.Typ[types.Invalid]
	}
	if _, ok := u.(*types.Named); ok {
		c.error(d.At, "类型定义出现循环")
		u = types.Typ[types.Invalid]
	}
	if named.Underlying() == nil {
		named.SetUnderlying(u)
	}
}

func (c *Checker) fillAlias(d *ast.AliasDecl) {
	rhs := c.resolveType(d.Type)
	if isInvalid(rhs) {
		rhs = types.Typ[types.Invalid]
	}
	tn := types.NewTypeName(gotok.NoPos, c.typesPkg, d.Name, nil)
	alias := types.NewAlias(tn, rhs)
	if obj := c.pkgScope.lookup(d.Name); obj != nil {
		obj.Typ = alias
	}
	_ = c.typesPkg.Scope().Insert(tn)
}

func underlyingOf(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	t = types.Unalias(t)
	if n, ok := t.(*types.Named); ok {
		if n.Underlying() == nil {
			return nil
		}
		return n.Underlying()
	}
	return t
}

func (c *Checker) makeSig(params []*ast.VarSpec, result ast.TypeExpr) *types.Signature {
	vars := make([]*types.Var, 0, len(params))
	seen := map[string]bool{}
	for _, p := range params {
		if p.Name != "" && p.Name != "_bad" && seen[p.Name] {
			c.error(p.At, "参数名字重复")
		}
		seen[p.Name] = true
		pt := c.resolveType(p.Type)
		if isInvalid(pt) {
			pt = types.Typ[types.Invalid]
		}
		vars = append(vars, types.NewParam(gotok.NoPos, c.typesPkg, p.Name, pt))
	}
	var results *types.Tuple
	if result != nil {
		rt := c.resolveType(result)
		if isInvalid(rt) {
			rt = types.Typ[types.Invalid]
		}
		results = types.NewTuple(types.NewVar(gotok.NoPos, c.typesPkg, "", rt))
	}
	return types.NewSignatureType(nil, nil, nil, types.NewTuple(vars...), results, false)
}

func (c *Checker) declareFuncs() {
	for _, d := range c.file.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		c.checkName(fn.Name, fn.NamePos)
		sig := c.makeSig(fn.Params, fn.Result)
		c.bind(&Object{Name: fn.Name, Kind: ObjFunc, Typ: sig}, fn.NamePos)
	}
}

func (c *Checker) checkValuesAndBodies() {
	for _, d := range c.file.Decls {
		switch d := d.(type) {
		case *ast.ConstDecl:
			c.checkConst(d, false)
		case *ast.VarDecl:
			c.checkVar(d, false)
		}
	}
	for _, d := range c.file.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			c.checkFunc(fn)
		}
	}
}

func (c *Checker) checkConst(d *ast.ConstDecl, local bool) {
	typ := c.resolveType(d.Type)
	got := c.checkExprHint(d.Value, typ)
	if !c.assignableTo(got, typ) {
		c.error(d.Value.Pos(), "常量类型不符")
	}
	obj := &Object{Name: d.Name, Kind: ObjConst, Typ: typ, Local: local}
	c.bind(obj, d.NamePos)
	c.info.Consts[d] = obj
}

func (c *Checker) checkFunc(fn *ast.FuncDecl) {
	obj := c.pkgScope.lookup(fn.Name)
	if obj == nil || obj.Kind != ObjFunc {
		return
	}
	sig, _ := obj.Typ.(*types.Signature)
	c.openScope()
	prevIn, prevHas, prevRes := c.inFunc, c.hasResult, c.fnResult
	c.inFunc = true
	c.hasResult = sig != nil && sig.Results() != nil && sig.Results().Len() > 0
	if c.hasResult {
		c.fnResult = sig.Results().At(0).Type()
	} else {
		c.fnResult = nil
	}
	if sig != nil && sig.Params() != nil {
		n := sig.Params().Len()
		if n > len(fn.Params) {
			n = len(fn.Params)
		}
		for i := 0; i < n; i++ {
			p := fn.Params[i]
			c.bind(&Object{Name: p.Name, Kind: ObjParam, Typ: sig.Params().At(i).Type(), Local: true}, p.At)
		}
	}
	c.checkBlock(fn.Body, false)
	if c.hasResult && !c.blockReturns(fn.Body) {
		c.error(fn.At, "缺少 return")
	}
	c.closeScope()
	c.inFunc, c.hasResult, c.fnResult = prevIn, prevHas, prevRes
}

func (c *Checker) checkBlock(b *ast.Block, newScope bool) {
	if b == nil {
		return
	}
	if newScope {
		c.openScope()
	}
	for _, s := range b.List {
		c.checkStmt(s)
	}
	if newScope {
		c.closeScope()
	}
}

func (c *Checker) checkStmt(s ast.Stmt) {
	if s == nil {
		return
	}
	switch s := s.(type) {
	case *ast.ConstDecl:
		c.checkConst(s, true)
	case *ast.VarDecl:
		c.checkVar(s, true)
	case *ast.AssignStmt:
		c.checkAssign(s)
	case *ast.IfStmt:
		c.checkCond(s.Cond)
		c.checkBlock(s.Body, true)
		for _, e := range s.Elifs {
			c.checkCond(e.Cond)
			c.checkBlock(e.Body, true)
		}
		if s.Else != nil {
			c.checkBlock(s.Else, true)
		}
	case *ast.WhileStmt:
		c.checkCond(s.Cond)
		c.checkBlock(s.Body, true)
	case *ast.ForStmt:
		c.checkCond(s.Cond)
		c.checkBlock(s.Body, true)
	case *ast.SwitchStmt:
		c.checkSwitch(s)
	case *ast.ReturnStmt:
		c.checkReturn(s)
	case *ast.ExprStmt:
		c.checkExprStmt(s)
	case *ast.Block:
		c.checkBlock(s, true)
	default:
		c.error(s.Pos(), "不支持的语句")
	}
}

func (c *Checker) checkCond(e ast.Expr) {
	t := c.checkExpr(e)
	if !isInvalid(t) && !c.isBoolType(t) {
		c.error(e.Pos(), "条件必须是 bool")
	}
}

func (c *Checker) checkReturn(s *ast.ReturnStmt) {
	if !c.inFunc {
		c.error(s.At, "这里不能 return")
		return
	}
	if !c.hasResult {
		if s.Value != nil {
			c.error(s.At, "函数没有返回值")
		}
		return
	}
	if s.Value == nil {
		c.error(s.At, "必须写出返回值")
		return
	}
	got := c.checkExprHint(s.Value, c.fnResult)
	if !c.assignableTo(got, c.fnResult) {
		c.error(s.At, "返回值类型不符")
	}
}

func (c *Checker) checkExprStmt(s *ast.ExprStmt) {
	call, ok := s.X.(*ast.CallExpr)
	if !ok {
		c.error(s.At, "不支持的语句")
		return
	}
	if c.isDelete(call) {
		c.checkDelete(call)
		return
	}
	c.evalCall(call)
}

func (c *Checker) checkSwitch(s *ast.SwitchStmt) {
	tag := c.checkExpr(s.Tag)
	info := &SwitchInfo{}
	if en, ok := c.enumInfoOf(tag); ok {
		info.Enum = true
		seen := map[string]bool{}
		hasDefault := false
		for _, cs := range s.Cases {
			if cs.IsDefault {
				hasDefault = true
				c.error(cs.At, "枚举 switch 不能写 default")
				c.checkBlock(cs.Body, true)
				continue
			}
			member, ok := c.enumMember(cs.Value, en)
			if !ok {
				c.error(cs.At, "枚举分支必须是枚举成员")
			} else if seen[member] {
				c.error(cs.At, "枚举分支重复")
			} else {
				seen[member] = true
			}
			c.checkBlock(cs.Body, true)
		}
		missing := false
		for _, m := range en.Members {
			if !seen[m] {
				missing = true
				break
			}
		}
		if missing {
			c.error(s.At, "枚举 switch 没有列全成员")
		}
		info.Exhaustive = !missing && !hasDefault
	} else {
		hasDefault := false
		for _, cs := range s.Cases {
			if cs.IsDefault {
				if hasDefault {
					c.error(cs.At, "default 重复")
				}
				hasDefault = true
			} else {
				ct := c.checkExprHint(cs.Value, tag)
				if !isInvalid(tag) && !c.assignableTo(ct, tag) {
					c.error(cs.At, "case 类型不符")
				}
			}
			c.checkBlock(cs.Body, true)
		}
		if !hasDefault {
			c.error(s.At, "switch 必须写 default")
		}
		info.Exhaustive = hasDefault
	}
	c.info.Switches[s] = info
}

func (c *Checker) enumMember(e ast.Expr, en *EnumInfo) (string, bool) {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		c.checkExpr(e)
		return "", false
	}
	c.checkExpr(sel)
	si := c.info.Sels[sel]
	if si == nil || si.Kind != SelEnum || !types.Identical(si.Typ, en.Typ) {
		return "", false
	}
	return si.Member, true
}

func (c *Checker) blockReturns(b *ast.Block) bool {
	if b == nil || len(b.List) == 0 {
		return false
	}
	return c.stmtReturns(b.List[len(b.List)-1])
}

func (c *Checker) stmtReturns(s ast.Stmt) bool {
	switch s := s.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.Block:
		return c.blockReturns(s)
	case *ast.IfStmt:
		if s.Else == nil || !c.blockReturns(s.Body) || !c.blockReturns(s.Else) {
			return false
		}
		for _, e := range s.Elifs {
			if !c.blockReturns(e.Body) {
				return false
			}
		}
		return true
	case *ast.SwitchStmt:
		info := c.info.Switches[s]
		if info == nil || !info.Exhaustive || len(s.Cases) == 0 {
			return false
		}
		for _, cs := range s.Cases {
			if !c.blockReturns(cs.Body) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (c *Checker) checkVar(d *ast.VarDecl, local bool) {
	if len(d.Specs) == 0 {
		return
	}
	init := &InitInfo{ErrIndex: -1}
	switch val := d.Value.(type) {
	case *ast.CallExpr:
		if c.isDelete(val) {
			c.error(d.At, "delete 只能当语句")
			c.checkDelete(val)
		} else {
			results := c.evalCall(val)
			c.matchLHS(d, results, init)
		}
	case *ast.IndexExpr:
		if len(d.Specs) == 2 && c.tryMapPair(d, val, init) {
			break
		}
		c.checkSingleVar(d, init)
	default:
		c.checkSingleVar(d, init)
	}
	c.info.Inits[d] = init
	c.bindVarSpecs(d, local)
	c.checkElse(d, init)
}

func (c *Checker) checkSingleVar(d *ast.VarDecl, init *InitInfo) {
	if len(d.Specs) != 1 {
		c.error(d.At, "只有调用函数或读取映射时才能一次接收多个值")
		for _, sp := range d.Specs {
			c.resolveType(sp.Type)
		}
		if d.Value != nil {
			c.checkExpr(d.Value)
		}
		return
	}
	hint := c.resolveType(d.Specs[0].Type)
	got := c.checkExprHint(d.Value, hint)
	if !c.assignableTo(got, hint) {
		c.error(d.Value.Pos(), "类型不符")
	}
	init.Results = []types.Type{got}
	if c.isErrorType(hint) {
		init.ErrIndex = 0
		init.ErrName = d.Specs[0].Name
	}
}

func (c *Checker) matchLHS(d *ast.VarDecl, results []types.Type, init *InitInfo) {
	init.Results = results
	init.ErrIndex = -1
	lhs := make([]types.Type, len(d.Specs))
	for i, sp := range d.Specs {
		lhs[i] = c.resolveType(sp.Type)
	}
	lastErr := len(results) > 0 && c.isErrorType(results[len(results)-1])
	if len(d.Specs) == len(results) {
		for i := range d.Specs {
			if !c.assignableTo(results[i], lhs[i]) {
				c.error(d.Specs[i].At, "返回值类型不符")
			}
		}
		if lastErr {
			init.ErrIndex = len(results) - 1
			init.ErrName = d.Specs[init.ErrIndex].Name
		}
		return
	}
	if len(results) == len(d.Specs)+1 && lastErr {
		for i := range d.Specs {
			if !c.assignableTo(results[i], lhs[i]) {
				c.error(d.Specs[i].At, "返回值类型不符")
			}
		}
		init.Omitted = true
		init.ErrIndex = len(results) - 1
		return
	}
	if !(len(results) == 1 && isInvalid(results[0])) {
		c.error(d.At, "返回值个数不符")
	}
}

func (c *Checker) tryMapPair(d *ast.VarDecl, idx *ast.IndexExpr, init *InitInfo) bool {
	base, isType := c.eval(idx.X)
	if isType || isInvalid(base) {
		return false
	}
	m, ok := types.Unalias(base).Underlying().(*types.Map)
	if !ok {
		return false
	}
	key := c.checkExprHint(idx.Index, m.Key())
	if !c.assignableTo(key, m.Key()) {
		c.error(idx.Index.Pos(), "键的类型不符")
	}
	vt := c.resolveType(d.Specs[0].Type)
	bt := c.resolveType(d.Specs[1].Type)
	if !c.assignableTo(m.Elem(), vt) {
		c.error(d.Specs[0].At, "返回值类型不符")
	}
	if !c.isBoolType(bt) {
		c.error(d.Specs[1].At, "映射的第二个值必须是 bool")
	}
	init.Results = []types.Type{m.Elem(), c.boolType}
	init.ErrIndex = -1
	c.info.Types[idx] = m.Elem()
	return true
}

func (c *Checker) bindVarSpecs(d *ast.VarDecl, local bool) {
	for _, sp := range d.Specs {
		obj := &Object{Name: sp.Name, Kind: ObjVar, Typ: c.resolveType(sp.Type), Local: local}
		c.bind(obj, sp.At)
		c.info.Specs[sp] = obj
	}
}

func (c *Checker) checkElse(d *ast.VarDecl, init *InitInfo) {
	if d.Else == nil {
		return
	}
	if !c.inFunc {
		c.error(d.Else.At, "这里不能 else return")
		return
	}
	if !c.hasResult {
		c.error(d.Else.At, "没有返回值的函数不能 else return")
		return
	}
	if init.ErrIndex < 0 {
		c.error(d.Else.At, "这里没有 error，不能 else return")
		return
	}
	if d.Else.Bare {
		if !init.Omitted {
			c.error(d.Else.At, "只有省略了 error 时才能写 else return")
			return
		}
		if !c.isErrorType(c.fnResult) {
			c.error(d.Else.At, "函数返回类型不是 error 时不能写 else return")
		}
		return
	}
	if d.Else.Value == nil {
		c.error(d.Else.At, "函数返回类型不是 error 时不能写 else return")
		return
	}
	got := c.checkExprHint(d.Else.Value, c.fnResult)
	if !c.assignableTo(got, c.fnResult) {
		c.error(d.Else.At, "else return 的类型与函数返回类型不符")
	}
}

func (c *Checker) checkAssign(s *ast.AssignStmt) {
	lt, ok := c.assignTarget(s.Target)
	if !ok {
		if s.Value != nil {
			c.checkExpr(s.Value)
		}
		return
	}
	rt := c.checkExprHint(s.Value, lt)
	if !c.assignableTo(rt, lt) {
		c.error(s.At, "不能赋值给这个类型")
	}
}

func (c *Checker) assignTarget(e ast.Expr) (types.Type, bool) {
	switch e := e.(type) {
	case *ast.Ident:
		obj := c.scope.lookup(e.Name)
		if obj == nil {
			c.error(e.At, "未定义的名字 "+e.Name)
			return types.Typ[types.Invalid], false
		}
		c.info.Idents[e] = obj
		if obj.Kind != ObjVar && obj.Kind != ObjParam {
			c.error(e.At, "只能给变量赋值")
			return obj.Typ, false
		}
		return obj.Typ, true
	case *ast.IndexExpr:
		base, isType := c.eval(e.X)
		if isType || isInvalid(base) {
			c.checkExpr(e.Index)
			return types.Typ[types.Invalid], false
		}
		switch u := types.Unalias(base).Underlying().(type) {
		case *types.Map:
			key := c.checkExprHint(e.Index, u.Key())
			if !c.assignableTo(key, u.Key()) {
				c.error(e.Index.Pos(), "键的类型不符")
			}
			return u.Elem(), true
		case *types.Slice:
			key := c.checkExpr(e.Index)
			if !c.isInteger(key) {
				c.error(e.Index.Pos(), "下标必须是整数")
			}
			return u.Elem(), true
		default:
			c.checkExpr(e.Index)
			c.error(e.At, "不能赋值")
			return types.Typ[types.Invalid], false
		}
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			if obj := c.scope.lookup(id.Name); obj != nil && (obj.Kind == ObjImport || (obj.Kind == ObjType && c.enumName[id.Name] != nil)) {
				c.error(e.At, "不能赋值")
				return types.Typ[types.Invalid], false
			}
		}
		base, isType := c.eval(e.X)
		if isType || isInvalid(base) {
			return types.Typ[types.Invalid], false
		}
		obj, _, _ := types.LookupFieldOrMethod(base, true, c.typesPkg, e.Sel)
		field, ok := obj.(*types.Var)
		if !ok {
			c.error(e.At, "只能给字段赋值")
			return types.Typ[types.Invalid], false
		}
		goName := field.Name()
		if field.Pkg() == c.typesPkg {
			goName = names.Export(field.Name())
		}
		c.info.Sels[e] = &SelInfo{Kind: SelField, GoName: goName, Typ: field.Type()}
		c.info.Types[e] = field.Type()
		return field.Type(), true
	default:
		c.error(e.Pos(), "不能赋值")
		return types.Typ[types.Invalid], false
	}
}
