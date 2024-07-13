package dynamic

import (
	"errors"
	"github.com/ZenLiuCN/fn"
	"maps"
)

// NewSymbols create a Symbols with global symbols
func NewSymbols() Symbols {
	return symbols(maps.Clone(gob))
}

// Symbols dump symbol names inside Symbol
func (s symbols) Symbols() []string {
	return fn.MapKeys(s)
}

var (
	// ErrMissingSymbol occurs when can't found a symbol.
	ErrMissingSymbol = errors.New("missing symbol")
	// ErrAlreadyInitialized occurs when a Dynamic reinitializing.
	ErrAlreadyInitialized = errors.New("already initialized dynamic")
	// ErrLinked occurs when a Dynamic relinking.
	ErrLinked = errors.New("already linked")
	// ErrUninitialized occurs use or link a Dynamic before initialized.
	ErrUninitialized = errors.New("module not initialized")
)
