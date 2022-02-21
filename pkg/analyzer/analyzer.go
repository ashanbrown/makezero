package analyzer

import (
	"flag"
	"go/ast"

	"github.com/ashanbrown/makezero/makezero"
	"golang.org/x/tools/go/analysis"
)

type analyzer struct {
	always bool
}

// NewAnalyzer returns a go/analysis-compatible analyzer
// Set "-always" to report any non-empty slice initialization.
func NewAnalyzer() *analysis.Analyzer {
	var flags flag.FlagSet
	a := analyzer{}
	flags.BoolVar(&a.always, "always", false, "report any non-empty slice initializations, regardless of intention")
	return &analysis.Analyzer{
		Name:  "makezero",
		Doc:   "detect unintended non-empty slice initializations",
		Run:   a.runAnalysis,
		Flags: flags,
	}
}

func (a *analyzer) runAnalysis(pass *analysis.Pass) (interface{}, error) {
	linter := makezero.NewLinter(a.always)
	nodes := make([]ast.Node, 0, len(pass.Files))
	for _, f := range pass.Files {
		nodes = append(nodes, f)
	}
	issues, err := linter.Run(pass.Fset, pass.TypesInfo, nodes...)
	if err != nil {
		return nil, err
	}
	reportIssues(pass, issues)
	return nil, nil
}

func reportIssues(pass *analysis.Pass, issues []makezero.Issue) {
	for _, i := range issues {
		diag := analysis.Diagnostic{
			Pos:      i.Pos(),
			Message:  i.Details(),
			Category: "restriction",
		}
		pass.Report(diag)
	}
}
