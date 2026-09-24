package check

import (
	"go/types"

	"garnet/compiler/ast"
	"garnet/compiler/names"
	"garnet/compiler/token"
)

func (c *Checker) checkExpr(e ast.Expr) types.Type {
	t, isType := c.eval(e)
	if isType {
		c.error(e.Pos(), "不能把类型当值使用")
		return types.Typ[types.Invalid]
	}
	return t
}

func (c *Checker) checkExprHint(e ast.Expr, hint types.Type) types.Type {
	if e == nil {
		return types.Typ[types.Invalid]
	}
	switch e := e.(type) {
	case *ast.SliceLit:
		return c.checkSliceLit(e, hint)
	case *ast.MapLit:
		return c.checkMapLit(e, hint)
	default:
		return c.checkExpr(e)
	}
}

func (c *Checker) eval(e ast.Expr) (types.Type, bool) {
	if e == nil {
		return types.Typ[types.Invalid], false
	}
	t, isType := c.evalRaw(e)
	if t == nil {
		t = types.Typ[types.Invalid]
	}
	c.info.Types[e] = t
	if isType {
		c.info.IsType[e] = true
	}
	return t, isType
}

func (c *Checker) evalRaw(e ast.Expr) (types.Type, bool) {
	switch e := e.(type) {
	case *ast.Ident:
		return c.evalIdent(e)
	case *ast.BasicLit:
		return c.evalBasic(e)
	case *ast.BinaryExpr:
		return c.evalBinary(e)
	case *ast.UnaryExpr:
		return c.evalUnary(e)
	case *ast.CallExpr:
		rs := c.evalCall(e)
		if len(rs) == 1 {
			return rs[0], false
		}
		if len(rs) == 0 {
			c.error(e.At, "这个函数没有返回值")
			return types.Typ[types.Invalid], false
		}
		c.error(e.At, "这里只能有一个返回值")
		return types.Typ[types.Invalid], false
	case *ast.SelectorExpr:
		return c.evalSelector(e)
	case *ast.IndexExpr:
		return c.evalIndex(e)
	case *ast.SliceLit:
		return c.checkSliceLit(e, nil), false
	case *ast.MapLit:
		return c.checkMapLit(e, nil), false
	case *ast.StructLit:
		return c.evalStructLit(e)
	case *ast.InterfaceLit:
		return c.evalInterfaceLit(e)
	default:
		c.error(e.Pos(), "不支持的表达式")
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalIdent(id *ast.Ident) (types.Type, bool) {
	if id.Name == "" || id.Name == "_bad" {
		return types.Typ[types.Invalid], false
	}
	obj := c.scope.lookup(id.Name)
	if obj == nil {
		c.error(id.At, "未定义的名字 "+id.Name)
		return types.Typ[types.Invalid], false
	}
	c.info.Idents[id] = obj
	switch obj.Kind {
	case ObjVar, ObjConst:
		obj.Used = true
		return obj.Typ, false
	case ObjParam, ObjFunc:
		return obj.Typ, false
	case ObjType:
		return obj.Typ, true
	case ObjImport:
		c.error(id.At, "包名不能当值使用")
		return types.Typ[types.Invalid], false
	default:
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalBasic(e *ast.BasicLit) (types.Type, bool) {
	switch e.Kind {
	case token.Int:
		return types.Typ[types.UntypedInt], false
	case token.Float:
		return types.Typ[types.UntypedFloat], false
	case token.String:
		return types.Typ[types.UntypedString], false
	case token.True, token.False:
		return types.Typ[types.UntypedBool], false
	case token.Nil:
		return types.Typ[types.UntypedNil], false
	default:
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalUnary(e *ast.UnaryExpr) (types.Type, bool) {
	switch e.Op {
	case token.And:
		t, isType := c.eval(e.X)
		if isType {
			c.error(e.At, "不能取类型的地址")
			return types.Typ[types.Invalid], false
		}
		if !isInvalid(t) && !c.addrOK(e.X) {
			c.error(e.At, "不能取地址")
		}
		if isInvalid(t) {
			return types.Typ[types.Invalid], false
		}
		return types.NewPointer(t), false
	case token.Not:
		t := c.checkExpr(e.X)
		if !isInvalid(t) && !c.isBoolType(t) {
			c.error(e.At, "条件必须是 bool")
		}
		return c.boolType, false
	case token.Sub:
		t, isType := c.eval(e.X)
		if isType || (!isInvalid(t) && (c.isEnumType(t) || !c.isNumeric(t))) {
			if !isInvalid(t) {
				c.error(e.At, "不能取负")
			}
			return types.Typ[types.Invalid], false
		}
		return t, false
	default:
		c.error(e.At, "不支持的运算符")
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) addrOK(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.Ident:
		obj := c.scope.lookup(e.Name)
		return obj != nil && (obj.Kind == ObjVar || obj.Kind == ObjParam)
	case *ast.SelectorExpr:
		si := c.info.Sels[e]
		if si == nil || si.Kind != SelField {
			return false
		}
		if isPointerType(c.info.Types[e.X]) {
			return true
		}
		return c.addrOK(e.X)
	case *ast.IndexExpr:
		bt := c.info.Types[e.X]
		if bt == nil {
			return false
		}
		_, ok := types.Unalias(bt).Underlying().(*types.Slice)
		return ok && c.addrOK(e.X)
	default:
		return false
	}
}

func (c *Checker) evalBinary(e *ast.BinaryExpr) (types.Type, bool) {
	xt, xIsType := c.eval(e.X)
	yt, yIsType := c.eval(e.Y)
	if xIsType || yIsType {
		c.error(e.At, "不能把类型当值使用")
		return types.Typ[types.Invalid], false
	}
	switch e.Op {
	case token.Land, token.Lor:
		if (!isInvalid(xt) && !c.isBoolType(xt)) || (!isInvalid(yt) && !c.isBoolType(yt)) {
			c.error(e.At, "条件必须是 bool")
		}
		return c.boolType, false
	case token.Eql, token.Neq:
		if !c.comparable(xt, yt) {
			c.error(e.At, "不能比较")
		}
		return c.boolType, false
	case token.Lss, token.Gtr, token.Leq, token.Geq:
		if _, ok := c.ordered(xt, yt); !ok {
			c.mixError(e.At, xt, yt, "不能比较大小")
		}
		return c.boolType, false
	case token.Add:
		if c.isStringish(xt) && c.isStringish(yt) {
			t, ok := c.combine(xt, yt)
			if !ok {
				c.mixError(e.At, xt, yt, "运算的类型不符")
				return types.Typ[types.Invalid], false
			}
			return t, false
		}
		fallthrough
	default:
		t, ok := c.arith(xt, yt, e.Op == token.Rem)
		if !ok {
			c.mixError(e.At, xt, yt, "运算的类型不符")
			return types.Typ[types.Invalid], false
		}
		return t, false
	}
}

func (c *Checker) mixError(pos token.Position, x, y types.Type, plain string) {
	if c.isNewtype(x) || c.isNewtype(y) || c.isEnumType(x) || c.isEnumType(y) {
		c.error(pos, "不能直接混算")
		return
	}
	c.error(pos, plain)
}

func (c *Checker) evalSelector(e *ast.SelectorExpr) (types.Type, bool) {
	if id, ok := e.X.(*ast.Ident); ok {
		if obj := c.scope.lookup(id.Name); obj != nil && obj.Kind == ObjImport {
			return c.evalPkgSel(c.imports[id.Name], e)
		}
		if obj := c.scope.lookup(id.Name); obj != nil && obj.Kind == ObjType {
			if en := c.enumName[id.Name]; en != nil {
				return c.evalEnumSel(en, e)
			}
		}
	}
	base, isType := c.eval(e.X)
	if isType {
		c.error(e.At, "不能把类型当值使用")
		return types.Typ[types.Invalid], false
	}
	return c.evalField(base, e)
}

func (c *Checker) evalPkgSel(imp *ImportInfo, e *ast.SelectorExpr) (types.Type, bool) {
	if imp == nil || imp.Pkg == nil {
		c.error(e.At, "未知的包")
		return types.Typ[types.Invalid], false
	}
	imp.Used = true
	obj := imp.Pkg.Scope().Lookup(e.Sel)
	if obj == nil {
		c.error(e.At, "包里没有 "+e.Sel)
		return types.Typ[types.Invalid], false
	}
	switch o := obj.(type) {
	case *types.Func:
		c.info.Sels[e] = &SelInfo{Kind: SelFunc, GoName: e.Sel, Typ: o.Type()}
		return o.Type(), false
	case *types.TypeName:
		c.info.Sels[e] = &SelInfo{Kind: SelType, GoName: e.Sel, Typ: o.Type()}
		return o.Type(), true
	case *types.Var:
		c.info.Sels[e] = &SelInfo{Kind: SelPkgVal, GoName: e.Sel, Typ: o.Type()}
		return o.Type(), false
	case *types.Const:
		c.info.Sels[e] = &SelInfo{Kind: SelPkgVal, GoName: e.Sel, Typ: o.Type()}
		return o.Type(), false
	default:
		c.error(e.At, "不能使用 "+e.Sel)
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalEnumSel(en *EnumInfo, e *ast.SelectorExpr) (types.Type, bool) {
	found := false
	for _, m := range en.Members {
		if m == e.Sel {
			found = true
			break
		}
	}
	if !found {
		c.error(e.At, "枚举没有成员 "+e.Sel)
		return types.Typ[types.Invalid], false
	}
	c.info.Sels[e] = &SelInfo{Kind: SelEnum, GoName: en.Name + "_" + e.Sel, Member: e.Sel, Typ: en.Typ}
	return en.Typ, false
}

func (c *Checker) evalField(base types.Type, e *ast.SelectorExpr) (types.Type, bool) {
	if isInvalid(base) {
		return types.Typ[types.Invalid], false
	}
	obj, _, _ := types.LookupFieldOrMethod(base, true, c.typesPkg, e.Sel)
	if obj == nil {
		c.error(e.At, "没有字段或方法 "+e.Sel)
		return types.Typ[types.Invalid], false
	}
	switch o := obj.(type) {
	case *types.Var:
		goName := o.Name()
		if o.Pkg() == c.typesPkg {
			goName = names.Export(o.Name())
		}
		c.info.Sels[e] = &SelInfo{Kind: SelField, GoName: goName, Typ: o.Type()}
		return o.Type(), false
	case *types.Func:
		sig := o.Type().(*types.Signature)
		c.info.Sels[e] = &SelInfo{Kind: SelMethod, GoName: o.Name(), Typ: sig}
		return sig, false
	default:
		c.error(e.At, "没有字段或方法 "+e.Sel)
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalIndex(e *ast.IndexExpr) (types.Type, bool) {
	base, isType := c.eval(e.X)
	if isType {
		c.error(e.At, "不能把类型当值使用")
		return types.Typ[types.Invalid], false
	}
	if isInvalid(base) {
		c.checkExpr(e.Index)
		return types.Typ[types.Invalid], false
	}
	switch u := types.Unalias(base).Underlying().(type) {
	case *types.Map:
		key := c.checkExprHint(e.Index, u.Key())
		if !c.assignableTo(key, u.Key()) {
			c.error(e.Index.Pos(), "键的类型不符")
		}
		return u.Elem(), false
	case *types.Slice:
		key := c.checkExpr(e.Index)
		if !isInvalid(key) && !c.isInteger(key) {
			c.error(e.Index.Pos(), "下标必须是整数")
		}
		return u.Elem(), false
	default:
		c.checkExpr(e.Index)
		c.error(e.At, "不能用下标读取")
		return types.Typ[types.Invalid], false
	}
}

func (c *Checker) evalCall(call *ast.CallExpr) []types.Type {
	if c.isDelete(call) {
		c.error(call.At, "delete 只能当语句")
		return []types.Type{types.Typ[types.Invalid]}
	}
	funT, isType := c.eval(call.Fun)
	if isType {
		if !c.isGoType(call.Fun) {
			c.error(call.At, "不能把这个类型当函数调用")
			return []types.Type{types.Typ[types.Invalid]}
		}
		if len(call.Args) != 1 {
			c.error(call.At, "类型转换只能有一个参数")
			return []types.Type{funT}
		}
		arg := c.checkExpr(call.Args[0])
		if !c.convertible(arg, funT) {
			c.error(call.At, "不能转换成这个类型")
		}
		c.info.Calls[call] = &CallInfo{Results: []types.Type{funT}, Convert: true}
		return []types.Type{funT}
	}
	sig, ok := underlyingSig(funT)
	if !ok {
		if !isInvalid(funT) {
			c.error(call.At, "只能调用函数")
		}
		return []types.Type{types.Typ[types.Invalid]}
	}
	stringToBytes := c.matchArgs(sig, call)
	var results []types.Type
	if sig.Results() != nil {
		for i := 0; i < sig.Results().Len(); i++ {
			results = append(results, sig.Results().At(i).Type())
		}
	}
	c.info.Calls[call] = &CallInfo{Results: results, StringToBytes: stringToBytes}
	return results
}

func underlyingSig(t types.Type) (*types.Signature, bool) {
	if t == nil || isInvalid(t) {
		return nil, false
	}
	sig, ok := types.Unalias(t).Underlying().(*types.Signature)
	return sig, ok
}

func (c *Checker) isGoType(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		si := c.info.Sels[e]
		return si != nil && si.Kind == SelType
	case *ast.Ident:
		if obj := c.info.Idents[e]; obj != nil && obj.Kind == ObjType {
			return false
		}
		return builtinType(e.Name) != nil
	default:
		return false
	}
}

func (c *Checker) isJSONUnmarshal(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel != "Unmarshal" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	imp := c.imports[id.Name]
	return imp != nil && imp.Path == "encoding/json"
}

func (c *Checker) matchArgs(sig *types.Signature, call *ast.CallExpr) bool {
	params := sig.Params()
	n := 0
	if params != nil {
		n = params.Len()
	}
	args := call.Args
	jsonU := c.isJSONUnmarshal(call.Fun)
	stringToBytes := false
	if sig.Variadic() {
		if n == 0 || len(args) < n-1 {
			c.error(call.At, "参数个数不对")
			return false
		}
		for i := 0; i < n-1; i++ {
			c.checkArg(args[i], params.At(i).Type(), false)
		}
		var elem types.Type = types.Typ[types.Invalid]
		if slice, ok := params.At(n - 1).Type().Underlying().(*types.Slice); ok {
			elem = slice.Elem()
		}
		for i := n - 1; i < len(args); i++ {
			c.checkArg(args[i], elem, false)
		}
		return false
	}
	if len(args) != n {
		c.error(call.At, "参数个数不对")
		return false
	}
	for i := 0; i < n; i++ {
		hint := params.At(i).Type()
		allow := jsonU && i == 0 && isByteSlice(hint)
		if c.checkArg(args[i], hint, allow) {
			stringToBytes = true
		}
	}
	return stringToBytes
}

func (c *Checker) checkArg(arg ast.Expr, hint types.Type, allowString bool) bool {
	if allowString {
		got := c.checkExpr(arg)
		if c.assignableTo(got, hint) {
			return false
		}
		if c.assignableTo(got, c.stringTyp) {
			return true
		}
		if !isInvalid(got) {
			c.error(arg.Pos(), "参数类型不符")
		}
		return false
	}
	got := c.checkExprHint(arg, hint)
	if !c.assignableTo(got, hint) {
		c.error(arg.Pos(), "参数类型不符")
	}
	return false
}

func (c *Checker) isDelete(call *ast.CallExpr) bool {
	id, ok := call.Fun.(*ast.Ident)
	if !ok || id.Name != "delete" {
		return false
	}
	obj := c.scope.lookup("delete")
	return obj == nil
}

func (c *Checker) checkDelete(call *ast.CallExpr) {
	if len(call.Args) != 2 {
		c.error(call.At, "delete 需要映射和键")
		return
	}
	base := c.checkExpr(call.Args[0])
	m, ok := types.Unalias(base).Underlying().(*types.Map)
	if !ok {
		if !isInvalid(base) {
			c.error(call.Args[0].Pos(), "delete 的第一个参数必须是映射")
		}
		c.checkExpr(call.Args[1])
		return
	}
	key := c.checkExprHint(call.Args[1], m.Key())
	if !c.assignableTo(key, m.Key()) {
		c.error(call.Args[1].Pos(), "delete 的键类型不符")
	}
	c.info.Calls[call] = &CallInfo{Delete: true}
}

func (c *Checker) checkSliceLit(e *ast.SliceLit, hint types.Type) types.Type {
	var elem types.Type
	var typ types.Type
	if hint != nil {
		if s, ok := types.Unalias(hint).Underlying().(*types.Slice); ok && !isInvalid(hint) {
			elem = s.Elem()
			typ = hint
		}
	}
	if elem == nil {
		if len(e.Elts) == 0 {
			c.error(e.At, "切片字面量需要类型")
			return types.Typ[types.Invalid]
		}
		elem = c.defaultType(c.checkExpr(e.Elts[0]))
		for _, el := range e.Elts[1:] {
			got := c.checkExprHint(el, elem)
			if !c.assignableTo(got, elem) {
				c.error(el.Pos(), "切片元素类型不符")
			}
		}
		typ = types.NewSlice(elem)
	} else {
		for _, el := range e.Elts {
			got := c.checkExprHint(el, elem)
			if !c.assignableTo(got, elem) {
				c.error(el.Pos(), "切片元素类型不符")
			}
		}
	}
	c.info.Types[e] = typ
	return typ
}

func (c *Checker) checkMapLit(e *ast.MapLit, hint types.Type) types.Type {
	var keyT, valT, typ types.Type
	if hint != nil {
		if m, ok := types.Unalias(hint).Underlying().(*types.Map); ok && !isInvalid(hint) {
			keyT, valT, typ = m.Key(), m.Elem(), hint
		}
	}
	if keyT == nil {
		if len(e.Entries) == 0 {
			c.error(e.At, "映射字面量需要类型")
			return types.Typ[types.Invalid]
		}
		keyT = c.defaultType(c.checkExpr(e.Entries[0].Key))
		valT = c.defaultType(c.checkExpr(e.Entries[0].Value))
		if !isInvalid(keyT) && !types.Comparable(keyT) {
			c.error(e.At, "映射的键必须能比较")
		}
		for _, ent := range e.Entries[1:] {
			k := c.checkExprHint(ent.Key, keyT)
			v := c.checkExprHint(ent.Value, valT)
			if !c.assignableTo(k, keyT) {
				c.error(ent.Key.Pos(), "键的类型不符")
			}
			if !c.assignableTo(v, valT) {
				c.error(ent.Value.Pos(), "值的类型不符")
			}
		}
		typ = types.NewMap(keyT, valT)
	} else {
		for _, ent := range e.Entries {
			k := c.checkExprHint(ent.Key, keyT)
			v := c.checkExprHint(ent.Value, valT)
			if !c.assignableTo(k, keyT) {
				c.error(ent.Key.Pos(), "键的类型不符")
			}
			if !c.assignableTo(v, valT) {
				c.error(ent.Value.Pos(), "值的类型不符")
			}
		}
	}
	c.info.Types[e] = typ
	return typ
}

func (c *Checker) evalStructLit(e *ast.StructLit) (types.Type, bool) {
	typ, isType := c.eval(e.Type)
	if !isType {
		c.error(e.At, "这里要写类型名")
		return types.Typ[types.Invalid], false
	}
	if _, ok := c.interfaceOf(typ); ok {
		if len(e.Fields) == 0 {
			if iface, _ := c.interfaceOf(typ); iface.NumMethods() > 0 {
				c.error(e.At, "接口组装缺少函数")
			}
		} else {
			c.error(e.At, "接口要用函数组装")
		}
		return typ, false
	}
	st, ok := c.structOf(typ)
	if !ok {
		c.error(e.At, "不是结构体")
		return typ, false
	}
	fields := map[string]*types.Var{}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		fields[f.Name()] = f
	}
	seen := map[string]bool{}
	for _, init := range e.Fields {
		f, ok := fields[init.Name]
		if !ok {
			c.error(init.At, "没有字段 "+init.Name)
			c.checkExpr(init.Value)
			continue
		}
		if seen[init.Name] {
			c.error(init.At, "字段重复")
		}
		seen[init.Name] = true
		got := c.checkExprHint(init.Value, f.Type())
		if !c.assignableTo(got, f.Type()) {
			c.error(init.At, "字段类型不符")
		}
	}
	return typ, false
}

func (c *Checker) evalInterfaceLit(e *ast.InterfaceLit) (types.Type, bool) {
	typ, isType := c.eval(e.Type)
	if !isType {
		c.error(e.At, "这里要写类型名")
		return types.Typ[types.Invalid], false
	}
	iface, ok := c.interfaceOf(typ)
	if !ok {
		c.error(e.At, "不是接口")
		return typ, false
	}
	have := map[string]*ast.FuncLit{}
	for _, m := range e.Methods {
		if _, ok := have[m.Name]; ok {
			c.error(m.NamePos, "函数重复")
		}
		have[m.Name] = m
	}
	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		lit, ok := have[method.Name()]
		if !ok {
			c.error(e.At, "接口组装缺少函数 "+method.Name())
			continue
		}
		sig, _ := method.Type().Underlying().(*types.Signature)
		if sig != nil {
			c.checkFuncLit(lit, sig)
		}
		delete(have, method.Name())
	}
	for name, lit := range have {
		c.error(lit.NamePos, "接口没有函数 "+name)
	}
	return typ, false
}

func (c *Checker) checkFuncLit(lit *ast.FuncLit, sig *types.Signature) {
	c.checkName(lit.Name, lit.NamePos)
	nparam := 0
	if sig.Params() != nil {
		nparam = sig.Params().Len()
	}
	if nparam != len(lit.Params) {
		c.error(lit.At, "参数个数不符")
	}
	n := nparam
	if n > len(lit.Params) {
		n = len(lit.Params)
	}
	for i := 0; i < n; i++ {
		pt := c.resolveType(lit.Params[i].Type)
		if !types.Identical(pt, sig.Params().At(i).Type()) {
			c.error(lit.Params[i].At, "参数类型不符")
		}
	}
	var want types.Type
	if sig.Results() != nil && sig.Results().Len() > 0 {
		want = sig.Results().At(0).Type()
	}
	if lit.Result == nil && want != nil {
		c.error(lit.At, "缺少返回类型")
	}
	if lit.Result != nil && want == nil {
		c.error(lit.At, "接口函数没有返回值")
	}
	if lit.Result != nil && want != nil && !types.Identical(c.resolveType(lit.Result), want) {
		c.error(lit.Result.Pos(), "返回类型不符")
	}
	prevIn, prevHas, prevRes := c.inFunc, c.hasResult, c.fnResult
	c.openScope()
	c.inFunc = true
	c.hasResult = want != nil
	c.fnResult = want
	for i, p := range lit.Params {
		var pt types.Type = types.Typ[types.Invalid]
		if sig.Params() != nil && i < sig.Params().Len() {
			pt = sig.Params().At(i).Type()
		}
		c.bind(&Object{Name: p.Name, Kind: ObjParam, Typ: pt, Local: true}, p.At)
	}
	c.checkBlock(lit.Body, false)
	if c.hasResult && !c.blockReturns(lit.Body) {
		c.error(lit.At, "缺少 return")
	}
	c.closeScope()
	c.inFunc, c.hasResult, c.fnResult = prevIn, prevHas, prevRes
}

func (c *Checker) structOf(t types.Type) (*types.Struct, bool) {
	if t == nil || isInvalid(t) {
		return nil, false
	}
	st, ok := types.Unalias(t).Underlying().(*types.Struct)
	return st, ok
}

func (c *Checker) interfaceOf(t types.Type) (*types.Interface, bool) {
	if t == nil || isInvalid(t) {
		return nil, false
	}
	iface, ok := types.Unalias(t).Underlying().(*types.Interface)
	return iface, ok
}

func (c *Checker) enumInfoOf(t types.Type) (*EnumInfo, bool) {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil, false
	}
	en, ok := c.enums[n]
	return en, ok
}

func isPointerType(t types.Type) bool {
	if t == nil || isInvalid(t) {
		return false
	}
	_, ok := types.Unalias(t).Underlying().(*types.Pointer)
	return ok
}
