package errstmt

import (
	"go/token"

	"github.com/jba/errside/ast"
	"github.com/jba/errside/types"
)

type AssignIfErrStmt struct {
	AssignStmt ast.Stmt
	IfStmt     ast.Stmt
	ErrVar     types.Object // the error variable != nil
}

func (a *AssignIfErrStmt) Pos() token.Pos { return a.AssignStmt.Pos() }
func (a *AssignIfErrStmt) End() token.Pos { return a.IfStmt.End() }
func (*AssignIfErrStmt) StmtNode()        {}
