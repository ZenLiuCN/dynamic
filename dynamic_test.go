package dynamic

import (
	"github.com/ZenLiuCN/fn"
	"testing"
	"time"
)

func TestUse(t *testing.T) {
	dyn := NewDynamic(NewSymbols(), debugging)
	fn.Panic(dyn.Initialize(moduleFunc, "sample", time.Now))
	fn.Panic(dyn.Link())
	defer dyn.Free(true)
	type args struct {
		dyn Dynamic
		sym string
	}
	type testCase[T any] struct {
		name string
		args args
	}
	tests := []testCase[typeFunc]{
		{"some", args{dyn: dyn, sym: symRun}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Use[typeFunc](tt.args.dyn, tt.args.sym)
			got(func(f typeFunc, err error) {
				if err != nil {
					t.Errorf("Use() = %p, error %v", got, err)
				} else {
					t.Log(f())
				}
			})
		})
	}
}

func TestUseAs(t *testing.T) {
	dyn := NewDynamic(NewSymbols(), debugging)
	fn.Panic(dyn.Initialize(moduleFunc, "sample", time.Now))
	fn.Panic(dyn.Link())
	defer dyn.Free(true)
	f, ok := dyn.Fetch(symRun)
	if !ok {
		t.Errorf("not found ")
	} else {
		t.Log(As[typeFunc](f)())
	}
}
