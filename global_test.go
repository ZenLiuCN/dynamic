package dynamic

import (
	"bytes"
	"encoding/hex"
	"github.com/ZenLiuCN/fn"
	"sync"
	"testing"
	"time"
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

func TestArchive(t *testing.T) {
	d := NewDynamic(NewSymbols(), debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleArchive, pkgSample, &pt))
	fn.Panic(d.Link())
	t.Log(d.Exports())
	fx := As[typeFunc](d.MustFetch(symRun))
	t.Log(fx())

	f := As[typeConst](d.MustFetch(symConst))()
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())
	f = As[typeFactory](d.MustFetch(symFactory))("archive")
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())

}
func TestConstant(t *testing.T) {
	d := NewDynamic(NewSymbols(), debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleConst, pkgSample, &pt))
	fn.Panic(d.Link())
	t.Log(d.Exports())
	//t.Log(d.InitTask(pkgSample))
	f := As[typeConst](d.MustFetch(symConst))()
	t.Logf("%+v", f)
	t.Log(f.Name())
	t.Log(f.Action())

}
func TestFactory(t *testing.T) {
	d := NewDynamic(NewSymbols(), debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleFactory, pkgSample, &pt))
	fn.Panic(d.Link())
	f := As[typeFactory](d.MustFetch(symFactory))
	i := f("some")
	t.Logf("%+v", i)
	t.Log(i.Name())
	t.Log(i.Action())

}
func TestFactoryRoutines(t *testing.T) {
	d := NewDynamic(NewSymbols(), debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleFactory, pkgSample, &pt))
	fn.Panic(d.Link())
	f := As[typeFactory](d.MustFetch(symFactory))
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
	m2 := NewDynamic(NewSymbols())
	println(hex.Dump(b.Bytes()))
	fn.Panic(m2.InitializeSerialized(b))
	fn.Panic(m2.Link())
	act := As[typeFunc](m2.MustFetch(symRun))
	println(act())

}
func TestLoad(t *testing.T) {
	ready()
	n, ok := m.Fetch(symRun)
	if !ok {
		println(m.Symbols())
		panic("not found sym")
	}
	println("one", As[typeFunc](n)())
	println("two", As[typeFunc](m.MustFetch(symRun))())
	println("three", As[typeFunc](m.MustFetch(symRun))())

}
func TestRoutines(t *testing.T) {
	ready()
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			println(As[typeFunc](m.MustFetch(symRun))())
		}()
	}
	w.Wait()
}
func BenchmarkLoad(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(NewSymbols())
		fn.Panic(dyn.Initialize(moduleFunc, pkgSample))
		fn.Panic(dyn.Link())
	}
}
func BenchmarkLoadAndExecute(b *testing.B) {
	ready()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(NewSymbols())
		fn.Panic(dyn.Initialize(moduleFunc, pkgSample))
		act := As[typeFunc](m.MustFetch(symRun))
		act()
	}
}

var m Dynamic

func ready() {
	if m == nil {
		m = NewDynamic(NewSymbols())
		fn.Panic(m.Initialize(moduleFunc, pkgSample, time.Now))
		fn.Panic(m.Link())
	}
}
func BenchmarkExecuteOnly(b *testing.B) {
	ready()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		As[typeFunc](m.MustFetch(symRun))()
	}
}
func Run() string {
	return time.Now().String()
}
func BenchmarkExecuteRaw(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Run()
	}
}
