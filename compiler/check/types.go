package check

import "go/types"

func isInvalid(t types.Type) bool {
	if t == nil {
		return true
	}
	u := t.Underlying()
	if u == nil {
		return true
	}
	b, ok := u.(*types.Basic)
	return ok && b.Kind() == types.Invalid
}

func isUntyped(t types.Type) bool {
	b, ok := types.Unalias(t).(*types.Basic)
	return ok && b.Info()&types.IsUntyped != 0
}

func isUntypedNil(t types.Type) bool {
	b, ok := types.Unalias(t).(*types.Basic)
	return ok && b.Kind() == types.UntypedNil
}

func isNillable(t types.Type) bool {
	if t == nil || isInvalid(t) {
		return false
	}
	switch types.Unalias(t).Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Interface, *types.Signature, *types.Chan:
		return true
	default:
		return false
	}
}

func isByteSlice(t types.Type) bool {
	if t == nil || isInvalid(t) {
		return false
	}
	s, ok := t.Underlying().(*types.Slice)
	if !ok {
		return false
	}
	b, ok := types.Unalias(s.Elem()).Underlying().(*types.Basic)
	return ok && (b.Kind() == types.Uint8 || b.Kind() == types.Byte)
}

func isNumericKind(k types.BasicKind) bool {
	switch k {
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
		types.Float32, types.Float64, types.Complex64, types.Complex128,
		types.UntypedInt, types.UntypedFloat, types.UntypedComplex, types.UntypedRune:
		return true
	default:
		return false
	}
}

func isIntKind(k types.BasicKind) bool {
	switch k {
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
		types.UntypedInt, types.UntypedRune:
		return true
	default:
		return false
	}
}

func isFloatKind(k types.BasicKind) bool {
	switch k {
	case types.Float32, types.Float64, types.UntypedFloat:
		return true
	default:
		return false
	}
}

func isComplexKind(k types.BasicKind) bool {
	switch k {
	case types.Complex64, types.Complex128, types.UntypedComplex:
		return true
	default:
		return false
	}
}

func (c *Checker) basicKind(t types.Type) types.BasicKind {
	if t == nil {
		return types.Invalid
	}
	t = types.Unalias(t)
	u := t.Underlying()
	if u == nil {
		return types.Invalid
	}
	b, ok := u.(*types.Basic)
	if !ok {
		return types.Invalid
	}
	return b.Kind()
}

func (c *Checker) isEnumType(t types.Type) bool {
	_, ok := c.enumInfoOf(t)
	return ok
}

func (c *Checker) isNewtype(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || c.isEnumType(t) {
		return false
	}
	switch n.Underlying().(type) {
	case *types.Struct, *types.Interface:
		return false
	default:
		return n.Underlying() != nil
	}
}

func (c *Checker) isErrorType(t types.Type) bool {
	return t != nil && !isInvalid(t) && types.Identical(t, c.errorType)
}

func (c *Checker) isBoolType(t types.Type) bool {
	if t == nil || c.isEnumType(t) {
		return false
	}
	k := c.basicKind(t)
	return k == types.Bool || k == types.UntypedBool
}

func (c *Checker) isNumeric(t types.Type) bool {
	if t == nil || c.isEnumType(t) {
		return false
	}
	return isNumericKind(c.basicKind(t))
}

func (c *Checker) isInteger(t types.Type) bool {
	if t == nil || c.isEnumType(t) {
		return false
	}
	return isIntKind(c.basicKind(t))
}

func (c *Checker) isStringish(t types.Type) bool {
	if t == nil || c.isEnumType(t) || c.isNewtype(t) {
		return false
	}
	k := c.basicKind(t)
	return k == types.String || k == types.UntypedString
}

func (c *Checker) isOrderedType(t types.Type) bool {
	if c.isStringish(t) || (c.isNumeric(t) && !isComplexKind(c.basicKind(t))) {
		return true
	}
	return false
}

func (c *Checker) defaultType(t types.Type) types.Type {
	b, ok := types.Unalias(t).(*types.Basic)
	if !ok {
		return t
	}
	switch b.Kind() {
	case types.UntypedInt:
		return types.Typ[types.Int]
	case types.UntypedFloat:
		return types.Typ[types.Float64]
	case types.UntypedString:
		return types.Typ[types.String]
	case types.UntypedBool:
		return types.Typ[types.Bool]
	case types.UntypedRune:
		return types.Universe.Lookup("rune").Type()
	default:
		return t
	}
}

func (c *Checker) assignableTo(v, t types.Type) bool {
	if isInvalid(v) || isInvalid(t) {
		return true
	}
	if c.isEnumType(t) || c.isEnumType(v) {
		return types.Identical(v, t)
	}
	if isUntyped(v) {
		return c.untypedAssignable(v, t)
	}
	return types.AssignableTo(v, t)
}

func (c *Checker) untypedAssignable(v, t types.Type) bool {
	b, ok := types.Unalias(v).(*types.Basic)
	if !ok {
		return false
	}
	// 赋给接口（如 fmt.Println 的 ...any）时，按常量的默认类型判断。
	if _, isIface := t.Underlying().(*types.Interface); isIface {
		if b.Kind() == types.UntypedNil {
			return true
		}
		return types.AssignableTo(types.Default(v), t)
	}
	switch b.Kind() {
	case types.UntypedNil:
		return isNillable(t)
	case types.UntypedBool:
		return c.basicKind(t) == types.Bool
	case types.UntypedString:
		return c.basicKind(t) == types.String
	case types.UntypedInt, types.UntypedRune:
		k := c.basicKind(t)
		return isIntKind(k) || isFloatKind(k) || isComplexKind(k)
	case types.UntypedFloat:
		k := c.basicKind(t)
		return isFloatKind(k) || isComplexKind(k)
	case types.UntypedComplex:
		return isComplexKind(c.basicKind(t))
	default:
		return false
	}
}

func (c *Checker) convertible(from, to types.Type) bool {
	if c.assignableTo(from, to) {
		return true
	}
	if isInvalid(from) || isInvalid(to) {
		return true
	}
	if c.isStringish(from) && isByteSlice(to.Underlying()) {
		return true
	}
	if types.Identical(from.Underlying(), to.Underlying()) && !c.isNewtype(from) && !c.isNewtype(to) {
		return true
	}
	return false
}

func (c *Checker) comparable(x, y types.Type) bool {
	if isInvalid(x) || isInvalid(y) {
		return true
	}
	if c.isEnumType(x) || c.isEnumType(y) {
		return types.Identical(x, y)
	}
	if isUntypedNil(x) {
		return isNillable(y)
	}
	if isUntypedNil(y) {
		return isNillable(x)
	}
	if isUntyped(x) && c.assignableTo(x, y) {
		return types.Comparable(c.defaultType(y))
	}
	if isUntyped(y) && c.assignableTo(y, x) {
		return types.Comparable(c.defaultType(x))
	}
	if types.Identical(x, y) {
		return types.Comparable(types.Unalias(x).Underlying())
	}
	return false
}

func (c *Checker) ordered(x, y types.Type) (types.Type, bool) {
	if isInvalid(x) || isInvalid(y) {
		return c.boolType, true
	}
	if !c.isOrderedType(x) || !c.isOrderedType(y) {
		return nil, false
	}
	if c.isStringish(x) || c.isStringish(y) {
		return c.combine(x, y)
	}
	return c.combineNumeric(x, y)
}

func (c *Checker) arith(x, y types.Type, integerOnly bool) (types.Type, bool) {
	if isInvalid(x) || isInvalid(y) {
		return types.Typ[types.Invalid], true
	}
	if c.isEnumType(x) || c.isEnumType(y) {
		return nil, false
	}
	if integerOnly {
		if !c.isInteger(x) || !c.isInteger(y) {
			return nil, false
		}
	} else if !c.isNumeric(x) || !c.isNumeric(y) {
		return nil, false
	}
	return c.combineNumeric(x, y)
}

func (c *Checker) combine(x, y types.Type) (types.Type, bool) {
	if isInvalid(x) || isInvalid(y) {
		return types.Typ[types.Invalid], true
	}
	if isUntyped(x) && isUntyped(y) {
		if c.basicKind(x) == types.UntypedString || c.basicKind(y) == types.UntypedString {
			return types.Typ[types.UntypedString], true
		}
	}
	if isUntyped(x) && c.assignableTo(x, y) {
		return y, true
	}
	if isUntyped(y) && c.assignableTo(y, x) {
		return x, true
	}
	if types.Identical(x, y) {
		return x, true
	}
	return nil, false
}

func (c *Checker) combineNumeric(x, y types.Type) (types.Type, bool) {
	if isUntyped(x) && isUntyped(y) {
		if isFloatKind(c.basicKind(x)) || isFloatKind(c.basicKind(y)) {
			return types.Typ[types.UntypedFloat], true
		}
		if isComplexKind(c.basicKind(x)) || isComplexKind(c.basicKind(y)) {
			return types.Typ[types.UntypedComplex], true
		}
		return types.Typ[types.UntypedInt], true
	}
	if isUntyped(x) && c.assignableTo(x, y) {
		return y, true
	}
	if isUntyped(y) && c.assignableTo(y, x) {
		return x, true
	}
	if types.Identical(x, y) {
		return x, true
	}
	return nil, false
}
