package sample

import "github.com/ZenLiuCN/dynamic"

// go:generate go install github.com/ZenLiuCN/dynamic/compile@latest
//
//go:generate compile
type constProto struct {
	name string
}

func (p constProto) Name() string {
	return p.name
}

func (p constProto) Action() string {
	return p.name
}

var (
	Consted dynamic.Proto
)

func init() {
	Consted = constProto{name: "constant"}
}

func Const() dynamic.Proto {
	return Consted
}
