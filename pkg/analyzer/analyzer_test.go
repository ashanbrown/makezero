package analyzer_test

import (
	"testing"

	"github.com/ashanbrown/makezero/pkg/analyzer"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, a.Flags.Set("always", "true"))
	analysistest.Run(t, testdata, a, "./always")
}
