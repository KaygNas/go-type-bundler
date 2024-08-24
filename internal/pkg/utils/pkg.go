package utils

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func NameOfPackage(pkg *packages.Package) string {
	pkgName := pkg.Name

	if pkgName == "" && len(pkg.Syntax) > 0 {
		firstFile := pkg.Syntax[0]
		pkgName = firstFile.Name.Name
	}

	return pkgName
}

type SelectorToPkg = map[string]*packages.Package

// create a map of selector to package
func CreateSelectorToPkg(astFile *ast.File, pkg *packages.Package) SelectorToPkg {
	selectorToPkg := make(SelectorToPkg)
	for _, imp := range astFile.Imports {
		rawPath := imp.Path.Value
		path := rawPath[1 : len(rawPath)-1] // remove quotes
		importedPkg := pkg.Imports[path]
		importedName := ""
		if imp.Name == nil {
			importedName = NameOfPackage(importedPkg)
		} else {
			importedName = imp.Name.Name
		}
		selectorToPkg[importedName] = importedPkg
	}
	return selectorToPkg
}
