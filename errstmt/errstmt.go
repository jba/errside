package errstmt

import (
	"go/token"

	"github.com/jba/errside/ast"
	"github.com/jba/errside/types"
)

type AssignIfErrStmt struct {
	assignStmt ast.Stmt
	ifStmt     ast.Stmt
	errVar     types.Object // the error variable != nil
}

func (a *AssignIfErrStmt) Pos() token.Pos { return a.assignStmt.Pos() }
func (a *AssignIfErrStmt) End() token.Pos { return a.ifStmt.End() }
