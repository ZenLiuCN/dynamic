package dynamic

import (
	"errors"
	"github.com/ZenLiuCN/fn"
	"github.com/pkujhd/goloader"
	"io"
	"os"
	"unsafe"
)

type (
	/*Dynamic module from object files

	Use Steps:

	1. InitializeMany or Initialize or InitializeSerialized to initialize this dynamic module
	2. Use this module
	3. Call [Dynamic.Free] to release the resources

	Note:

	1. Must fetch and use one symbol as desired type inside one specific goroutine.
	2. Dynamic itself can be used safe between goroutines, but not thread-safe.
	*/
	Dynamic interface {
		InitializeMany(file, pkg []string, types ...any) (err error) //Initialize from many object files
		Initialize(file, pkg string, types ...any) (err error)       //Initialize from one object file
		InitializeSerialized(in io.Reader, types ...any) (err error) //Initialize from serialized linker
		Symbols() []string                                           //resolved symbols
		MissingSymbols() []string                                    //dump the missing symbols
		Serialize(out io.Writer) error                               //write linker data to an output binary format [gob] which may loaded by InitializeSerialized
		Fetch(sym string) (u unsafe.Pointer, ok bool)                //fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized
		MustFetch(sym string) (u unsafe.Pointer)                     // fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized or ErrMissingSymbol
		Free()                                                       //resources of dynamic module
	}
	dynamic struct {
		files  []string
		pkg    []string
		sym    map[string]uintptr
		linker *goloader.Linker
		module *goloader.CodeModule
	}
)

var (
	ErrMissingSymbol = errors.New("missing symbol")
	ErrUninitialized = errors.New("module not initialized")
)

func New() Dynamic {
	return new(dynamic)
}

// Inspect display symbols inside an object file
func Inspect(file, pkg string) ([]string, error) {
	return goloader.Parse(file, pkg)
}

func (s *dynamic) InitializeMany(file, pkg []string, types ...any) (err error) {
	if s.module != nil {
		s.Free()
	}
	s.sym = make(map[string]uintptr)
	if len(types) > 0 {
		goloader.RegTypes(s.sym, types...)
	}
	s.files = append(s.files, file...)
	s.pkg = append(s.pkg, pkg...)
	if err = goloader.RegSymbol(s.sym); err != nil {
		return
	}
	if s.linker, err = goloader.ReadObjs(file, pkg); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.sym); err != nil {
		return
	}
	return
}
func (s *dynamic) Initialize(file, pkg string, types ...any) (err error) {
	if s.module != nil {
		s.Free()
	}
	s.sym = make(map[string]uintptr)
	s.files = append(s.files, file)
	s.pkg = append(s.pkg, pkg)
	if len(types) > 0 {
		goloader.RegTypes(s.sym, types...)
	}
	if err = goloader.RegSymbol(s.sym); err != nil {
		return
	}
	if s.linker, err = goloader.ReadObj(file, pkg); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.sym); err != nil {
		return
	}
	return
}
func (s *dynamic) InitializeSerialized(in io.Reader, types ...any) (err error) {
	if s.module != nil {
		s.Free()
	}
	s.sym = make(map[string]uintptr)
	if len(types) > 0 {
		goloader.RegTypes(s.sym, types...)
	}
	if err = goloader.RegSymbol(s.sym); err != nil {
		return
	}
	if s.linker, err = goloader.UnSerialize(in); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.sym); err != nil {
		return
	}
	return
}
func (s *dynamic) require(sym string) (u uintptr, o bool) {
	u, o = s.module.Syms[sym]
	return
}
func (s *dynamic) Fetch(sym string) (u unsafe.Pointer, ok bool) {
	if s.module == nil {
		ok = false
		return
	}
	var p uintptr
	p, ok = s.require(sym)
	if !ok {
		return
	}
	funcPtrContainer := (uintptr)(unsafe.Pointer(&p))
	return unsafe.Pointer(&funcPtrContainer), ok
}
func (s *dynamic) MustFetch(sym string) (u unsafe.Pointer) {
	if s.module == nil {
		panic(ErrUninitialized)
	}
	p, ok := s.require(sym)
	if !ok {
		panic(ErrMissingSymbol)
	}
	funcPtrContainer := uintptr(unsafe.Pointer(&p))
	return unsafe.Pointer(&funcPtrContainer)
}
func (s *dynamic) Symbols() []string {
	return fn.MapKeys(s.sym)
}
func (s *dynamic) MissingSymbols() []string {
	if s.module == nil {
		panic(ErrUninitialized)
	}
	return goloader.UnresolvedSymbols(s.linker, s.sym)
}
func (s *dynamic) Serialize(out io.Writer) error {
	if s.module == nil {
		panic(ErrUninitialized)
	}
	return goloader.Serialize(s.linker, out)
}

func (s *dynamic) Free() {
	if s.module != nil {
		_ = os.Stdout.Sync()
		s.module.Unload()
		s.module = nil
		s.sym = nil
		s.linker = nil
		{
			n := len(s.pkg)
			if n > 0 {
				if n < 10 {
					s.pkg = s.pkg[:0]
				} else {
					s.pkg = nil
				}
			}
		}
		{
			n := len(s.files)
			if n > 0 {
				if n < 10 {
					s.files = s.files[:0]
				} else {
					s.files = nil
				}
			}
		}
	}
}
