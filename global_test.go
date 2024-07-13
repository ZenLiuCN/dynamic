package dynamic

import (
	"bytes"
	"encoding/hex"
	"github.com/ZenLiuCN/fn"
	"slices"
	"sync"
	"testing"
	"time"
)

const moduleFunc = "testdata/func.o"
const moduleFactory = "testdata/factory.o"
const symRun = "sample.Run"
const symFactory = "sample.NewFactory"

type typeFunc = func() string
type typeFactory = func(name string) Proto

var debugging = false

func TestFactory(t *testing.T) {
	d := NewDynamic(NewSymbols(), debugging)
	var pt Proto
	fn.Panic(d.Initialize(moduleFactory, "sample", &pt))
	fn.Panic(d.Link())
	f := As[typeFactory](d.MustFetch(symFactory))
	i := f("some")
	t.Logf("%+v", i)
	t.Log(i.Name())
	t.Log(i.Action())

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
	symbol := m2.Symbols()
	exported := fn.MapKeys(m2.GetModule().Syms)
	for _, s := range exported {
		if slices.Index(symbol, s) < 0 {
			t.Logf("'%s' not set in sym", s)
		}
	}
}
func TestLoad(t *testing.T) {
	ready()
	n, ok := m.Fetch(symRun)
	if !ok {
		println(m.Symbols())
		panic("not found sym")
	}
	act := As[typeFunc](n)
	println(act())
	act = As[typeFunc](m.MustFetch(symRun))
	println(act())
}
func TestRoutines(t *testing.T) {
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
		fn.Panic(dyn.Initialize(moduleFunc, "sample"))
		fn.Panic(dyn.Link())
	}
}
func BenchmarkLoadAndExecute(b *testing.B) {
	ready()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(NewSymbols())
		fn.Panic(dyn.Initialize(moduleFunc, "sample"))
		act := As[typeFunc](m.MustFetch(symRun))
		act()
	}
}

var m Dynamic

func ready() {
	if m == nil {
		m = NewDynamic(NewSymbols())
		fn.Panic(m.Initialize(moduleFunc, "sample", time.Now))
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
