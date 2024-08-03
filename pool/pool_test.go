package pool

import (
	"github.com/ZenLiuCN/dynamic"
	"github.com/ZenLiuCN/fn"
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestNewPool(t *testing.T) {
	p := fn.Panic1(NewPool())
	d := dynamic.Proto(nil)
	p.RegisterTypes(&d)
	fn.Panic(p.LoadFile("../testdata/constant.a", "sample"))
	s0 := p.Require("sample", "Run")
	t.Log(dynamic.As[func() string](&s0)())
	sp := spew.NewDefaultConfig()
	sp.MaxDepth = 5
	s1 := p.Require("sample", "NewFactory")
	sp.Dump(dynamic.As[func(string) dynamic.Proto](&s1)("123"))
	s2 := p.Require("sample", "Const")
	sp.Dump(dynamic.As[func() dynamic.Proto](&s2)())
	for name, dyn := range p.Modules {
		sp.Dump(name, dyn)
		for s, symbol := range dyn.GetLinker().ObjSymbolMap {
			sp.Dump(s, symbol.Name)
		}
		for s, symbol := range dyn.GetModule().Syms {
			sp.Dump(s, symbol)
		}
	}

}
