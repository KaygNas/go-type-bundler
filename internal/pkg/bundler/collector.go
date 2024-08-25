package bundler

import (
	"go/ast"
	"go/token"
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"

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

func (c *collectorImpl) Collect(pkg *packages.Package, entryTypes []string) *packages.Package {
	requiredTypeDecls := c.collectTypeDeclsByNames(pkg, entryTypes)
	newPkg := c.filterTypeSpecs(pkg, requiredTypeDecls)
	return newPkg
}

// collectTypeDeclsByNames collects the type declarations by the given type names.
// this func will recursively collect the type declarations from the imported packages.
func (c *collectorImpl) collectTypeDeclsByNames(pkg *packages.Package, typeNames []string) requiredTypeSpecs {
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

	// foreach there decl
	// if the decl has selector expression which meas it is an imported package
	// then collect the type declarations from the imported package
	// and add the collected type declarations to the typeDecls
	for _, requiredPkg := range requiredPkgs {
		requiredTypeSpecsFromPkg := c.collectTypeDeclsByNames(requiredPkg.pkg, requiredPkg.typeNames)
		for typeSpec := range requiredTypeSpecsFromPkg {
			requiredTypeSpecs[typeSpec] = true
		}
	}

	return requiredTypeSpecs
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

// filterTypeSpecs filters out the declarations that are not required.
func (c *collectorImpl) filterTypeSpecs(pkg *packages.Package, requiredTypeSpecs requiredTypeSpecs) *packages.Package {
	if len(requiredTypeSpecs) == 0 {
		return pkg
	}

	newImports := make(map[string]*packages.Package)
	for key, imp := range pkg.Imports {
		newImports[key] = c.filterTypeSpecs(imp, requiredTypeSpecs)
	}
	pkg.Imports = newImports

	for _, astFile := range pkg.Syntax {
		newDecls := make([]ast.Decl, 0)
		for i := 0; i < len(astFile.Decls); i++ {
			if genDecl, ok := astFile.Decls[i].(*ast.GenDecl); ok {
				typeSpecs := make([]ast.Spec, 0)
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if _, required := requiredTypeSpecs[typeSpec]; required {
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

	return pkg
}
