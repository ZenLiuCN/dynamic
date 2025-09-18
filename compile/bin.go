package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	. "github.com/ZenLiuCN/dynamic"
	"github.com/pkujhd/goloader"
	"github.com/urfave/cli/v3"
)

func main() {

	if err := (&cli.Command{
		Name:        "builder",
		Description: "dynamic module compiler which Compile go sources into objfile",
		Usage:       "dynamic module compiler",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "compile",
				Action: compile,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "pack", Aliases: []string{"a"}, Usage: "pack dependency packages"},
					&cli.BoolFlag{Name: "noPkg", Aliases: []string{"n"}, Usage: "pack without dependencies"},
					&cli.StringSliceFlag{Name: "includes", Aliases: []string{"c"}, Usage: "pack dependency packages only included, if provided excludes will no effect."},
					&cli.StringSliceFlag{Name: "excludes", Aliases: []string{"e"}, Usage: "pack dependencies packages excluded"},
					&cli.StringFlag{Name: "pkg", Aliases: []string{"k"}, Usage: "package import path, required with -a or --pack"},
				},
				Usage: "compile go source to objfile or go archive. the arguments can be list of go sources or '.' for lookup at working directory.",
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "files",
						Min:  1,
						Max:  -1,
					},
				},
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
					&cli.StringFlag{Name: "pkg", Aliases: []string{"p"}, Usage: "package path or default main"},
				},
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "files",
						Min:  1,
						Max:  -1,
					},
				},
			},
			{
				Name:   "linkable",
				Action: linkers,
				Usage:  "display imports of linkable file",
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "files",
						Min:  1,
						Max:  -1,
					},
				},
			},
			{
				Name:   "module",
				Action: module,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "noPkg", Aliases: []string{"n"}, Usage: "pack without dependencies"},
					&cli.StringSliceFlag{Name: "includes", Aliases: []string{"c"}, Usage: "pack dependencies packages only included, if provided, excludes will no effect."},
					&cli.StringSliceFlag{Name: "excludes", Aliases: []string{"e"}, Usage: "pack dependencies packages excluded"},
					&cli.StringFlag{Name: "pkg", Aliases: []string{"k"}, Usage: "package import path, required with -a or --pack"},
				},
				Usage: "compile current work directory as go module to linkable",
			},
		},
	}).Run(context.Background(), os.Args); err != nil {
		log.Fatalf("failure %s", err)
	}
}

func module(ctx context.Context, cmd *cli.Command) (err error) {
	d := cmd.Bool("debug")
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
	i := cmd.StringSlice("c")
	e := cmd.StringSlice("e")
	pk := cmd.String("k")
	if pk == "" {
		return fmt.Errorf("required argument -k|--pkgPath missing")
	}
	return Packs(d, o, pk, cmd.Bool("n"), i, e)
}

func linkers(ctx context.Context, cmd *cli.Command) (err error) {
	var f *os.File
	var l *goloader.Linker
	for _, s := range cmd.StringArgs("files") {
		if f, err = os.OpenFile(s, os.O_RDONLY, os.ModePerm); err != nil {
			return
		}
		l, err = goloader.UnSerialize(f)
		if err != nil {
			return
		}
		v := LinkerImportsIter(l)
		miss := goloader.UnresolvedSymbols(l, nil)
		s := new(strings.Builder)
		for _, s2 := range miss {
			if !strings.HasPrefix(s2, "runtime") {
				s.WriteString("\t" + s2 + "\n")
			}
		}
		log.Printf("\npackags:\n%s\nmissing\n%s", v.String(), s.String())
	}
	return
}

func imports(ctx context.Context, cmd *cli.Command) (err error) {
	for _, s := range cmd.StringArgs("files") {
		var v *Info
		err, v = ObjectImportsIter(s, cmd.String("pkg"))
		if err != nil {
			return
		}
		log.Printf("\n%s", v.String())
	}
	return
}
func compile(ctx context.Context, cmd *cli.Command) (err error) {
	d := cmd.Bool("debug")
	o := cmd.StringArgs("files")
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
	if cmd.Bool("a") {
		i := cmd.StringSlice("c")
		e := cmd.StringSlice("e")
		pk := cmd.String("k")
		if pk == "" {
			return fmt.Errorf("required argument -k|--pkgPath missing")
		}
		return Packs(d, o, pk, cmd.Bool("n"), i, e)
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

func clean(ctx context.Context, cmd *cli.Command) (err error) {
	d := cmd.Bool("debug")
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

func prepare(ctx context.Context, cmd *cli.Command) (err error) {
	d := cmd.Bool("debug")
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
