package errstmt

import (
	"go/token"

	"github.com/jba/errside/ast"
	"github.com/jba/errside/types"
)

type AssignIfErrStmt struct {
	FirstStmt ast.Stmt
	IfStmt    *ast.IfStmt
	ErrVar    *ast.Ident // the error variable != nil
}

func NewAssignIfErrStmt(aStmt *ast.AssignStmt, iStmt *ast.IfStmt, evar types.Object) *AssignIfErrStmt {
	llen := len(aStmt.Lhs)
	a := &AssignIfErrStmt{
		IfStmt: iStmt,
		ErrVar: aStmt.Lhs[llen-1].(*ast.Ident),
	}
	if len(aStmt.Lhs) > 1 {
		aStmt.Lhs = aStmt.Lhs[:llen-1]
		a.FirstStmt = aStmt
	} else {
		a.FirstStmt = &ast.ExprStmt{X: aStmt.Rhs[0]}
	}
	return a
}

func (a *AssignIfErrStmt) Pos() token.Pos { return a.FirstStmt.Pos() }
func (a *AssignIfErrStmt) End() token.Pos { return a.IfStmt.End() }
func (*AssignIfErrStmt) StmtNode()        {}
