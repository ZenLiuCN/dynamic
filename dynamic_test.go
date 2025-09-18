package dynamic

import (
	"os"
	"testing"
	"time"

	"github.com/ZenLiuCN/fn"

	"bytes"
	"encoding/hex"
	"sync"
)

const (
	moduleFunc    = "testdata/func.o"
	moduleConst   = "testdata/constant.o"
	moduleFactory = "testdata/factory.o"
	moduleArchive = "testdata/constant.a"
	symRun        = "sample.Run"
	symConst      = "sample.Const"
	symFactory    = "sample.NewFactory"
	pkgSample     = "sample"
)

type (
	typeFunc    = func() string
	typeFactory = func(name string) Proto
	typeConst   = func() Proto
)

var debugging = false
var sym = fn.Panic1(NewSymbols())

func TestArchive(t *testing.T) {
	d := NewDynamic(sym, debugging)
	for _, i := range sym.ExistsSymbols() {
		println(i)
	}
	var pt Proto
	fn.Panic(d.Initialize(moduleArchive, pkgSample, &pt))
	fn.Panic(d.Link())
	t.Log(d.Exports())
	fx := AsOnce[typeFunc](d.MustFetch(symRun))
	t.Log(fx())

	f := AsOnce[typeConst](d.MustFetch(symConst))()
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())
	f = AsOnce[typeFactory](d.MustFetch(symFactory))("archive")
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())

}
func TestConstant(t *testing.T) {
	d := NewDynamic(sym, debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleConst, pkgSample, &pt))
	fn.Panic(d.Link())
	t.Log(d.Exports())
	//t.Log(d.InitTask(pkgSample))
	f := AsOnce[typeConst](d.MustFetch(symConst))()
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())

}
func TestFactory(t *testing.T) {
	d := NewDynamic(sym, debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleFactory, pkgSample, &pt))
	fn.Panic(d.Link())
	f := AsOnce[typeFactory](d.MustFetch(symFactory))
	i := f("some")
	t.Logf("%+v", i)
	t.Log(i.Name())
	t.Log(i.Action())

}
func TestFactoryRoutines(t *testing.T) {
	d := NewDynamic(sym, debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleFactory, pkgSample, &pt))
	fn.Panic(d.Link())
	f := AsOnce[typeFactory](d.MustFetch(symFactory))
	i := f("some")
	t.Logf("%+v", i)
	t.Log(i.Name())
	t.Log(i.Action())
	var w sync.WaitGroup
	for n := 0; n < 10; n++ {
		w.Add(1)
		go func() {
			defer w.Done()
			t.Log(i.Name())
			t.Log(i.Action())
		}()
	}
	w.Wait()

}
func TestSerialize(t *testing.T) {
	ready()
	b := new(bytes.Buffer)
	fn.Panic(m.Serialize(b))
	m2 := NewDynamic(sym)
	println(hex.Dump(b.Bytes()))
	fn.Panic(m2.InitializeSerialized(b))
	fn.Panic(m2.Link())
	act := AsOnce[typeFunc](m2.MustFetch(symRun))
	println(act())

}
func TestLoad(t *testing.T) {
	ready()
	n, ok := m.Fetch(symRun)
	if !ok {
		println(m.ExistsSymbols())
		panic("not found sym")
	}
	println("one", As[typeFunc](&n)())
	println("two", AsOnce[typeFunc](m.MustFetch(symRun))())
	println("three", AsOnce[typeFunc](m.MustFetch(symRun))())

}
func TestRoutines(t *testing.T) {
	ready()
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			println(AsOnce[typeFunc](m.MustFetch(symRun))())
		}()
	}
	w.Wait()
}

var m *Dynamic

func ready() {
	if m == nil {
		m = NewDynamic(sym)
		fn.Panic(m.Initialize(moduleFunc, pkgSample, time.Now))
		fn.Panic(m.Link())
	}
}

func TestUse(t *testing.T) {
	dyn := NewDynamic(sym, debugging)
	fn.Panic(dyn.Initialize(moduleFunc, pkgSample, time.Now))
	fn.Panic(dyn.Link())
	defer dyn.Free(true)
	type args struct {
		dyn *Dynamic
		sym string
	}
	type testCase[T any] struct {
		name string
		args args
	}
	tests := []testCase[typeFunc]{
		{"some", args{dyn: dyn, sym: symRun}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Use[typeFunc](tt.args.dyn, tt.args.sym)
			got(func(f typeFunc, err error) {
				if err != nil {
					t.Errorf("Use() = %p, error %v", got, err)
				} else {
					t.Log(f())
				}
			})
		})
	}
}

func TestAs(t *testing.T) {
	dyn := NewDynamic(sym, debugging)
	fn.Panic(dyn.Initialize(moduleFunc, pkgSample))
	fn.Panic(dyn.Link())
	defer dyn.Free(true)
	f, ok := dyn.Fetch(symRun)
	if !ok {
		t.Errorf("not found ")
	} else {
		fx := As[typeFunc](&f)
		for i := 0; i < 19; i++ {
			t.Logf("%p => %s", &f, fx())
		}
	}
}

func TestLinkable(t *testing.T) {
	dyn := NewDynamic(sym, debugging)
	f := fn.Panic1(os.Open("testdata/constant.linkable"))
	fn.Panic(dyn.InitializeSerialized(f))
	fn.Panic(f.Close())
	l := dyn.GetLinker()
	for _, pkg := range l.Packages {
		t.Log(pkg.File, pkg.PkgPath)
	}
}
