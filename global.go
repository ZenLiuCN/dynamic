package dynamic

import (
	"errors"
	"github.com/ZenLiuCN/fn"
	"github.com/pkujhd/goloader"
	"os"
)

var (
	gob     map[string]uintptr
	modules map[string]*dynamic
)

func init() {
	gob = make(map[string]uintptr)
	fn.Panic(goloader.RegSymbol(gob))
}

var (
	// ErrAlreadyExists occurs when register an already registered global Dynamic
	ErrAlreadyExists = errors.New("already registered into global")
	// ErrNotExists occurs when unregister a not registered global Dynamic
	ErrNotExists = errors.New("not registered into global")
)

// UseGlobalSo register symbols form a golang dynamic library (.so)
func UseGlobalSo(p string) error {
	return goloader.RegSymbolWithSo(gob, p)
}

// UseGlobalExecute register symbols form a go executable, those symbols can't be linked,
// this function should only be use for testing dependencies.
func UseGlobalExecute(p string) error {
	return goloader.RegSymbolWithPath(gob, p)
}

// UseGlobalTypes register types as global dependencies.
func UseGlobalTypes(p ...any) {
	goloader.RegTypes(gob, p...)
}

// UseGlobalObject load a relocatable object file  and register to global dependencies.
func UseGlobalObject(file string, pkg string) (err error) {
	if _, ok := modules[file]; ok {
		return ErrAlreadyExists
	}
	n := new(dynamic)
	n.symbols = gob
	err = n.Initialize(file, pkg)
	if err != nil {
		return err
	}
	err = n.Link()
	if err != nil {
		return err
	}
	modules[file] = n
	register(n)
	return
}

// UseGlobalLinker load a serialized linker and register to global dependencies.
func UseGlobalLinker(file string) (err error) {
	if _, ok := modules[file]; ok {
		return ErrAlreadyExists
	}
	n := new(dynamic)
	n.symbols = gob
	var f *os.File
	f, err = os.OpenFile(file, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	err = n.InitializeSerialized(f)
	if err != nil {
		return err
	}
	err = n.Link()
	if err != nil {
		return err
	}
	modules[file] = n
	register(n)
	return
}

// GlobalDynamics returns a copy map of global shared dynamics. should not modify any data. the result is map[FilePath|ModuleName] Dynamic
func GlobalDynamics() (v map[string]Dynamic) {
	v = make(map[string]Dynamic, len(modules))
	for s, d := range modules {
		v[s] = d
	}
	return
}

// GlobalSymbols returns a copy of global symbols
func GlobalSymbols() (v map[string]uintptr) {
	v = make(map[string]uintptr, len(gob))
	for s, d := range gob {
		v[s] = d
	}
	return
}

// CloseGlobalDynamics close all global dynamics and reload runtime symbols. this should only use when all Dynamics are free!
func CloseGlobalDynamics() error {
	for k, d := range modules {
		unregister(d)
		d.Free(true)
		delete(modules, k)
	}
	return goloader.RegSymbol(gob)
}

// RegisterGlobalDynamic register an user Dynamic into global dependencies. name must be unique
func RegisterGlobalDynamic(name string, d Dynamic) error {
	if _, ok := modules[name]; ok {
		return ErrAlreadyExists
	}
	modules[name] = d.(*dynamic)
	register(d.(*dynamic))
	return nil
}

// UnregisterGlobalDynamic unregister an user Dynamic by register name from global dependencies.
func UnregisterGlobalDynamic(name string) error {
	if d, ok := modules[name]; !ok {
		return ErrNotExists
	} else {
		unregister(d)
		delete(modules, name)
	}
	return nil
}

func register(d *dynamic) {
	if d.module == nil {
		return
	}
	for s, u := range d.module.Syms {
		if _, ok := gob[s]; !ok {
			gob[s] = u
		}
	}
}
func unregister(d *dynamic) {
	if d.module == nil {
		return
	}
	for s, u := range d.module.Syms {
		if x, ok := gob[s]; ok && x == u {
			delete(gob, s)
		}
	}
}
