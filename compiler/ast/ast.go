package ast

import "unigo/compiler/token"

type Node interface {
	Pos() token.Position
}

type Decl interface {
	Node
	declNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type TypeExpr interface {
	Node
	typeNode()
}

type File struct {
	Package *PackageDecl
	Imports []*ImportDecl
	Decls   []Decl
}

type PackageDecl struct {
	At   token.Position
	Name string
}

func (n *PackageDecl) Pos() token.Position { return n.At }

type ImportDecl struct {
	At   token.Position
	Name string
	Path string
}

func (n *ImportDecl) Pos() token.Position { return n.At }

type ConstDecl struct {
	At    token.Position
	Name  string
	NamePos token.Position
	Type  TypeExpr
	Value Expr
}

func (n *ConstDecl) Pos() token.Position { return n.At }
func (*ConstDecl) declNode()             {}
func (*ConstDecl) stmtNode()             {}

type VarSpec struct {
	At      token.Position
	Name    string
	Type    TypeExpr
}

func (n *VarSpec) Pos() token.Position { return n.At }

type ElseReturn struct {
	At    token.Position
	Bare  bool
	Value Expr
}

func (n *ElseReturn) Pos() token.Position { return n.At }

type VarDecl struct {
	At    token.Position
	Paren bool
	Specs []*VarSpec
	Value Expr
	Else  *ElseReturn
}

func (n *VarDecl) Pos() token.Position { return n.At }
func (*VarDecl) declNode()             {}
func (*VarDecl) stmtNode()             {}

type FuncSig struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Params  []*VarSpec
	Result  TypeExpr
}

func (n *FuncSig) Pos() token.Position { return n.At }

type FuncDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Params  []*VarSpec
	Result  TypeExpr
	Body    *Block
}

func (n *FuncDecl) Pos() token.Position { return n.At }
func (*FuncDecl) declNode()             {}

type EnumMember struct {
	At   token.Position
	Name string
}

type EnumDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Members []EnumMember
}

func (n *EnumDecl) Pos() token.Position { return n.At }
func (*EnumDecl) declNode()             {}

type StructDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Fields  []*VarSpec
}

func (n *StructDecl) Pos() token.Position { return n.At }
func (*StructDecl) declNode()             {}

type InterfaceDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Methods []*FuncSig
}

func (n *InterfaceDecl) Pos() token.Position { return n.At }
func (*InterfaceDecl) declNode()             {}

type NewtypeDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Type    TypeExpr
}

func (n *NewtypeDecl) Pos() token.Position { return n.At }
func (*NewtypeDecl) declNode()             {}

type AliasDecl struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Type    TypeExpr
}

func (n *AliasDecl) Pos() token.Position { return n.At }
func (*AliasDecl) declNode()             {}

type Block struct {
	At   token.Position
	List []Stmt
}

func (n *Block) Pos() token.Position { return n.At }
func (*Block) stmtNode()             {}

type AssignStmt struct {
	At     token.Position
	Target Expr
	Value  Expr
}

func (n *AssignStmt) Pos() token.Position { return n.At }
func (*AssignStmt) stmtNode()             {}

type Elif struct {
	At   token.Position
	Cond Expr
	Body *Block
}

type IfStmt struct {
	At    token.Position
	Cond  Expr
	Body  *Block
	Elifs []Elif
	Else  *Block
}

func (n *IfStmt) Pos() token.Position { return n.At }
func (*IfStmt) stmtNode()             {}

type WhileStmt struct {
	At   token.Position
	Cond Expr
	Body *Block
}

func (n *WhileStmt) Pos() token.Position { return n.At }
func (*WhileStmt) stmtNode()             {}

type ForStmt struct {
	At   token.Position
	Cond Expr
	Body *Block
}

func (n *ForStmt) Pos() token.Position { return n.At }
func (*ForStmt) stmtNode()             {}

type CaseClause struct {
	At        token.Position
	IsDefault bool
	Value     Expr
	Body      *Block
}

func (n *CaseClause) Pos() token.Position { return n.At }

type SwitchStmt struct {
	At    token.Position
	Tag   Expr
	Cases []*CaseClause
}

func (n *SwitchStmt) Pos() token.Position { return n.At }
func (*SwitchStmt) stmtNode()             {}

type ReturnStmt struct {
	At    token.Position
	Value Expr
}

func (n *ReturnStmt) Pos() token.Position { return n.At }
func (*ReturnStmt) stmtNode()             {}

type ExprStmt struct {
	At token.Position
	X  Expr
}

func (n *ExprStmt) Pos() token.Position { return n.At }
func (*ExprStmt) stmtNode()             {}

type Ident struct {
	At   token.Position
	Name string
}

func (n *Ident) Pos() token.Position { return n.At }
func (*Ident) exprNode()             {}

type BasicLit struct {
	At    token.Position
	Kind  token.Kind
	Value string
}

func (n *BasicLit) Pos() token.Position { return n.At }
func (*BasicLit) exprNode()             {}

type BinaryExpr struct {
	At token.Position
	Op token.Kind
	X  Expr
	Y  Expr
}

func (n *BinaryExpr) Pos() token.Position { return n.At }
func (*BinaryExpr) exprNode()             {}

type UnaryExpr struct {
	At token.Position
	Op token.Kind
	X  Expr
}

func (n *UnaryExpr) Pos() token.Position { return n.At }
func (*UnaryExpr) exprNode()             {}

type CallExpr struct {
	At   token.Position
	Fun  Expr
	Args []Expr
}

func (n *CallExpr) Pos() token.Position { return n.At }
func (*CallExpr) exprNode()             {}

type SelectorExpr struct {
	At  token.Position
	X   Expr
	Sel string
}

func (n *SelectorExpr) Pos() token.Position { return n.At }
func (*SelectorExpr) exprNode()             {}

type IndexExpr struct {
	At    token.Position
	X     Expr
	Index Expr
}

func (n *IndexExpr) Pos() token.Position { return n.At }
func (*IndexExpr) exprNode()             {}

type SliceLit struct {
	At   token.Position
	Elts []Expr
}

func (n *SliceLit) Pos() token.Position { return n.At }
func (*SliceLit) exprNode()             {}

type MapEntry struct {
	Key   Expr
	Value Expr
}

type MapLit struct {
	At      token.Position
	Entries []MapEntry
}

func (n *MapLit) Pos() token.Position { return n.At }
func (*MapLit) exprNode()             {}

type FieldInit struct {
	At    token.Position
	Name  string
	Value Expr
}

type StructLit struct {
	At     token.Position
	Type   Expr
	Fields []FieldInit
}

func (n *StructLit) Pos() token.Position { return n.At }
func (*StructLit) exprNode()             {}

type FuncLit struct {
	At      token.Position
	Name    string
	NamePos token.Position
	Params  []*VarSpec
	Result  TypeExpr
	Body    *Block
}

func (n *FuncLit) Pos() token.Position { return n.At }
func (*FuncLit) exprNode()             {}

type InterfaceLit struct {
	At      token.Position
	Type    Expr
	Methods []*FuncLit
}

func (n *InterfaceLit) Pos() token.Position { return n.At }
func (*InterfaceLit) exprNode()             {}

type IdentType struct {
	At   token.Position
	Name string
}

func (n *IdentType) Pos() token.Position { return n.At }
func (*IdentType) typeNode()             {}

type SelectorType struct {
	At   token.Position
	Pkg  string
	Name string
}

func (n *SelectorType) Pos() token.Position { return n.At }
func (*SelectorType) typeNode()             {}

type PointerType struct {
	At   token.Position
	Base TypeExpr
}

func (n *PointerType) Pos() token.Position { return n.At }
func (*PointerType) typeNode()             {}

type SliceType struct {
	At   token.Position
	Elem TypeExpr
}

func (n *SliceType) Pos() token.Position { return n.At }
func (*SliceType) typeNode()             {}

type MapType struct {
	At    token.Position
	Key   TypeExpr
	Value TypeExpr
}

func (n *MapType) Pos() token.Position { return n.At }
func (*MapType) typeNode()             {}
