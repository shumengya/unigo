package token

type Kind int

const (
	EOF Kind = iota
	Illegal

	Ident
	String
	Int
	Float

	Package
	Import
	From
	Var
	Const
	Func
	If
	Else
	Elif
	While
	For
	Switch
	Case
	Default
	Enum
	Struct
	Interface
	Newtype
	Alias
	Return
	Map
	True
	False
	Nil

	Fallthrough
	Iota
	Class
	New
	Extends
	Implements
	TypeWord

	Add
	Sub
	Mul
	Quo
	Rem
	Assign
	Eql
	Neq
	Lss
	Gtr
	Leq
	Geq
	Land
	Lor
	Not
	And

	Lparen
	Rparen
	Lbrace
	Rbrace
	Lbrack
	Rbrack
	Comma
	Semi
	Colon
	Period
)

func (k Kind) String() string {
	if int(k) < 0 || int(k) >= len(kindNames) {
		return "token"
	}
	return kindNames[k]
}

var kindNames = []string{
	EOF:          "文件结尾",
	Illegal:      "非法记号",
	Ident:        "名字",
	String:       "字符串",
	Int:          "整数",
	Float:        "小数",
	Package:      "package",
	Import:       "import",
	From:         "from",
	Var:          "var",
	Const:        "const",
	Func:         "func",
	If:           "if",
	Else:         "else",
	Elif:         "elif",
	While:        "while",
	For:          "for",
	Switch:       "switch",
	Case:         "case",
	Default:      "default",
	Enum:         "enum",
	Struct:       "struct",
	Interface:    "interface",
	Newtype:      "newtype",
	Alias:        "alias",
	Return:       "return",
	Map:          "map",
	True:         "true",
	False:        "false",
	Nil:          "nil",
	Fallthrough:  "fallthrough",
	Iota:         "iota",
	Class:        "class",
	New:          "new",
	Extends:      "extends",
	Implements:   "implements",
	TypeWord:     "type",
	Add:          "+",
	Sub:          "-",
	Mul:          "*",
	Quo:          "/",
	Rem:          "%",
	Assign:       "=",
	Eql:          "==",
	Neq:          "!=",
	Lss:          "<",
	Gtr:          ">",
	Leq:          "<=",
	Geq:          ">=",
	Land:         "&&",
	Lor:          "||",
	Not:          "!",
	And:          "&",
	Lparen:       "(",
	Rparen:       ")",
	Lbrace:       "{",
	Rbrace:       "}",
	Lbrack:       "[",
	Rbrack:       "]",
	Comma:        ",",
	Semi:         ";",
	Colon:        ":",
	Period:       ".",
}

func IsForbidden(k Kind) bool {
	switch k {
	case Fallthrough, Iota, Class, New, Extends, Implements, TypeWord:
		return true
	default:
		return false
	}
}

type Position struct {
	File   string
	Offset int
	Line   int
	Col    int
}

type Token struct {
	Kind        Kind
	Lit         string
	Pos         Position
	SpaceBefore int
}
