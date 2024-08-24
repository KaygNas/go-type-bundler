package bundler

import (
	"bytes"
	"errors"
	"go/ast"
	"go/printer"
	"go/token"
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"strings"

	"golang.org/x/tools/go/packages"
)

type GeneratorImpl struct {
	ctx struct {
		pkg *packages.Package
		cs  types.ConflictResolver
		s   strings.Builder
	}
}

func (g *GeneratorImpl) Generate(pkg *packages.Package, cs types.ConflictResolver) (code string, err error) {
	utils.Debug("Generating code for package %s %v", utils.NameOfPackage(pkg), pkg.ID)

	g.ctx.pkg = pkg
	g.ctx.cs = cs
	g.ctx.s = *new(strings.Builder)

	if len(pkg.Errors) > 0 {
		utils.Debug("Package has errors: %v", pkg.Errors)
		err = errors.New("package has errors")
		return
	}

	g.writePkgTypeDecls(pkg)

	formated, formatErr := utils.FormatCode(g.ctx.s.String())
	if formatErr != nil {
		err = formatErr
		return
	}

	code = formated

	return
}

func (g *GeneratorImpl) GeneratePackageClause(pkg *packages.Package) string {
	pkgName := utils.NameOfPackage(pkg)
	packageClaus := ""

	if pkgName != "" {
		packageClaus = "package " + pkgName + "\n\n"
	}

	return packageClaus
}

func (g *GeneratorImpl) writePkgTypeDecls(pkg *packages.Package) {
	for _, astFile := range pkg.Syntax {
		ast.SortImports(pkg.Fset, astFile)
		selectorToPkg := utils.CreateSelectorToPkg(astFile, pkg)
		for _, decl := range astFile.Decls {
			decl, isGenDecl := decl.(*ast.GenDecl)

			if !isGenDecl || decl.Tok != token.TYPE {
				continue
			}

			g.writeTypeDecl(decl, selectorToPkg)
		}
	}
}

func (g *GeneratorImpl) writeTypeDecl(typeDecl *ast.GenDecl, selectorToPkg utils.SelectorToPkg) {
	for _, spec := range typeDecl.Specs {
		typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
		if !isTypeSpec {
			continue
		}
		typeSpec.Name.Name = g.ctx.cs.ResolveIdentName(types.PkgID(g.ctx.pkg.ID), typeSpec.Name.Name)
		typeSpec.Type = g.convertSelectorExpr(typeSpec.Type, selectorToPkg)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), typeDecl)
	g.ctx.s.Write(buf.Bytes())
	g.ctx.s.Write([]byte("\n\n"))
}

func (g *GeneratorImpl) convertSelectorExpr(expr ast.Expr, selectorToPkg utils.SelectorToPkg) ast.Expr {
	if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		sel := selectorExpr.Sel
		selector := selectorExpr.X.(*ast.Ident).Name
		selectorPkg := selectorToPkg[selector]
		sel.Name = g.ctx.cs.ResolveIdentName(types.PkgID(selectorPkg.ID), sel.Name)
		return selectorExpr.Sel
	}

	if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		starExpr.X = g.convertSelectorExpr(starExpr.X, selectorToPkg)
		return starExpr
	} else if structType, isStructType := expr.(*ast.StructType); isStructType {
		return g.convertStructTypeSelectorExpr(structType, selectorToPkg)
	}

	return expr
}

func (g *GeneratorImpl) convertStructTypeSelectorExpr(structType *ast.StructType, selectorToPkg utils.SelectorToPkg) *ast.StructType {
	for _, field := range structType.Fields.List {
		if childStructType, isStructType := field.Type.(*ast.StructType); isStructType {
			g.convertStructTypeSelectorExpr(childStructType, selectorToPkg)
		} else {
			field.Type = g.convertSelectorExpr(field.Type, selectorToPkg)
		}
	}
	return structType
}
