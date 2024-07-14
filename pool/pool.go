package pool

import (
	"errors"
	. "github.com/ZenLiuCN/dynamic"
	"github.com/ZenLiuCN/fn"
	"github.com/pkujhd/goloader"
	"io"
	"slices"
	"sync"
)

type Pool struct {
	Symbols
	Modules map[string]Dynamic
	Loaded  []Dynamic
	sync.RWMutex
}

var (
	ErrAlreadyLoad    = errors.New("module already loaded")
	ErrNotLoad        = errors.New("module not loaded")
	ErrMissingPackage = errors.New("package not loaded")
	ErrCorrupted      = errors.New("recording corrupted")
)

func (p *Pool) RegisterSo(path string) error {
	return goloader.RegSymbolWithSo(p.Symbols, path)
}
func (p *Pool) RegisterExecute(path string) error {
	return goloader.RegSymbolWithPath(p.Symbols, path)
}
func (p *Pool) RegisterTypes(t ...any) {
	goloader.RegTypes(p.Symbols, t...)
}

// LoadFile load from go archive or go object file
func (p *Pool) LoadFile(file, pkgPath string) (err error) {
	p.Lock()
	defer p.Unlock()
	if pkgPath == "" {
		pkgPath = "main"
	}
	if _, ok := p.Modules[pkgPath]; ok {
		return ErrAlreadyLoad
	}
	d := NewDynamic(p.Symbols)
	if err = d.Initialize(file, pkgPath); err != nil {
		return
	}
	p.Modules[pkgPath] = d
	p.Loaded = append(p.Loaded, d)
	if err = d.Link(); err != nil {
		return
	}
	p.register(d)
	return
}
func (p *Pool) register(d Dynamic) {
	for s, u := range d.GetModule().Syms {
		if _, ok := p.Symbols[s]; ok {
			continue
		} else {
			p.Symbols[s] = u
		}
	}
}
func (p *Pool) unregister(d Dynamic) {
	for s, u := range d.GetModule().Syms {
		if x, ok := p.Symbols[s]; ok && x == u {
			delete(p.Symbols, s)
		}
	}
}

// LoadLinkable load from serialized link
func (p *Pool) LoadLinkable(bin io.Reader) (err error) {
	p.Lock()
	defer p.Unlock()
	d := NewDynamic(p.Symbols)
	if err = d.InitializeSerialized(bin); err != nil {
		return
	}
	l := d.GetLinker()
	for _, pkg := range l.Packages {
		if _, ok := p.Modules[pkg.PkgPath]; ok {
			return ErrAlreadyLoad
		}
		p.Modules[pkg.PkgPath] = d
	}
	p.Loaded = append(p.Loaded, d)
	if err = d.Link(); err != nil {
		return
	}
	p.register(d)
	return
}

// ReloadFile from go archive or go object file
func (p *Pool) ReloadFile(file, pkgPath string) (err error) {
	p.Lock()
	defer p.Unlock()
	if pkgPath == "" {
		pkgPath = "main"
	}
	if m, ok := p.Modules[pkgPath]; !ok {
		return ErrNotLoad
	} else {
		i := slices.Index(p.Loaded, m)
		if i < 0 {
			return ErrCorrupted
		}
		x := p.Loaded[i:]
		for i := len(x); i >= 0; i-- {
			dyn := x[i]
			delete(p.Modules, fn.MapKeyOf(p.Modules, dyn))
			p.unregister(dyn)
			dyn.Free(false)
		}
		p.Loaded = p.Loaded[:i]
	}
	d := NewDynamic(p.Symbols)
	if err = d.Initialize(file, pkgPath); err != nil {
		return
	}
	p.Modules[pkgPath] = d
	p.Loaded = append(p.Loaded, d)
	if err = d.Link(); err != nil {
		return
	}
	p.register(d)
	return
}

// ReloadLinkable from serialized link
func (p *Pool) ReloadLinkable(bin io.Reader) (err error) {
	p.Lock()
	defer p.Unlock()
	d := NewDynamic(p.Symbols)
	if err = d.InitializeSerialized(bin); err != nil {
		return
	}
	l := d.GetLinker()
	var dyn []Dynamic
	for _, pkg := range l.Packages {
		if dx, ok := p.Modules[pkg.PkgPath]; ok {
			dyn = append(dyn, dx)
		}
	}
	if len(dyn) > 0 {
		var i = 0xFFFFFFFF
		for _, dy := range dyn {
			x := slices.Index(p.Loaded, dy)
			if x < 0 {
				return ErrCorrupted
			}
			if x < i {
				i = x
			}
		}
		x := p.Loaded[i:]
		for i := len(x); i >= 0; i-- {
			dyn := x[i]
			delete(p.Modules, fn.MapKeyOf(p.Modules, dyn))
			p.unregister(dyn)
			dyn.Free(false)
		}
		p.Loaded = p.Loaded[:i]
	}
	for _, pkg := range l.Packages {
		p.Modules[pkg.PkgPath] = d
	}
	p.Loaded = append(p.Loaded, d)
	if err = d.Link(); err != nil {
		return
	}
	p.register(d)
	return
}

// Require fetch symbol from package
func (p *Pool) Require(pkgPath, symbolName string) Sym {
	p.RLock()
	defer p.RUnlock()
	if pkgPath == "" {
		pkgPath = "main"
	}
	if m, ok := p.Modules[pkgPath]; ok {
		return m.MustFetch(pkgPath + "." + symbolName)
	}
	panic(ErrMissingPackage)
}

// NewPool create new pool
func NewPool() (p *Pool, err error) {
	p = new(Pool)
	p.Modules = make(map[string]Dynamic)
	p.Symbols, err = NewSymbols()
	return
}
