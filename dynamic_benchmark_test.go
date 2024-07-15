package dynamic

import (
	"github.com/ZenLiuCN/fn"
	"testing"
	"time"
)

func BenchmarkLoad(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(sym)
		fn.Panic(dyn.Initialize(moduleFunc, pkgSample))
		fn.Panic(dyn.Link())
	}
}
func BenchmarkLoadAndExecute(b *testing.B) {
	ready()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dyn := NewDynamic(sym)
		fn.Panic(dyn.Initialize(moduleFunc, pkgSample))
		act := AsOnce[typeFunc](m.MustFetch(symRun))
		act()
	}
}

func BenchmarkExecuteOnly(b *testing.B) {
	ready()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AsOnce[typeFunc](m.MustFetch(symRun))()
	}
}
func BenchmarkHoldingExecuteOnly(b *testing.B) {
	ready()
	f := m.MustFetch(symRun)
	fx := As[typeFunc](&f)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fx()
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
