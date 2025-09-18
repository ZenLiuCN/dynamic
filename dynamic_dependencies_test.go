package dynamic

import (
	"testing"

	"github.com/ZenLiuCN/fn"
)

func TestDependencies(t *testing.T) {
	b := NewDynamic(fn.Panic1(NewSymbols()))
	fn.Panic(b.Initialize("testdata/constant.a", "sample"))
	t.Log(b.MissingSymbols())
	t.Log(b.ExistsSymbols())
}
