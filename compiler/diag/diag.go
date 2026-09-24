package diag

import (
	"fmt"
	"strings"

	"unigo/compiler/token"
)

type Error struct {
	Pos token.Position
	Msg string
}

func (e Error) Error() string {
	if e.Pos.File == "" && e.Pos.Line == 0 {
		return e.Msg
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.File, e.Pos.Line, e.Pos.Col, e.Msg)
}

type List []Error

func (l List) Error() string {
	var b strings.Builder
	for i, e := range l {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(e.Error())
	}
	return b.String()
}

func (l *List) Add(pos token.Position, msg string) {
	*l = append(*l, Error{Pos: pos, Msg: msg})
}

func (l List) Err() error {
	if len(l) == 0 {
		return nil
	}
	return l
}
