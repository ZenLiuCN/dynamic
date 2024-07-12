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
	ErrAlreadyExists = errors.New("already imported same file into global")
)

func UseGlobalSo(p string) error {
	return goloader.RegSymbolWithSo(gob, p)
}
func UseGlobalPath(p string) error {
	return goloader.RegSymbolWithPath(gob, p)
}
func UseGlobalTypes(p ...any) {
	goloader.RegTypes(gob, p...)
}
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
	err = n.Make()
	if err != nil {
		return err
	}
	modules[file] = n
	return
}
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
	err = n.Make()
	if err != nil {
		return err
	}
	modules[file] = n
	return
}

// GlobalDynamics the global shared dynamics. should not modify any data.
func GlobalDynamics() (v map[string]Dynamic) {
	v = make(map[string]Dynamic, len(modules))
	for s, d := range modules {
		v[s] = d
	}
	return
}

// CloseGlobalDynamics close all global dynamics and reload runtime symbols. this should only use when all dynamics are free!
func CloseGlobalDynamics() error {
	for k, d := range modules {
		d.Free(true)
		delete(modules, k)
	}
	for s := range gob {
		delete(gob, s)
	}
	return goloader.RegSymbol(gob)
}
