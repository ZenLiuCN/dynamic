package sample

import "github.com/ZenLiuCN/dynamic"

type proto struct {
	name string
}

func (p proto) Name() string {
	return p.name
}

func (p proto) Action() string {
	return p.name
}

func NewFactory(name string) dynamic.Proto {
	return proto{name: name}
}
