package dynamic

import (
	"errors"
	"github.com/ZenLiuCN/fn"
	"maps"
)

func NewSymbols() Symbols {
	return symbols(maps.Clone(gob))
}
func (s symbols) Symbols() []string {
	return fn.MapKeys(s)
}

var (
	ErrMissingSymbol      = errors.New("missing symbol")
	ErrAlreadyInitialized = errors.New("already initialized dynamic")
	ErrLinked             = errors.New("already linked")
	ErrUninitialized      = errors.New("module not initialized")
)
