package dynamic

import (
	"github.com/ZenLiuCN/fn"
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestDependencies(t *testing.T) {
	b := NewDynamic(fn.Panic1(NewSymbols()))
	fn.Panic(b.Initialize("testdata/constant.a", "sample"))
	t.Log(spew.Sprint(b.MissingSymbols()))
	t.Log(spew.Sdump(b.ExistsSymbols()))
}
