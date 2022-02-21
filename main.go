package main

import (
	"flag"
	"go/ast"
	"log"
	"os"

	"golang.org/x/tools/go/packages"

	"github.com/ashanbrown/makezero/makezero"
)

func main() {
	log.SetFlags(0) // remove log timestamp

	setExitStatus := flag.Bool("set_exit_status", false,
		"Set exit status to 1 if any issues are found")
	always := flag.Bool("always", false,
		"require every make to have zero length regardless of whether append is used")
	flag.Parse()

	cfg := packages.Config{
		Mode: packages.NeedSyntax | packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(&cfg, os.Args[1:]...)
	if err != nil {
		log.Fatalf("Could not load packages: %s", err)
	}
	linter := makezero.NewLinter(*always)

	var issues []makezero.Issue
	for _, p := range pkgs {
		nodes := make([]ast.Node, 0, len(p.Syntax))
		for _, n := range p.Syntax {
			nodes = append(nodes, n)
		}
		newIssues, err := linter.Run(p.Fset, p.TypesInfo, nodes...)
		if err != nil {
			log.Fatalf("failed: %s", err)
		}
		issues = append(issues, newIssues...)
	}

	for _, issue := range issues {
		log.Println(issue)
	}

	if *setExitStatus && len(issues) > 0 {
		os.Exit(1)
	}
}
