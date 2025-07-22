package waluku

import "strings"

type SymbolPair []string

func (s SymbolPair) String() string {
	return strings.Join(s, "")
}

func (s SymbolPair) Quoted() string {
	return s[1]
}
