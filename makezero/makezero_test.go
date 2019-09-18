package makezero

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeZero(t *testing.T) {
	t.Run("when append is used", func(t *testing.T) {
		t.Run("finds appends to non-zero length initialized slices", func(t *testing.T) {
			linter := NewLinter(false)
			expectIssues(t, linter, `
package bar

func foo() []int {
  x := make([]int, 5)
  append(x, 1)
}`, "append to slice `x` with non-zero initialized length at testing.go:6:3")
		})

		t.Run("works with custom types that are slices", func(t *testing.T) {
			linter := NewLinter(false)
			expectIssues(t, linter, `
package bar

type intSlice []int
func foo() {
  x := make(intSlice, 5)
  append(x, 1)
}`, "append to slice `x` with non-zero initialized length at testing.go:7:3")
		})

		t.Run("can report any initializes without length", func(t *testing.T) {
			linter := NewLinter(true)
			expectIssues(t, linter, `
package bar

func foo() {
  x := make([]int, 5)
}`, "slice `x` does not have non-zero initial length at testing.go:5:3")
		})

		t.Run("doesn't confuse maps with slices", func(t *testing.T) {
			linter := NewLinter(true)
			expectIssues(t, linter, `
package bar

func foo() {
  x := make(map[string]int, 5)
}`)
		})

		t.Run("can report any initializes without length", func(t *testing.T) {
			linter := NewLinter(true)
			expectIssues(t, linter, `
package bar

func foo() {
  x := make([]int, 5) // nozero
  append(x, 1) //nozero
  append(x, 1) //nozeroxxx
}`, "append to slice `x` with non-zero initialized length at testing.go:7:3")
		})

	})

	t.Run("ignores more complex constructs than basic variables", func(t *testing.T) {
		linter := NewLinter(false)
		expectIssues(t, linter, `
package bar

func foo() {
	var x [][]int
  x[0] = make([]int, 5)
}`)
	})
}

func expectIssues(t *testing.T, linter *Linter, contents string, issues ...string) {
	actualIssues := parseFile(t, linter, contents)
	actualIssueStrs := make([]string, 0, len(actualIssues))
	for _, i := range actualIssues {
		actualIssueStrs = append(actualIssueStrs, i.String())
	}
	assert.ElementsMatch(t, issues, actualIssueStrs)
}

func parseFile(t *testing.T, linter *Linter, contents string) []Issue {
	fset := token.NewFileSet()
	expr, err := parser.ParseFile(fset, "testing.go", contents, parser.ParseComments)
	if err != nil {
		t.Fatalf("unable to parse file contents: %s", err)
	}
	issues, err := linter.Run(fset, nil, expr)
	if err != nil {
		t.Fatalf("unable to parse file: %s", err)
	}
	return issues
}
