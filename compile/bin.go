package main

import (
	"fmt"
	. "github.com/ZenLiuCN/dynamic"
	"github.com/pkujhd/goloader"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
)

func main() {
	app := cli.NewApp()
	app.Usage = "dynamic module compiler"
	app.Action = action
	app.Name = "Compile"
	app.Description = "dynamic module compiler which Compile go sources into objfile"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}},
	}
	app.Args = true
	app.Commands = []*cli.Command{
		{Name: "prepare", Action: prepare, Usage: "copy internals of go sdk"},
		{Name: "clean", Action: clean, Usage: "remove copied internals of go sdk"},
		{Name: "imports",
			Action: imports,
			Usage:  "display imports of objfile",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "pkg", Aliases: []string{"p"}, Usage: "package path or default main"},
			},
			Args: true,
		},
		{Name: "linker",
			Action: linkers,
			Usage:  "display imports of linker file",
			Args:   true,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failure %s", err)
	}
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
func action(ctx *cli.Context) error {
	d := ctx.Bool("debug")
	o := ctx.Args().Slice()
	if len(o) == 0 {
		return fmt.Errorf("missing target sources list")
	}
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("missing go sdk: %w ", err)
	}
	if err = Imports(d, o); err != nil {
		return err
	}
	return Compile(d, o)
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
