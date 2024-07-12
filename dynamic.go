package dynamic

import (
	"github.com/pkujhd/goloader"
	"io"
	"os"
	"strings"
	"unsafe"
)

type (
	//Dynamic module from object files or serialized Liner file
	//
	//Use Steps:
	//
	//	1. InitializeMany or Initialize or InitializeSerialized to initialize this dynamic module
	//	2. [Dynamic.Make] to link the code to runtime
	//	3. Use this module
	//	3. Call [Dynamic.Free] to release the resources
	//
	//Note:
	//
	//	1. Must fetch and use one symbol as desired type inside one specific goroutine.
	//	2. Dynamic itself can be used safe between goroutines, but not thread-safe.
	Dynamic interface {
		Symbols
		InitializeMany(file, pkg []string, types ...any) (err error) //Initialize from many object files
		Initialize(file, pkg string, types ...any) (err error)       //Initialize from one object file
		InitializeSerialized(in io.Reader, types ...any) (err error) //Initialize from serialized linker
		Make() (err error)                                           // make the code module
		MissingSymbols() []string                                    //dump the missing symbols
		Serialize(out io.Writer) error                               //write linker data to an output binary format [gob] which may loaded by InitializeSerialized
		Fetch(sym string) (u unsafe.Pointer, ok bool)                //fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized
		MustFetch(sym string) (u unsafe.Pointer)                     // fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized or ErrMissingSymbol
		Free(sync bool)                                              //resources of dynamic module, sync parameter to sync the stdout or not

		GetLinker() *goloader.Linker
		GetModule() *goloader.CodeModule
	}
	// Symbols contains global resolved symbol and with local module symbols
	//
	// If two Dynamic shares the same Symbols instance, it may depend on each other after linked.
	Symbols interface {
		Symbols() []string //resolved symbols
	}
	symbols map[string]uintptr
	dynamic struct {
		files []string
		pkg   []string
		symbols
		linker *goloader.Linker
		module *goloader.CodeModule
	}
)

func NewDynamic(sym Symbols) (d Dynamic) {
	x := new(dynamic)
	x.symbols = sym.(symbols)
	return x
}
func (s *dynamic) GetLinker() *goloader.Linker {
	return s.linker
}
func (s *dynamic) GetModule() *goloader.CodeModule {
	return s.module
}
func (s *dynamic) InitializeMany(file, pkg []string, types ...any) (err error) {
	if s.linker != nil {
		return ErrAlreadyInitialized
	}
	if len(types) > 0 {
		goloader.RegTypes(s.symbols, types...)
	}
	s.files = append(s.files, file...)
	s.pkg = append(s.pkg, pkg...)
	if s.linker, err = goloader.ReadObjs(file, pkg); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.symbols); err != nil {
		return
	}
	return
}
func (s *dynamic) Initialize(file, pkg string, types ...any) (err error) {
	if s.linker != nil {
		return ErrAlreadyInitialized
	}
	s.files = append(s.files, file)
	s.pkg = append(s.pkg, pkg)
	if len(types) > 0 {
		goloader.RegTypes(s.symbols, types...)
	}
	if s.linker, err = goloader.ReadObj(file, pkg); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.symbols); err != nil {
		return
	}
	return
}
func (s *dynamic) InitializeSerialized(in io.Reader, types ...any) (err error) {
	if s.linker != nil {
		return ErrAlreadyInitialized
	}
	if len(types) > 0 {
		goloader.RegTypes(s.symbols, types...)
	}
	if s.linker, err = goloader.UnSerialize(in); err != nil {
		return
	}
	if s.module, err = goloader.Load(s.linker, s.symbols); err != nil {
		return
	}
	return
}

func (s *dynamic) Make() (err error) {
	if s.linker != nil {
		return ErrUninitialized
	}
	if s.module != nil {
		return ErrLinked
	}
	if s.module, err = goloader.Load(s.linker, s.symbols); err != nil {
		return
	}
	return
}

func (s *dynamic) Fetch(sym string) (u unsafe.Pointer, ok bool) {
	if s.module == nil {
		ok = false
		return
	}
	sym = checkPackage(sym)
	var p uintptr
	p, ok = s.module.Syms[sym]
	if !ok {
		return
	}
	funcPtrContainer := (uintptr)(unsafe.Pointer(&p))
	return unsafe.Pointer(&funcPtrContainer), ok
}

func checkPackage(sym string) string {
	if strings.IndexByte(sym, '.') < 0 {
		return "main." + sym
	}
	return sym
}
func (s *dynamic) MustFetch(sym string) (u unsafe.Pointer) {
	if s.module == nil {
		panic(ErrUninitialized)
	}
	sym = checkPackage(sym)
	p, ok := s.module.Syms[sym]
	if !ok {
		panic(ErrMissingSymbol)
	}
	funcPtrContainer := uintptr(unsafe.Pointer(&p))
	return unsafe.Pointer(&funcPtrContainer)
}

func (s *dynamic) MissingSymbols() []string {
	if s.linker == nil {
		panic(ErrUninitialized)
	}
	return goloader.UnresolvedSymbols(s.linker, s.symbols)
}
func (s *dynamic) Serialize(out io.Writer) error {
	if s.linker == nil {
		panic(ErrUninitialized)
	}
	return goloader.Serialize(s.linker, out)
}

func (s *dynamic) Free(sync bool) {
	if s.linker != nil {
		if s.module != nil {
			if sync {
				_ = os.Stdout.Sync()
			}
			s.module.Unload()
			s.module = nil
		}
		s.symbols = nil
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
