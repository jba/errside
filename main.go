package main

import (
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
			fmt.Printf("%s: %v", dir, err)
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
		for filename, file := range pkg.Files {
			if err := processFile(filename, file, fset); err != nil {
				return err
			}
		}
	}
	return nil
}

func processFile(filename string, file *ast.File, fset *token.FileSet) error {
	fmt.Printf("== file %s ==\n", filename)
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	_, err := conf.Check("floop", fset, []*ast.File{file}, info)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BlockStmt:
			if n != nil {
				processBlockStmt(n, fset, info)
			}
			return false

		default:
			return true
		}
	})
	return nil
}

func processBlockStmt(bs *ast.BlockStmt, fset *token.FileSet, info *types.Info) {
	errVars := map[types.Object]bool{}
	for _, stmt := range bs.List {
		switch stmt := stmt.(type) {
		case *ast.AssignStmt:
			if len(stmt.Rhs) > 1 {
				fmt.Println("can't handle multiple rhs")
				return
			}
			rhs := stmt.Rhs[0]
			ty := info.Types[rhs].Type
			lt := lastType(ty)
			if isErrorType(lt) {
				if id, ok := stmt.Lhs[len(stmt.Lhs)-1].(*ast.Ident); ok {
					def := info.Defs[id]
					use := info.Uses[id]
					if def == nil {
					} else {
						errVars[def] = true
					}
					if use == nil {
					} else {
						errVars[use] = true
					}

				}
			}

		case *ast.IfStmt:
			if onError(stmt.Cond, info.Uses, errVars) == True {
				fmt.Printf(">>> stmt at %s\n", fset.Position(stmt.Pos()))
			}
		}
	}
}

func lastType(ty types.Type) types.Type {
	if tu, ok := ty.(*types.Tuple); ok {
		return tu.At(tu.Len() - 1).Type()
	}
	return ty
}

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

func onError(expr ast.Expr, defs map[*ast.Ident]types.Object, errVars map[types.Object]bool) tribool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		switch e.Op {
		case token.EQL:
			return not(errEqualsNil(e.X, e.Y, defs, errVars))
		case token.NEQ:
			return errEqualsNil(e.X, e.Y, defs, errVars)
		default:
			return Unknown
		}
	case *ast.ParenExpr:
		return onError(e.X, defs, errVars)
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			return not(onError(e.X, defs, errVars))
		}
		return Unknown

	default:
		return Unknown
	}
}

func errEqualsNil(e1, e2 ast.Expr, defs map[*ast.Ident]types.Object, errVars map[types.Object]bool) tribool {
	id1, ok1 := e1.(*ast.Ident)
	id2, ok2 := e2.(*ast.Ident)
	if !ok1 || !ok2 {
		return Unknown
	}
	obj1 := defs[id1]
	obj2 := defs[id2]
	if errVars[obj1] && id2.Name == "nil" {
		return True
	}
	if errVars[obj2] && id1.Name == "nil" {
		return True
	}
	return False
}
