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
	t.Log(dynamic.As[func() string](p.Require("sample", "Run"))())
	sp := spew.NewDefaultConfig()
	sp.MaxDepth = 5
	sp.Dump(dynamic.As[func(string) dynamic.Proto](p.Require("sample", "NewFactory"))("123"))
	sp.Dump(dynamic.As[func() dynamic.Proto](p.Require("sample", "Const"))())
	sp.Dump(p)

}
