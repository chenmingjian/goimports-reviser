//go:build linter
// +build linter

package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/chenmingjian/goimports-reviser/pkg/module"
	"github.com/chenmingjian/goimports-reviser/reviser"
)

const errMessage = "imports must be formatted"

func NewAnalyzer(flagSet *flag.FlagSet, localPkgPrefixes string, options ...reviser.Option) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "goimportsreviserlint",
		Doc:   "Linter for imports sorting",
		Run:   run(localPkgPrefixes, options...),
		Flags: *flagSet,
	}
}

func run(localPkgPrefixes string, options ...reviser.Option) func(pass *analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := func(formattedFile *ast.File, hasChanged bool) func(node ast.Node) bool {
			return func(node ast.Node) bool {
				file, ok := node.(*ast.File)
				if !ok {
					return true
				}
				if !hasChanged {
					return true
				}

				if len(file.Imports) != len(formattedFile.Imports) {
					pass.Reportf(
						file.Pos(),
						errMessage,
					)
				}

				for i, originalDecl := range file.Decls {
					origDd, ok := originalDecl.(*ast.GenDecl)
					if !ok {
						continue
					}

					if origDd.Tok != token.IMPORT {
						continue
					}

					if origDd != formattedFile.Decls[i] {
						pass.Reportf(
							file.Pos()+origDd.Lparen,
							errMessage,
						)
					}
				}

				return true
			}
		}

		var projectName string

		for _, f := range pass.Files {
			filePath := pass.Fset.File(f.Package).Name()

			if projectName == "" {
				var err error
				projectName, err = module.DetermineProjectName("", filePath)
				if err != nil {
					return nil, err
				}
			}

			formattedFileContent, hasChanged, err := reviser.Execute(projectName, filePath, localPkgPrefixes, options...)
			if err != nil {
				return nil, err
			}

			if !hasChanged {
				continue
			}

			formattedFile, err := parser.ParseFile(token.NewFileSet(), filePath, formattedFileContent, parser.ImportsOnly)
			if err != nil {
				panic(err)
			}

			ast.Inspect(f, inspect(formattedFile, hasChanged))
		}

		return nil, nil
	}
}
