package main

import (
	"fmt"
	. "github.com/ZenLiuCN/dynamic"
	"github.com/pkujhd/goloader"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Usage = "dynamic module compiler"
	app.Name = "Compiler"
	app.Description = "dynamic module compiler which Compile go sources into objfile"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
		},
	}
	app.Args = true
	app.Commands = []*cli.Command{
		{
			Name:   "compile",
			Action: compile,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "pack", Aliases: []string{"a"}, Usage: "pack dependency packages"},
				&cli.StringSliceFlag{Name: "includes", Aliases: []string{"c"}, Usage: "pack dependency packages only included"},
				&cli.StringFlag{Name: "pkg", Aliases: []string{"k"}, Usage: "package import path, required with -a or --pack"},
			},
			Args:  true,
			Usage: "compile go source to objfile or go archive. the arguments can be list of go sources or '.' for lookup at working directory.",
		},
		{
			Name:   "prepare",
			Action: prepare,
			Usage:  "copy internals of go sdk",
		},
		{
			Name:   "clean",
			Action: clean,
			Usage:  "remove copied internals of go sdk",
		},
		{
			Name:   "imports",
			Action: imports,
			Usage:  "display imports of go objfile or go archive file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "pkg",
					Aliases: []string{"p"},
					Usage:   "package path or default main",
				},
			},
			Args: true,
		},
		{
			Name:   "linkable",
			Action: linkers,
			Usage:  "display imports of linkable file",
			Args:   true,
		},
		{
			Name:   "module",
			Action: module,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "includes",
					Aliases: []string{"c"},
					Usage:   "pack dependencies packages only included",
				},
				&cli.StringFlag{
					Name:    "pkg",
					Aliases: []string{"k"},
					Usage:   "package import path, required with -a or --pack",
				},
			},
			Usage: "compile current work directory as go module to linkable",
			Args:  false,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failure %s", err)
	}
}

func module(ctx *cli.Context) (err error) {
	d := ctx.Bool("debug")
	var o []string
	o, err = lookup()
	if err != nil {
		return
	}
	_, err = exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("missing go sdk: %w ", err)
	}
	if err = Imports(d, o); err != nil {
		return fmt.Errorf("generate importcfg : %w ", err)
	}
	i := ctx.StringSlice("c")
	pk := ctx.String("k")
	if pk == "" {
		return fmt.Errorf("required argument -k|--pkgPath missing")
	}
	return Packs(d, o, pk, i)
}

func linkers(ctx *cli.Context) (err error) {
	var f *os.File
	var l *goloader.Linker
	for _, s := range ctx.Args().Slice() {
		if f, err = os.OpenFile(s, os.O_WRONLY, os.ModePerm); err != nil {
			return
		}
		l, err = goloader.UnSerialize(f)
		if err != nil {
			return
		}
		v := LinkerImportsIter(l)
		log.Printf("\n%s", v.String())
	}
	return
}

func imports(ctx *cli.Context) (err error) {
	for _, s := range ctx.Args().Slice() {
		var v *Info
		err, v = ObjectImportsIter(s, ctx.String("pkg"))
		if err != nil {
			return
		}
		log.Printf("\n%s", v.String())
	}
	return
}
func compile(ctx *cli.Context) (err error) {
	d := ctx.Bool("debug")
	o := ctx.Args().Slice()
	if len(o) == 0 {
		return fmt.Errorf("missing target sources list")
	}
	if len(o) == 1 && o[0] == "." {
		if d {
			log.Printf("will use all .go files as sources")
		}

		o, err = lookup()
		if err != nil {
			return
		}
		log.Printf("found go sources at working directory: %v", o)
	}
	_, err = exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("missing go sdk: %w ", err)
	}
	if err = Imports(d, o); err != nil {
		return fmt.Errorf("generate importcfg : %w ", err)
	}
	if ctx.Bool("a") {
		i := ctx.StringSlice("c")
		pk := ctx.String("k")
		if pk == "" {
			return fmt.Errorf("required argument -k|--pkgPath missing")
		}
		return Packs(d, o, pk, i)
	}
	return Compile(d, o, true)
}

func lookup() (v []string, err error) {
	var wd string
	wd, err = os.Getwd()
	if err != nil {
		return
	}
	var e []os.DirEntry
	e, err = os.ReadDir(wd)
	if err != nil {
		return
	}
	for _, entry := range e {
		if entry.IsDir() {
			continue
		}
		n := entry.Name()
		if strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			v = append(v, n)
		}
	}
	return
}

func clean(ctx *cli.Context) (err error) {
	d := ctx.Bool("debug")
	dir := os.ExpandEnv("$GOROOT/src/cmd/objfile")
	if d {
		log.Printf("clean go sdk: %s", dir)
	}
	if _, err = os.Stat(dir); err == nil {
		err = os.RemoveAll(dir)
		if d {
			log.Printf("removed %s", dir)
		}
	} else if d {
		log.Printf("did nothing for %s", dir)
	}
	return
}

func prepare(ctx *cli.Context) (err error) {
	d := ctx.Bool("debug")
	src := os.ExpandEnv("$GOROOT/src/cmd/internal")
	dir := os.ExpandEnv("$GOROOT/src/cmd/objfile")
	if d {
		log.Printf("prepare go sdk from %s to %s", src, dir)
	}
	if _, err = os.Stat(dir); err != nil && os.IsNotExist(err) {
		err = CopyDir(src, dir, nil)
		if d {
			log.Printf("copied %s from %s", dir, src)
		}
	} else if d {
		log.Printf("did nothing for %s", dir)
	}
	return
}
