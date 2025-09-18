package pool

import (
	"github.com/ZenLiuCN/dynamic"
	"github.com/ZenLiuCN/fn"

	"testing"
)

func TestNewPool(t *testing.T) {
	p := fn.Panic1(NewPool())
	d := dynamic.Proto(nil)
	p.RegisterTypes(&d)
	fn.Panic(p.LoadFile("../testdata/constant.a", "sample"))
	s0 := p.Require("sample", "Run")
	t.Log(dynamic.As[func() string](&s0)())

	s1 := p.Require("sample", "NewFactory")
	t.Logf("%#+v", s1)
}
