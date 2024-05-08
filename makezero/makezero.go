// Package makezero provides a linter for appends to slices initialized with non-zero length.
package makezero

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"regexp"
	"strings"
)

// a decl might include multiple var,
// so var name with decl make final uniq obj.
type uniqDecl struct {
	varName string
	decl    interface{}
}

type Issue interface {
	Details() string
	Pos() token.Pos
	Position() token.Position
	String() string
}

type AppendIssue struct {
	name     string
	pos      token.Pos
	position token.Position

	calledFuncs []string
	indexAssign bool
}

func (a AppendIssue) Details() string {
	details := make([]string, 1)
	details[0] = fmt.Sprintf("append to slice `%s` with non-zero initialized length", a.name)
	if len(a.calledFuncs) > 0 {
		details = append(details, fmt.Sprintf("and called by funcs: %s", strings.Join(a.calledFuncs, ",")))
	}
	if a.indexAssign {
		details = append(details, "and has assigned by index")
	}
	return strings.Join(details, ", ")
}

func (a AppendIssue) Pos() token.Pos {
	return a.pos
}

func (a AppendIssue) Position() token.Position {
	return a.position
}

func (a AppendIssue) String() string { return toString(a) }

type MustHaveNonZeroInitLenIssue struct {
	name     string
	pos      token.Pos
	position token.Position
}

func (i MustHaveNonZeroInitLenIssue) Details() string {
	return fmt.Sprintf("slice `%s` does not have non-zero initial length", i.name)
}

func (i MustHaveNonZeroInitLenIssue) Pos() token.Pos {
	return i.pos
}

func (i MustHaveNonZeroInitLenIssue) Position() token.Position {
	return i.position
}

func (i MustHaveNonZeroInitLenIssue) String() string { return toString(i) }

func toString(i Issue) string {
	return fmt.Sprintf("%s at %s", i.Details(), i.Position())
}

type visitor struct {
	initLenMustBeZero bool

	comments []*ast.CommentGroup // comments to apply during this visit
	info     *types.Info

	nonZeroLengthSliceDecls map[uniqDecl]struct{}
	funcCallDecls           map[uniqDecl][]string
	indexAssignDecls        map[uniqDecl]bool
	fset                    *token.FileSet
	issues                  []Issue
}

type Linter struct {
	initLenMustBeZero bool
}

func NewLinter(initialLengthMustBeZero bool) *Linter {
	return &Linter{
		initLenMustBeZero: initialLengthMustBeZero,
	}
}

func (l Linter) Run(fset *token.FileSet, info *types.Info, nodes ...ast.Node) ([]Issue, error) {
	var issues []Issue
	for _, node := range nodes {
		var comments []*ast.CommentGroup
		if file, ok := node.(*ast.File); ok {
			comments = file.Comments
		}
		visitor := visitor{
			nonZeroLengthSliceDecls: make(map[uniqDecl]struct{}),
			funcCallDecls:           make(map[uniqDecl][]string),
			indexAssignDecls:        make(map[uniqDecl]bool),
			initLenMustBeZero:       l.initLenMustBeZero,
			info:                    info,
			fset:                    fset,
			comments:                comments,
		}
		ast.Walk(&visitor, node)
		issues = append(issues, visitor.issues...)
	}
	return issues, nil
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.CallExpr:
		fun, isAppend := v.isAppendFunc(node)
		if !isAppend {
			v.recordFuncCall(node)
			break
		}

		if sliceIdent, ok := node.Args[0].(*ast.Ident); ok {
			obj, ok := v.hasNonZeroInitialLength(sliceIdent)
			if ok && !v.hasNoLintOnSameLine(fun) {
				v.issues = append(v.issues,
					AppendIssue{
						name:     sliceIdent.Name,
						pos:      fun.Pos(),
						position: v.fset.Position(fun.Pos()),

						calledFuncs: v.funcCallDecls[obj],
						indexAssign: v.indexAssignDecls[obj],
					})
			}
		}
	case *ast.AssignStmt:
		v.recordIndexAssign(node)
		for i, right := range node.Rhs {
			if right, ok := right.(*ast.CallExpr); ok {
				fun, ok := right.Fun.(*ast.Ident)
				if !ok || fun.Name != "make" {
					continue
				}
				left := node.Lhs[i]
				if len(right.Args) == 2 {
					// ignore if not a slice or it has explicit zero length
					if !v.isSlice(right.Args[0]) {
						continue
					} else if lit, ok := right.Args[1].(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "0" {
						continue
					}
					if v.initLenMustBeZero && !v.hasNoLintOnSameLine(fun) {
						v.issues = append(v.issues, MustHaveNonZeroInitLenIssue{
							name:     v.textFor(left),
							pos:      node.Pos(),
							position: v.fset.Position(node.Pos()),
						})
					}
					v.recordNonZeroLengthSlices(left)
				}
			}
		}
	}
	return v
}

func (v *visitor) textFor(node ast.Node) string {
	typeBuf := new(bytes.Buffer)
	if err := printer.Fprint(typeBuf, v.fset, node); err != nil {
		log.Fatalf("ERROR: unable to print type: %s", err)
	}
	return typeBuf.String()
}

func (v *visitor) getIdentDecl(node ast.Node) (uniqDecl, bool) {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return uniqDecl{}, false
	}
	if ident.Obj == nil {
		return uniqDecl{}, false
	}
	return uniqDecl{
		varName: ident.Obj.Name,
		decl:    ident.Obj.Decl,
	}, true
}

func (v *visitor) hasNonZeroInitialLength(ident *ast.Ident) (uniqDecl, bool) {
	obj, ok := v.getIdentDecl(ident)
	if !ok {
		log.Printf("WARNING: could not determine with %q at %s is a slice (missing object type)",
			ident.Name, v.fset.Position(ident.Pos()).String())
		return uniqDecl{}, false
	}
	_, exists := v.nonZeroLengthSliceDecls[obj]
	return obj, exists
}

func (v *visitor) recordNonZeroLengthSlices(node ast.Node) {
	obj, ok := v.getIdentDecl(node)
	if !ok {
		return
	}
	v.nonZeroLengthSliceDecls[obj] = struct{}{}
}

func (v *visitor) recordFuncCall(node *ast.CallExpr) {
	funcName := v.textFor(node.Fun)
	for _, arg := range node.Args {
		obj, ok := v.getIdentDecl(arg)
		if !ok {
			continue
		}
		v.funcCallDecls[obj] = append(v.funcCallDecls[obj], funcName)
	}
}

func (v *visitor) isAppendFunc(node *ast.CallExpr) (fun *ast.Ident, ok bool) {
	fun, ok = node.Fun.(*ast.Ident)
	if !ok {
		return nil, false
	}
	if fun.Name != "append" {
		return nil, false
	}
	return fun, true
}

func (v *visitor) recordIndexAssign(node *ast.AssignStmt) {
	for _, left := range node.Lhs {
		var x ast.Expr
		switch expr := left.(type) {
		case *ast.IndexExpr:
			x = expr.X
		case *ast.IndexListExpr:
			x = expr.X
		}
		if x == nil {
			continue
		}
		obj, ok := v.getIdentDecl(x)
		if ok {
			v.indexAssignDecls[obj] = true
		}
	}
}

func (v *visitor) isSlice(node ast.Node) bool {
	// determine type if this is a user-defined type
	if ident, ok := node.(*ast.Ident); ok {
		obj := ident.Obj
		if obj == nil {
			if v.info != nil {
				_, ok := v.info.ObjectOf(ident).Type().(*types.Slice)
				return ok
			}
			return false
		}
		spec, ok := obj.Decl.(*ast.TypeSpec)
		if !ok {
			return false
		}
		node = spec.Type
	}

	if node, ok := node.(*ast.ArrayType); ok {
		return node.Len == nil // only slices have zero length
	}
	return false
}

func (v *visitor) hasNoLintOnSameLine(node ast.Node) bool {
	nolint := regexp.MustCompile(`^\s*nozero\b`)
	nodePos := v.fset.Position(node.Pos())
	for _, c := range v.comments {
		commentPos := v.fset.Position(c.Pos())
		if commentPos.Line == nodePos.Line && nolint.MatchString(c.Text()) {
			return true
		}
	}
	return false
}
