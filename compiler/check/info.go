package check

import (
	gotok "go/token"
	"go/types"

	"garnet/compiler/ast"
	"garnet/compiler/token"
)

type ObjKind int

const (
	ObjVar ObjKind = iota
	ObjConst
	ObjParam
	ObjFunc
	ObjType
	ObjImport
)

type Object struct {
	Name  string
	Kind  ObjKind
	Typ   types.Type
	Used  bool
	Local bool
}

type SelKind int

const (
	SelField SelKind = iota
	SelMethod
	SelFunc
	SelPkgVal
	SelType
	SelEnum
)

type SelInfo struct {
	Kind   SelKind
	GoName string
	Member string
	Typ    types.Type
}

type CallInfo struct {
	Results       []types.Type
	StringToBytes bool
	Delete        bool
	Convert       bool
}

type InitInfo struct {
	Omitted  bool
	ErrIndex int
	ErrName  string
	Results  []types.Type
}

type ImportInfo struct {
	Local string
	Path  string
	Name  string
	Pkg   *types.Package
	Used  bool
}

type EnumInfo struct {
	Name    string
	Members []string
	Typ     types.Type
}

type SwitchInfo struct {
	Enum        bool
	Exhaustive  bool
}

type Info struct {
	Types    map[ast.Expr]types.Type
	IsType   map[ast.Expr]bool
	Idents   map[*ast.Ident]*Object
	Sels     map[*ast.SelectorExpr]*SelInfo
	Calls    map[*ast.CallExpr]*CallInfo
	Inits    map[*ast.VarDecl]*InitInfo
	Specs    map[*ast.VarSpec]*Object
	Consts   map[*ast.ConstDecl]*Object
	Switches map[*ast.SwitchStmt]*SwitchInfo
	Imports  []*ImportInfo
	Enums    []*EnumInfo
	Pkg      *types.Package
	Fset     *gotok.FileSet
}

func newInfo() *Info {
	return &Info{
		Types:    map[ast.Expr]types.Type{},
		IsType:   map[ast.Expr]bool{},
		Idents:   map[*ast.Ident]*Object{},
		Sels:     map[*ast.SelectorExpr]*SelInfo{},
		Calls:    map[*ast.CallExpr]*CallInfo{},
		Inits:    map[*ast.VarDecl]*InitInfo{},
		Specs:    map[*ast.VarSpec]*Object{},
		Consts:   map[*ast.ConstDecl]*Object{},
		Switches: map[*ast.SwitchStmt]*SwitchInfo{},
	}
}

func (i *Info) ImportByPkg(pkg *types.Package) *ImportInfo {
	for _, im := range i.Imports {
		if im.Pkg == pkg {
			return im
		}
	}
	return nil
}

type pos = token.Position
