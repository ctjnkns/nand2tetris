package symboltable

import (
	"fmt"
)

type SymbolTable struct {
	entries map[string]Entry
	index   map[Kind]int
}

type Kind int

const (
	NONE Kind = iota
	STATIC
	FIELD
	ARG
	VAR
)

type Entry struct {
	EntryType string
	Kind      Kind
	Index     int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		entries: map[string]Entry{},
		index:   map[Kind]int{},
	}
}

func (s *SymbolTable) Reset() {
	s.entries = map[string]Entry{}
	s.index = map[Kind]int{}
}

func (s *SymbolTable) Define(name, entryType string, kind Kind) error {
	switch kind {
	case STATIC, FIELD, ARG, VAR:
	default:
		return fmt.Errorf("unknown kind provided: %d", kind)
	}

	s.entries[name] = Entry{
		EntryType: entryType,
		Kind:      kind,
		Index:     s.index[kind],
	}
	s.index[kind]++

	return nil
}

func (s *SymbolTable) VarCount(kind Kind) int {
	return s.index[kind]
}

func (s *SymbolTable) KindOf(name string) Kind {
	if entry, ok := s.entries[name]; ok {
		return entry.Kind
	}

	return NONE
}

func (s *SymbolTable) TypeOf(name string) string {
	return s.entries[name].EntryType
}

func (s *SymbolTable) IndexOf(name string) int {
	return s.entries[name].Index
}
