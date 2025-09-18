package dynamic

import (
	"errors"

	"github.com/pkujhd/goloader"
)

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

func NewSymbols() (t Symbols, err error) {
	t = make(map[string]uintptr)
	err = goloader.RegSymbol(t)
	return
}
