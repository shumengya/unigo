package names

import (
	"unicode"
	"unicode/utf8"
)

func Export(name string) string {
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size <= 1 {
		return "X" + name
	}
	if r >= 'a' && r <= 'z' {
		return string(r-'a'+'A') + name[size:]
	}
	if unicode.Is(unicode.Lu, r) {
		return name
	}
	if unicode.Is(unicode.Ll, r) {
		return string(unicode.ToUpper(r)) + name[size:]
	}
	return "X" + name
}
