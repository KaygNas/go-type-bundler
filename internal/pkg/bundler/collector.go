package bundler

import (
	"go/ast"
	"go/token"
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"sort"

	"golang.org/x/tools/go/packages"
)

type collectorImpl struct {
}

type requiredTypeSpecs map[*ast.TypeSpec]bool

type requiredPkg struct {
	pkg       *packages.Package
	typeNames []string
}

func NewCollector() types.Collector {
	return &collectorImpl{}
}

func (c *collectorImpl) Collect(pkg *packages.Package, entryTypes []string) []*packages.Package {
	if len(entryTypes) == 0 {
		entryTypes = c.collectAllTypes(pkg)
	}
	requiredPackages := c.collectRequiredPackages(pkg, entryTypes)
	requiredPackages = append(requiredPackages, &requiredPkg{pkg: pkg, typeNames: entryTypes})
	requiredPackages = c.tidyRequiredPkgs(requiredPackages)
	newPkgs := c.filterTypeSpecs(requiredPackages)
	return newPkgs
}

func (c *collectorImpl) collectAllTypes(pkg *packages.Package) []string {
	typeNames := make([]string, 0)
	for _, astFile := range pkg.Syntax {
		for _, decl := range astFile.Decls {
			genDecl, isGenDecl := decl.(*ast.GenDecl)

			if !isGenDecl {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeNames = append(typeNames, typeSpec.Name.Name)
				}
			}
		}
	}
	return typeNames
}

// collectRequiredPackages collects the type declarations by the given type names.
// this func will recursively collect the type declarations from the imported packages.
func (c *collectorImpl) collectRequiredPackages(pkg *packages.Package, typeNames []string) []*requiredPkg {
	requiredTypeNames := make(map[string]bool)
	for _, typeName := range typeNames {
		requiredTypeNames[typeName] = true
	}

	requiredTypeSpecs := make(requiredTypeSpecs)
	requiredPkgs := make([]*requiredPkg, 0)
	// foreach files of the package
	// foreach decls of the file
	// if the decl is a type declaration and the name of the type is in the given types
	// then add the decl to the typeDecls
	for _, astFile := range pkg.Syntax {
		selectorToPkg := utils.CreateSelectorToPkg(astFile, pkg)

		for _, decl := range astFile.Decls {
			genDecl, isGenDecl := decl.(*ast.GenDecl)

			if !isGenDecl {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := requiredTypeNames[typeSpec.Name.Name]; ok {
						requiredTypeSpecs[typeSpec] = true
						requiredPkgs = append(requiredPkgs, c.collectPkgFromExpr(pkg, typeSpec.Type, selectorToPkg)...)
					}
				}
			}
		}
	}

	for _, requiredPkg := range requiredPkgs {
		requiredTypeSpecsFromPkg := c.collectRequiredPackages(requiredPkg.pkg, requiredPkg.typeNames)
		requiredPkgs = append(requiredPkgs, requiredTypeSpecsFromPkg...)
	}

	return requiredPkgs
}

func (c *collectorImpl) collectPkgFromExpr(pkg *packages.Package, expr ast.Expr, selectorToPkg utils.SelectorToPkg) []*requiredPkg {
	if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		requiredPkgs := make([]*requiredPkg, 0)
		sel := selectorExpr.Sel
		selector := selectorExpr.X.(*ast.Ident).Name
		selectorPkg, isSelectorPkgExist := selectorToPkg[selector]
		if !isSelectorPkgExist {
			utils.Warn("Selector package %s not found", selector)
			return requiredPkgs
		}
		requiredPkgs = append(requiredPkgs, &requiredPkg{
			pkg:       selectorPkg,
			typeNames: []string{sel.Name},
		})
		return requiredPkgs
	}

	if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		return c.collectPkgFromExpr(pkg, starExpr.X, selectorToPkg)
	}

	if structType, isStructType := expr.(*ast.StructType); isStructType {
		return c.collectPkgFromStructTypeExpr(pkg, structType, selectorToPkg)
	}

	if arrayType, isArrayType := expr.(*ast.ArrayType); isArrayType {
		return c.collectPkgFromExpr(pkg, arrayType.Elt, selectorToPkg)
	}

	if sliceType, isSliceType := expr.(*ast.SliceExpr); isSliceType {
		return c.collectPkgFromExpr(pkg, sliceType.X, selectorToPkg)
	}

	if mapType, isMapType := expr.(*ast.MapType); isMapType {
		requiredPkgs := make([]*requiredPkg, 0)
		requiredPkgs = append(requiredPkgs, c.collectPkgFromExpr(pkg, mapType.Key, selectorToPkg)...)
		requiredPkgs = append(requiredPkgs, c.collectPkgFromExpr(pkg, mapType.Value, selectorToPkg)...)
		return requiredPkgs
	}

	if ident, isIdent := expr.(*ast.Ident); isIdent {
		requiredPkgs := make([]*requiredPkg, 0)
		requiredPkgs = append(requiredPkgs, &requiredPkg{
			pkg:       pkg,
			typeNames: []string{ident.Name},
		})
		return requiredPkgs
	}

	return []*requiredPkg{}
}

func (c *collectorImpl) collectPkgFromStructTypeExpr(pkg *packages.Package, structType *ast.StructType, selectorToPkg utils.SelectorToPkg) []*requiredPkg {
	requiredPkgs := make([]*requiredPkg, 0)
	for _, field := range structType.Fields.List {
		if childStructType, isStructType := field.Type.(*ast.StructType); isStructType {
			pkgs := c.collectPkgFromStructTypeExpr(pkg, childStructType, selectorToPkg)
			requiredPkgs = append(requiredPkgs, pkgs...)
		} else {
			pkgs := c.collectPkgFromExpr(pkg, field.Type, selectorToPkg)
			requiredPkgs = append(requiredPkgs, pkgs...)
		}
	}
	return requiredPkgs
}

func (c *collectorImpl) tidyRequiredPkgs(requiredPkgs []*requiredPkg) []*requiredPkg {
	pkgMap := make(map[string]*requiredPkg)
	for _, requiredPkg := range requiredPkgs {
		if _, ok := pkgMap[requiredPkg.pkg.ID]; !ok {
			pkgMap[requiredPkg.pkg.ID] = requiredPkg
		} else {
			mergedTypeNames := pkgMap[requiredPkg.pkg.ID].typeNames
			mergedTypeNames = append(mergedTypeNames, requiredPkg.typeNames...)
			pkgMap[requiredPkg.pkg.ID].typeNames = mergedTypeNames
		}
	}

	tidyRequiredPkgs := make([]*requiredPkg, 0)

	pkgIds := make([]string, 0, len(pkgMap))
	for pkgId := range pkgMap {
		pkgIds = append(pkgIds, pkgId)
	}
	sort.StringSlice(pkgIds).Sort()

	for _, id := range pkgIds {
		tidyRequiredPkgs = append(tidyRequiredPkgs, pkgMap[id])
	}

	return tidyRequiredPkgs
}

// filterTypeSpecs filters out the declarations that are not required.
func (c *collectorImpl) filterTypeSpecs(requiredPackages []*requiredPkg) []*packages.Package {
	if len(requiredPackages) == 0 {
		return []*packages.Package{}
	}

	pkgs := make([]*packages.Package, 0)
	for _, requiredPkg := range requiredPackages {
		pkg := requiredPkg.pkg
		requiredNames := make(map[string]bool)
		for _, typeName := range requiredPkg.typeNames {
			requiredNames[typeName] = true
		}

		for _, astFile := range pkg.Syntax {
			newDecls := make([]ast.Decl, 0)
			for i := 0; i < len(astFile.Decls); i++ {
				if genDecl, ok := astFile.Decls[i].(*ast.GenDecl); ok {

					// import declarations should be included in the new file
					// so that the selector expressions can be resolved.
					if genDecl.Tok == token.IMPORT {
						newDecls = append(newDecls, genDecl)
						continue
					}

					typeSpecs := make([]ast.Spec, 0)
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if _, required := requiredNames[typeSpec.Name.Name]; required {
								typeSpecs = append(typeSpecs, typeSpec)
							}
						}
					}
					if len(typeSpecs) > 0 {
						genDecl.Specs = typeSpecs
						newDecls = append(newDecls, genDecl)
					}
				}
			}
			astFile.Decls = newDecls
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs
}
