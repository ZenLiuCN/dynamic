package dynamic

import (
	"fmt"
	"github.com/pkujhd/goloader"
	"io"
	"log"
	"os"
	"strings"
	"unsafe"
)

type (
	//Sym is a simple alias of uintptr.
	Sym uintptr
	//Dynamic module from object files or serialized Liner file, this interface can not be implement outside this package.
	//
	//Use Steps:
	//
	//	1. InitializeMany or Initialize or InitializeSerialized to initialize this dynamic module.
	//	2. [Dynamic.Link] to link the code to runtime and other global dependencies.
	//	3. Use this module.
	//	3. Call [Dynamic.Free] to release the resources.
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
		Link() (err error)                                           //link and create code module
		MissingSymbols() []string                                    //dump the missing symbols
		Serialize(out io.Writer) error                               //write linker data to an output binary format [gob] which may loaded by InitializeSerialized
		Fetch(sym string) (u Sym, ok bool)                           //fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized
		MustFetch(sym string) (u Sym)                                // fetch a symbol as unsafe.Pointer, which can cast to the desired type, throws ErrUninitialized or ErrMissingSymbol
		Free(sync bool)                                              //resources of dynamic module, sync parameter to sync the stdout or not
		GetLinker() *goloader.Linker                                 //fetch the internal [goloader.Linker], it is nil before initial stage by [Dynamic.Initialize]
		GetModule() *goloader.CodeModule                             //fetch the internal [goloader.CodeModule], it is nil before link stage by [Dynamic.Link]
		internal()
	}
	// Symbols contains global resolved symbols
	//
	// If two Dynamic shares the same Symbols instance, it may depend on each other after make.
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
		debug  bool
	}
)

// NewDynamic create new dynamic with provided Symbols, an optional debug parameter will enable debug logging inside Dynamic
func NewDynamic(sym Symbols, debug ...bool) (d Dynamic) {
	x := new(dynamic)
	x.symbols = sym.(symbols)
	x.debug = debug != nil && len(debug) > 0 && debug[0]
	return x
}
func (s *dynamic) internal() {}
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
		if s.debug {
			log.Println("register types", types)
		}
		goloader.RegTypes(s.symbols, types...)
	}
	s.files = append(s.files, file...)
	s.pkg = append(s.pkg, pkg...)
	if s.linker, err = goloader.ReadObjs(file, pkg); err != nil {
		return
	}
	if s.debug {
		log.Printf("create linker: %+v", s.linker)
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
		if s.debug {
			log.Println("register types", types)
		}
		goloader.RegTypes(s.symbols, types...)
	}
	if s.linker, err = goloader.ReadObj(file, pkg); err != nil {
		return
	}
	if s.debug {
		log.Printf("create linker: %+v", s.linker)
	}
	return
}
func (s *dynamic) InitializeSerialized(in io.Reader, types ...any) (err error) {
	if s.linker != nil {
		return ErrAlreadyInitialized
	}
	if len(types) > 0 {
		if s.debug {
			log.Println("register types", types)
		}
		goloader.RegTypes(s.symbols, types...)
	}
	if s.linker, err = goloader.UnSerialize(in); err != nil {
		return
	}
	if s.debug {
		log.Printf("loaded linker: %+v", s.linker)
	}
	return
}

func (s *dynamic) Link() (err error) {
	if s.linker == nil {
		return ErrUninitialized
	}
	if s.module != nil {
		return ErrLinked
	}
	if s.module, err = goloader.Load(s.linker, s.symbols); err != nil {
		return
	}
	if s.debug {
		log.Printf("create module: %+v", s.module)
	}
	return
}

func (s *dynamic) Fetch(sym string) (u Sym, ok bool) {
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
	if s.debug {
		log.Printf("found symbol: %x", p)
	}
	return (Sym)(unsafe.Pointer(&p)), ok
}

func checkPackage(sym string) string {
	if strings.IndexByte(sym, '.') < 0 {
		return "main." + sym
	}
	return sym
}
func (s *dynamic) MustFetch(sym string) (u Sym) {
	if s.module == nil {
		panic(ErrUninitialized)
	}
	sym = checkPackage(sym)
	p, ok := s.module.Syms[sym]
	if !ok {
		panic(ErrMissingSymbol)
	}
	if s.debug {
		log.Printf("found symbol: %x", p)
	}
	return (Sym)(unsafe.Pointer(&p))
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
		if s.debug {
			log.Printf("free Dynamic: %+v", s)
		}
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

// Use create a function to fetch and use symbol on the fly
func Use[T any](dyn Dynamic, sym string) func(func(t T, err error)) {
	dbg := dyn.(*dynamic).debug
	return func(f func(t T, err error)) {
		var x T
		defer func() {
			switch y := recover().(type) {
			case nil:
				if dbg {
					log.Printf("execute %v", x)
				}
				f(x, nil)
			case error:
				if dbg {
					log.Printf("execute error %v, %v", x, y)
				}
				f(x, y)
			default:
				if dbg {
					log.Printf("execute other error %v, %v", x, y)
				}
				f(x, fmt.Errorf("%v", y))
			}
		}()
		p := dyn.MustFetch(sym)
		x = As[T](p)
	}
}

// As convert fetched Sym to contract type
func As[T any](ptr Sym) (x T) {
	px := (*T)(unsafe.Pointer(&ptr))
	x = *px
	return
}
