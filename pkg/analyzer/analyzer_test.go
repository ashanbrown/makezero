package analyzer_test

import (
	"testing"

	"github.com/ashanbrown/makezero/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAppend(t *testing.T) {
	testdata := analysistest.TestData()
	a := analyzer.NewAnalyzer()
	analysistest.Run(t, testdata, a, "./append")
}

func TestAlways(t *testing.T) {
	testdata := analysistest.TestData()
	a := analyzer.NewAnalyzer()
	err := a.Flags.Set("always", "true")
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
	analysistest.Run(t, testdata, a, "./always")
}
