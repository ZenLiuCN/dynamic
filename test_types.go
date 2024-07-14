package dynamic

type Proto interface {
	Name() string
	Action() string
}

var (
	TestDebug = true
)
