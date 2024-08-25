package utils

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

func NameOfPackage(pkg *packages.Package) string {
	pkgName := pkg.Name

	if pkgName == "" {
		if len(pkg.Syntax) > 0 {
			firstFile := pkg.Syntax[0]
			pkgName = firstFile.Name.Name
		} else {
			pkgName = pkg.ID
		}
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

func CollectTypeDeclsFromAstFile(astFile *ast.File) (genDecls []*ast.GenDecl) {
	genDecls = make([]*ast.GenDecl, 0)
	for _, decl := range astFile.Decls {
		genDecl, isGenDecl := decl.(*ast.GenDecl)
		if !isGenDecl || genDecl.Tok != token.TYPE {
			continue
		}
		genDecls = append(genDecls, genDecl)
	}
	return
}

func CollectTypeSpecsFromDecl(genDecl *ast.GenDecl) (typeSpecs []*ast.TypeSpec) {
	typeSpecs = make([]*ast.TypeSpec, 0)
	if genDecl.Tok != token.TYPE {
		return
	}

	for _, spec := range genDecl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			typeSpecs = append(typeSpecs, typeSpec)
		}
	}
	return
}
