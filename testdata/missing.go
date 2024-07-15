package sample

import "log"

// go:generate go install github.com/ZenLiuCN/dynamic/compile@latest
//
//go:generate compile missing.go
func Print(args ...any) {
	log.Println(args...)
}
