package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
)

func main() {
	flag.Parse()
	ok := true
	for _, dir := range flag.Args() {
		if err := processDir(dir); err != nil {
			fmt.Printf("%s: %v\n", dir, err)
			ok = false
		}
	}
	if !ok {
		os.Exit(1)
	}
}

func processDir(dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		var files []*ast.File
		for _, file := range pkg.Files {
			files = append(files, file)
		}
		info := &types.Info{
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
			Types: make(map[ast.Expr]types.TypeAndValue),
		}
		conf := types.Config{Importer: importer.Default()}
		_, err := conf.Check("floop", fset, files, info)
		if err != nil {
			return err
		}
		for filename, file := range pkg.Files {
			eis, err := processFile(filename, file, fset, info)
			if err != nil {
				return err
			}
			if len(eis) > 0 {
				if err := displayFile(filename, eis); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func processFile(filename string, file *ast.File, fset *token.FileSet, info *types.Info) ([]*ErrInfo, error) {
	fmt.Printf("== file %s ==\n", filename)
	var eis []*ErrInfo

	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BlockStmt:
			if n != nil {
				e := processBlockStmt(n, fset, info)
				eis = append(eis, e...)
			}
			return false

		default:
			return true
		}
	})
	return eis, nil
}

func processBlockStmt(bs *ast.BlockStmt, fset *token.FileSet, info *types.Info) []*ErrInfo {
	var eis []*ErrInfo
	lastSet := map[types.Object]*ast.Ident{}
	for _, stmt := range bs.List {
		switch stmt := stmt.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					def := info.Defs[id]
					use := info.Uses[id]
					if def != nil && use != nil && def != use {
						panic("confused")
					}
					if def != nil {
						lastSet[def] = id
					} else if use != nil {
						lastSet[use] = id
					}
				}
			}

		case *ast.IfStmt:
			if onError(stmt.Cond, info) == True {
				eis = append(eis, newErrInfo(fset, stmt))
			}
		}
	}
	return eis
}

type ErrInfo struct {
	filename   string
	start, end int
}

func (e *ErrInfo) includes(line int) bool {
	return e.start <= line && line <= e.end
}

func newErrInfo(fset *token.FileSet, n ast.Node) *ErrInfo {
	p1 := fset.Position(n.Pos())
	p2 := fset.Position(n.End())
	return &ErrInfo{
		filename: p1.Filename,
		start:    p1.Line,
		end:      p2.Line,
	}
}

// func lastType(ty types.Type) types.Type {
// 	if tu, ok := ty.(*types.Tuple); ok {
// 		return tu.At(tu.Len() - 1).Type()
// 	}
// 	return ty
// }

// isErrorType reports whether t is the built-in error type.
func isErrorType(t types.Type) bool {
	nt, ok := t.(*types.Named)
	if !ok {
		return false
	}
	tn := nt.Obj()
	return tn.Pkg() == nil && tn.Name() == "error"
}

type tribool int

const (
	Unknown tribool = iota
	False
	True
)

func not(t tribool) tribool {
	switch t {
	case False:
		return True
	case True:
		return False
	}
	return Unknown
}

func onError(expr ast.Expr, info *types.Info) tribool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		switch e.Op {
		case token.EQL:
			return not(errEqualsNil(e.X, e.Y, info))
		case token.NEQ:
			return errEqualsNil(e.X, e.Y, info)
		default:
			return Unknown
		}
	case *ast.ParenExpr:
		return onError(e.X, info)
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			return not(onError(e.X, info))
		}
		return Unknown

	default:
		return Unknown
	}
}

func errEqualsNil(e1, e2 ast.Expr, info *types.Info) tribool {
	t1 := info.TypeOf(e1)
	t2 := info.TypeOf(e2)
	if isErrorType(t1) && isNil(t2) {
		return True
	}
	if isErrorType(t2) && isNil(t1) {
		return True
	}
	return False
}

func isNil(t types.Type) bool {
	if b, ok := t.(*types.Basic); ok {
		if b.Kind() == types.UntypedNil {
			return true
		}
	}
	return false
}

func includesLine(eis []*ErrInfo, line int) bool {
	for _, e := range eis {
		if e.includes(line) {
			return true
		}
	}
	return false
}

func displayFile(filename string, eis []*ErrInfo) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	n := 0
	for s.Scan() {
		n++
		line := s.Text()
		if includesLine(eis, n) {
			fmt.Printf("\t\t\t")
		}
		fmt.Println(line)
	}
	return nil
}
