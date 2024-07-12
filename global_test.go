package dynamic

import (
	"bytes"
	"encoding/hex"
	"github.com/ZenLiuCN/fn"
	"sync"
	"testing"
	"time"
)

type fnc = func() string

func TestSerialize(t *testing.T) {
	ready()
	b := new(bytes.Buffer)
	fn.Panic(m.Serialize(b))
	m2 := NewDynamic(NewSymbols())
	println(hex.Dump(b.Bytes()))
	fn.Panic(m2.InitializeSerialized(b))
	act := *(*fnc)(m2.MustFetch("sample.Run"))
	println(act())
	for _, s := range m2.Symbols() {
		println(s)
	}
}
func TestLoad(t *testing.T) {
	ready()
	fn.Panic(m.Initialize("testdata/main.o", "sample"))
	n, ok := m.Fetch("sample.Run")
	if !ok {
		println(m.Symbols())
		panic("not found sym")
	}
	act := *(*fnc)(n)
	println(act())
	act = *(*fnc)(m.MustFetch("sample.Run"))
	println(act())
}
func TestRoutines(t *testing.T) {
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			println((*(*fnc)(m.MustFetch("sample.Run")))())
		}()
	}
	w.Wait()
}
func BenchmarkLoad(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(NewSymbols())
		fn.Panic(dyn.Initialize("testdata/main.o", "sample"))
	}
}
func BenchmarkLoadAndExecute(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(NewSymbols())
		fn.Panic(dyn.Initialize("testdata/main.o", "sample"))
		n := dyn.MustFetch("sample.Run")
		act := *(*fnc)(n)
		act()
	}
}

var m Dynamic

func ready() {
	if m == nil {
		m = NewDynamic(NewSymbols())
		fn.Panic(m.Initialize("testdata/main.o", "sample", time.Now))
	}
}
func BenchmarkExecuteOnly(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		(*(*fnc)(m.MustFetch("sample.Run")))()
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
