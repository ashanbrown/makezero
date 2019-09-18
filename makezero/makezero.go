// makezero provides a linter for appends to slices initialized with non-zero length.
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
)

type AppendIssue struct {
	name     string
	position token.Position
}

func (a AppendIssue) String() string {
	return fmt.Sprintf(`append to slice "%s" with non-zero initialized length at %s`, a.name, a.position)
}

type MustHaveNonZeroInitLenIssue struct {
	name     string
	position token.Position
}

func (a MustHaveNonZeroInitLenIssue) String() string {
	return fmt.Sprintf(`slice "%s" does not have non-zero initial length at %s`, a.name, a.position)
}

type Issue interface {
	String() string
}

type visitor struct {
	initLenMustBeZero bool

	comments []*ast.CommentGroup // comments to apply during this visit
	info     *types.Info

	nonZeroLengthSliceDecls map[interface{}]interface{}
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
			nonZeroLengthSliceDecls: make(map[interface{}]interface{}),
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
		fun, ok := node.Fun.(*ast.Ident)
		if !ok || fun.Name != "append" {
			break
		}
		if sliceIdent, ok := node.Args[0].(*ast.Ident); ok &&
			v.hasNonZeroInitialLength(sliceIdent) &&
			!v.hasNoLintOnSameLine(fun) {
			v.issues = append(v.issues, AppendIssue{name: sliceIdent.Name, position: v.fset.Position(fun.Pos())})
		}
	case *ast.AssignStmt:
		for i, right := range node.Rhs {
			switch right := right.(type) {
			case *ast.CallExpr:
				fun, ok := right.Fun.(*ast.Ident)
				if !ok || fun.Name != "make" {
					continue
				}
				left := node.Lhs[i]
				if len(right.Args) == 2 {
					// ignore if not a slice or it has explicit zero length
					if !v.isSlice(right.Args[0]) {
						break
					} else if lit, ok := right.Args[1].(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "0" {
						break
					}
					if v.initLenMustBeZero && !v.hasNoLintOnSameLine(fun) {
						v.issues = append(v.issues, MustHaveNonZeroInitLenIssue{name: v.textFor(left), position: v.fset.Position(node.Pos())})
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

func (v *visitor) hasNonZeroInitialLength(ident *ast.Ident) bool {
	_, exists := v.nonZeroLengthSliceDecls[ident.Obj.Decl]
	return exists
}

func (v *visitor) recordNonZeroLengthSlices(node ast.Node) {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return
	}
	v.nonZeroLengthSliceDecls[ident.Obj.Decl] = struct{}{}
}

func (v *visitor) isSlice(node ast.Node) bool {
	// determine type if this is a user-defined type
	if ident, ok := node.(*ast.Ident); ok {
		obj := ident.Obj
		if obj == nil {
			if v.info != nil {
				_, ok := v.info.ObjectOf(ident).Type().(*types.Slice)
				return ok
			} else {
				return false
			}
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

var nolint = regexp.MustCompile(`^\s*nozero\b`)

func (v *visitor) hasNoLintOnSameLine(node ast.Node) bool {
	nodePos := v.fset.Position(node.Pos())
	for _, c := range v.comments {
		commentPos := v.fset.Position(c.Pos())
		if commentPos.Line == nodePos.Line && nolint.MatchString(c.Text()) {
			return true
		}
	}
	return false
}
