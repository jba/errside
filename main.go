package main

import (
	"flag"
	"fmt"
	"go/token"

	"os"

	"github.com/jba/errside/ast"
	"github.com/jba/errside/errstmt"
	"github.com/jba/errside/importer"
	"github.com/jba/errside/parser"
	"github.com/jba/errside/printer"
	"github.com/jba/errside/types"
)

var errcol = flag.Int("e", 40, "error column")

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
			err := processFile(filename, file, fset, info)
			if err != nil {
				return err
			}
			conf := &printer.Config{
				Mode:     printer.UseSpaces,
				Tabwidth: 4,
				Errcol:   *errcol,
			}
			if err := conf.Fprint(os.Stdout, fset, file); err != nil {
				return err
			}
		}
	}
	return nil
}

func processFile(filename string, file *ast.File, fset *token.FileSet, info *types.Info) error {
	fmt.Printf("== file %s ==\n", filename)

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
	var newList []ast.Stmt
	for i, stmt := range bs.List {
		newList = append(newList, stmt)
		ifStmt, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}
		// We have an if statement.
		// Does the if's test compare an identifier to nil?
		obj, tb := onError(ifStmt.Cond, info)
		if tb != True {
			continue
		}
		// Yes it does.
		// Was the previous statement an assignment?
		if i == 0 {
			continue
		}
		aStmt, ok := bs.List[i-1].(*ast.AssignStmt)
		if !ok {
			continue
		}
		// Yes it was.
		// Was the last expr on the lhs of the assignment the same identifier
		// tested in the if statement?
		obj2 := lastObj(aStmt.Lhs, info)
		if obj != obj2 {
			continue
		}
		// Yes it was. We have something like
		//    ..., err := ..
		//    if err != nil { ... }

		// The last two elements of newList are the assignment and if statements.
		// Replace both with a new "statement".
		n := len(newList)
		newList[n-2] = errstmt.NewAssignIfErrStmt(aStmt, ifStmt, obj)
		newList = newList[:n-1]
	}
	bs.List = newList
}

// lastObj returns the types.Object for the last expression in exprs, if
// it is an identifer. Otherwise it returns nil.
func lastObj(exprs []ast.Expr, info *types.Info) types.Object {
	if len(exprs) == 0 {
		return nil
	}
	id, ok := exprs[len(exprs)-1].(*ast.Ident)
	if !ok {
		return nil
	}
	return info.ObjectOf(id)
}

// onError reports whether expr is an inequality check between nil and
// an identifier of type error. It also returns the Object associated
// with the identifier.
// Examples:
//     err != nil
//     !(err == nil)
//	   nil != err
//     ((err != nil))
func onError(expr ast.Expr, info *types.Info) (types.Object, tribool) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		switch e.Op {
		case token.EQL:
			obj, t := errEqualsNil(e.X, e.Y, info)
			return obj, not(t)
		case token.NEQ:
			return errEqualsNil(e.X, e.Y, info)
		default:
			return nil, Unknown
		}
	case *ast.ParenExpr:
		return onError(e.X, info)
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			obj, t := onError(e.X, info)
			return obj, not(t)
		}
		return nil, Unknown

	default:
		return nil, Unknown
	}
}

// errEqualsNil reports whether the two exprs are an identifier of type error and
// nil. It returns the types.Object associated with the identifier.
func errEqualsNil(e1, e2 ast.Expr, info *types.Info) (types.Object, tribool) {
	t1 := info.TypeOf(e1)
	t2 := info.TypeOf(e2)
	var errExpr ast.Expr
	if isErrorType(t1) && isNil(t2) {
		errExpr = e1
	} else if isErrorType(t2) && isNil(t1) {
		errExpr = e2
	}
	if errExpr == nil {
		return nil, False
	}
	if id, ok := errExpr.(*ast.Ident); ok {
		return info.ObjectOf(id), True
	}
	return nil, False
}

// isNil reports whether type t the "untyped nil" type
func isNil(t types.Type) bool {
	if b, ok := t.(*types.Basic); ok {
		if b.Kind() == types.UntypedNil {
			return true
		}
	}
	return false
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
