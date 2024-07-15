package dynamic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/ZenLiuCN/fn"
	"github.com/pkujhd/goloader"
	"github.com/pkujhd/goloader/obj"
	"slices"

	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CopyFile from src to dest with optional src file info
func CopyFile(src string, dest string, si fs.FileInfo) (err error) {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fn.IgnoreClose(sf)
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fn.IgnoreClose(df)
	_, err = io.Copy(df, sf)
	if err == nil {
		if si == nil {
			si, err = os.Stat(src)
			if err != nil {
				return
			}
		}
		err = os.Chmod(dest, si.Mode())
	}
	return
}

// CopyDir from src to dest with optional src file info
func CopyDir(src string, dest string, si fs.FileInfo) (err error) {
	if si == nil {
		si, err = os.Stat(src)
		if err != nil {
			return err
		}
	}
	err = os.MkdirAll(dest, si.Mode())
	if err != nil {
		return err
	}
	var sp string
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		sp, err = filepath.Rel(src, filepath.Dir(path))
		if err != nil {
			return err
		}
		dp := filepath.Join(dest, sp, info.Name())
		if info.IsDir() {
			err = CopyDir(path, dp, info)
		} else {
			err = CopyFile(path, dp, info)
		}
		return err
	})

}

// Compile an object file output to working directory
func Compile(debug bool, o []string, remove bool) (err error) {
	var cmd *exec.Cmd
	if len(o) == 1 {
		cmd = exec.Command("go", "tool", "compile", "-importcfg", "importcfg", o[0])
	} else {
		cmd = exec.Command("go", append([]string{"tool", "compile", "-importcfg", "importcfg", "-pack"}, o...)...)
	}
	if debug {
		log.Printf("execute: %v", cmd.Args)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil && remove {
		err = os.Remove("importcfg")
	}
	return
}

var (
	pkgPrefix    = []byte("packagefile ")
	vendorPrefix = []byte("vendor/")
	skipPackages = []string{
		"runtime",
		"internal",
		"github.com/ZenLiuCN/dynamic",
		"github.com/ZenLiuCN/fn",
		"github.com/pkujhd/goloader",
		"go",
		"syscall",
		"reflect",
		"os",
		"archive",
		"bufio",
		"bytes",
		"cmp",
		"compress",
		"container",
		"context",
		"crypto",
		"database",
		"debug",
		"embed",
		"encoding",
		"errors",
		"expvar",
		"flag",
		"fmt",
		"hash",
		"html",
		"image",
		"index",
		"io",
		"log",
		"maps",
		"math",
		"mime",
		"net",
		"path",
		"plugin",
		"regexp",
		"slices",
		"sort",
		"strconv",
		"strings",
		"sync",
		"testing",
		"text",
		"time",
		"unicode",
	}
)

func checkAll(p string) bool {
	for _, skipPackage := range skipPackages {
		if check(p, skipPackage) {
			return true
		}
	}
	return false
}
func check(p string, pkg string) bool {
	return strings.HasPrefix(p, pkg+"/") || p == pkg
}
func Packs(dbg bool, sources []string, pkgPath string, includes []string, excludes []string) (err error) {
	defer func() {
		if err == nil && !dbg {
			err = os.Remove("importcfg")
		}
	}()
	err = Compile(dbg, sources, false)
	if err != nil {
		return
	}
	ic := len(includes) != 0
	ec := len(excludes) != 0
	b := fn.Panic1(os.ReadFile("importcfg"))
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Split(bufio.ScanLines)
	d := make(map[string]string)
	for s.Scan() {
		t := s.Bytes()
		t = bytes.TrimPrefix(t, pkgPrefix)
		i := bytes.IndexByte(t, '=')
		pkg := t[:i]
		if len(pkg) > len(vendorPrefix) && bytes.Equal(pkg[:7], vendorPrefix) {
			continue
		}
		p := string(pkg)
		if checkAll(p) {
			continue
		}
		if ic && slices.IndexFunc(includes, func(s string) bool {
			return s == p || strings.HasPrefix(p, s)
		}) >= 0 {
			d[p] = string(t[i+1:])
		} else if !ic && ec && slices.IndexFunc(excludes, func(s string) bool {
			return s == p || strings.HasPrefix(p, s)
		}) < 0 {
			d[p] = string(t[i+1:])
		}
	}
	px := strings.TrimSuffix(sources[0], ".go")
	n := px
	if len(sources) > 1 {
		n += ".a"
	} else {
		n += ".o"
	}
	var ss Symbols
	ss, err = NewSymbols()
	if err != nil {
		return
	}
	l := NewDynamic(ss)
	err = l.Initialize(n, pkgPath)
	if err != nil {
		return
	}
	if dbg {
		for _, s2 := range l.MissingSymbols() {
			log.Printf("required %s", s2)
		}
	}
	for pkg, file := range d {
		if dbg {
			log.Printf("will pack with %s from %s", pkg, file)
		}
		err = l.LoadDependencies(Dependency{
			File:    file,
			PkgPath: pkg,
			Symbols: nil,
		})
		if err != nil {
			return errors.Join(fmt.Errorf("loading dependecy %s from %s", pkg, file), err)
		}
	}
	var o *os.File
	o, err = os.OpenFile(px+".linkable", os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	defer fn.IgnoreClose(o)
	err = l.Serialize(o)
	return
}

// Imports generate import cfg as importcfg file in current working directory.
func Imports(debug bool, f []string) (err error) {
	if debug {
		log.Printf("sources: %v", f)
	}
	var cfg *os.File
	if cfg, err = os.OpenFile("importcfg", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm); err != nil {
		return
	}
	defer fn.IgnoreClose(cfg)
	cmd := exec.Command("go", append([]string{"list", "-export", "-f", "{{.Imports}}"}, f...)...)
	if debug {
		log.Printf("execute: %v", cmd.Args)
	}
	var out string
	var bout []byte
	if bout, err = cmd.Output(); err != nil {
		return fmt.Errorf("inpsect Imports: %w\nerr:%s\nout:%s", err, err.(*exec.ExitError).Stderr, string(bout))
	}
	out = strings.TrimSpace(string(bout))
	if out != "" && out[0] == '[' {
		out = out[1 : len(out)-1]
	}
	in := strings.Split(out, " ")
	if debug {
		fmt.Println("deps", in)
	}
	cmd = exec.Command("go", append([]string{"list", "-export", "-f", "{{if .Export}}packagefile {{.ImportPath}}={{.Export}}{{end}}", "std"}, in...)...)
	if debug {
		log.Printf("execute: %v", cmd.Args)
	}
	if bout, err = cmd.Output(); err != nil {
		return fmt.Errorf("inpsect dependecies: %w\nerr:%s\nout:%s", err, err.(*exec.ExitError).Stderr, string(bout))
	}
	if debug {
		fmt.Println("importcfg", string(bout))
	}
	_, err = cfg.Write(bout)
	return
}

// ObjectImportsIter resolve all imported packages and version (only if it's a module).
//
// this use for parse dependencies
func ObjectImportsIter(file, pkgPath string) (err error, info *Info) {
	v := &obj.Pkg{Syms: make(map[string]*obj.ObjSymbol, 0), File: file, PkgPath: pkgPath}
	if v.PkgPath == obj.EmptyString {
		v.PkgPath = "main"
	}
	if err = v.Symbols(); err != nil {
		return
	}
	info = parseInfo(v)
	info.File = file
	info.PkgPath = pkgPath
	return
}

// LinkerImportsIter resolve all imported packages and version if it's a module.
//
// this use for parse dependencies
func LinkerImportsIter(link *goloader.Linker) (infos Infos) {
	for _, pkg := range link.Packages {
		info := parseInfo(pkg)
		info.File = pkg.File
		info.PkgPath = pkg.PkgPath
		infos = append(infos, info)
	}
	return
}

// Infos is a stringer slice of Info
type Infos []*Info

func (i Infos) String() string {
	s := strings.Builder{}
	for _, v := range i {
		s.WriteString(v.String())
	}
	return s.String()
}

// Info contains the import information of a linker
type Info struct {
	File    string
	PkgPath string
	Imports map[string]string // with pairs of package import path and version
}

func (i Info) String() string {
	s := strings.Builder{}
	for p, v := range i.Imports {
		if v != "" {
			s.WriteString(fmt.Sprintf("\t%s@%s\n", p, v))
		} else {
			s.WriteString(fmt.Sprintf("\t%s\n", p))
		}
	}
	return s.String()
}

func parseInfo(v *obj.Pkg) (i *Info) {
	i = new(Info)
	i.Imports = make(map[string]string)
	for _, pkg := range v.ImportPkgs {
		i.Imports[pkg] = ""
	}
	k := fn.MapKeys(i.Imports)
	for _, f := range v.CUFiles {
		f = strings.TrimPrefix(f, "gofile..")
		if strings.HasPrefix(f, "$GOROOT") {
			continue
		}
		if strings.IndexByte(f, '!') >= 0 {
			f = parseName(f)
		}
		for _, s := range k {
			x := strings.Index(f, s)
			if x >= 0 {
				f = f[x:]
				if i.Imports[s] == "" {
					y := strings.IndexByte(f, '@')
					ver := f[y+1:]
					y = strings.IndexByte(ver, '/')
					ver = ver[:y]
					i.Imports[s] = ver
				}
			}
		}
	}
	return
}

func parseName(f string) string {
	v := strings.Builder{}
	x := false
	for _, i := range []byte(f) {
		switch {
		case i == '!':
			x = true
		case x:
			x = false
			v.WriteByte(i - 32)
		default:
			v.WriteByte(i)
		}
	}
	return v.String()
}

// Inspect display symbols inside an object file
func Inspect(file, pkg string) ([]string, error) {
	return goloader.Parse(file, pkg)
}
