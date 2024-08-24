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
}

func (g *GeneratorImpl) Generate(pkg *packages.Package, cs types.ConflictResolver) (code string, err error) {
	utils.Debug("Generating code for package %s %v", pkg.Name, pkg.PkgPath)

	if len(pkg.Errors) > 0 {
		utils.Debug("Package has errors: %v", pkg.Errors)
		err = errors.New("package has errors")
		return
	}

	var s strings.Builder

	g.writePackageClause(pkg, &s)

	g.writePkgTypeDecls(pkg, &s, cs)

	formated, formatErr := utils.FormatCode(s.String())
	if formatErr != nil {
		err = formatErr
		return
	}

	code = formated

	return
}

func (g *GeneratorImpl) writePackageClause(pkg *packages.Package, s *strings.Builder) {
	if len(pkg.Syntax) > 0 {
		firstFile := pkg.Syntax[0]
		s.WriteString("package ")
		s.WriteString(firstFile.Name.Name)
		s.WriteString("\n\n")
	}
}

func (g *GeneratorImpl) writePkgTypeDecls(pkg *packages.Package, s *strings.Builder, cs types.ConflictResolver) {
	//TODO: rename types which import external types
	for _, pkg := range pkg.Imports {
		g.writePkgTypeDecls(pkg, s, cs)
	}

	for _, astFile := range pkg.Syntax {
		ast.SortImports(pkg.Fset, astFile)
		for _, decl := range astFile.Decls {
			decl, isGenDecl := decl.(*ast.GenDecl)

			if !isGenDecl || decl.Tok != token.TYPE {
				continue
			}

			g.writeTypeDecl(decl, s, cs)
		}
	}
}

func (g *GeneratorImpl) writeTypeDecl(typeDecl *ast.GenDecl, s *strings.Builder, cs types.ConflictResolver) {
	for _, spec := range typeDecl.Specs {
		typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
		if !isTypeSpec {
			continue
		}
		typeSpec.Type = convertSelectorExpr(typeSpec.Type)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), typeDecl)
	s.Write(buf.Bytes())
	s.Write([]byte("\n\n"))
}

func convertSelectorExpr(expr ast.Expr) ast.Expr {
	if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		return selectorExpr.Sel
	}

	if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		starExpr.X = convertSelectorExpr(starExpr.X)
		return starExpr
	} else if structType, isStructType := expr.(*ast.StructType); isStructType {
		return convertStructTypeSelectorExpr(structType)
	}

	return expr
}

func convertStructTypeSelectorExpr(structType *ast.StructType) *ast.StructType {
	for _, field := range structType.Fields.List {
		if childStructType, isStructType := field.Type.(*ast.StructType); isStructType {
			convertStructTypeSelectorExpr(childStructType)
		} else {
			field.Type = convertSelectorExpr(field.Type)
		}
	}
	return structType
}
