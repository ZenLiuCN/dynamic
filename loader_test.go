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
	b := new(bytes.Buffer)
	fn.Panic(m.Initialize("testdata/main.plugin", "sample"))
	fn.Panic(m.Serialize(b))
	m2 := new(dynamic)
	println(hex.Dump(b.Bytes()))
	fn.Panic(m2.InitializeSerialized(b))
	act := *(*fnc)(m2.MustFetch("sample.Run"))
	println(act())
}
func TestLoad(t *testing.T) {
	m := new(dynamic)
	fn.Panic(m.Initialize("testdata/main.plugin", "sample"))
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
		m := new(dynamic)
		fn.Panic(m.Initialize("testdata/main.plugin", "sample"))
	}
}
func BenchmarkLoadAndExecute(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := new(dynamic)
		fn.Panic(m.Initialize("testdata/main.plugin", "sample"))
		n := m.MustFetch("sample.Run")
		act := *(*fnc)(n)
		act()
	}
}

var m *dynamic

func init() {
	m = new(dynamic)
	fn.Panic(m.Initialize("sample/main.plugin", "sample"))
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
