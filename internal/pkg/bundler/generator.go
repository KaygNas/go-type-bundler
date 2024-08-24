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
		pkg              *packages.Package
		cs               types.ConflictResolver
		s                strings.Builder
		typeSpecRenaming map[string]string
	}
}

func (g *GeneratorImpl) Generate(pkg *packages.Package, cs types.ConflictResolver) (code string, err error) {
	utils.Debug("Generating code for package %s %v", utils.NameOfPackage(pkg), pkg.ID)

	g.ctx.pkg = pkg
	g.ctx.cs = cs
	g.ctx.s = *new(strings.Builder)
	g.ctx.typeSpecRenaming = make(map[string]string)

	if len(pkg.Errors) > 0 {
		utils.Debug("Package has errors: %v", pkg.Errors)
		err = errors.New("package has errors")
		return
	}

	g.writePkgTypeDecls(pkg)

	code = g.ctx.s.String()

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
	typeSpecs := make([]*ast.TypeSpec, 0, len(typeDecl.Specs))

	for _, spec := range typeDecl.Specs {
		typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
		if !isTypeSpec {
			continue
		}
		typeSpecs = append(typeSpecs, typeSpec)

		// setting up the renaming map
		oldName := typeSpec.Name.Name
		newName := g.ctx.cs.ResolveIdentName(types.PkgID(g.ctx.pkg.ID), oldName)
		if _, isExist := g.ctx.typeSpecRenaming[oldName]; isExist {
			utils.Warn("Type %s already renamed to %s", oldName, newName)
		} else {
			g.ctx.typeSpecRenaming[oldName] = newName
		}
	}

	for _, typeSpec := range typeSpecs {
		typeSpec.Name.Name = g.ctx.typeSpecRenaming[typeSpec.Name.Name]
		typeSpec.Type = g.convertExpr(typeSpec.Type, selectorToPkg)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), typeDecl)
	g.ctx.s.Write(buf.Bytes())
	g.ctx.s.Write([]byte("\n\n"))
}

func (g *GeneratorImpl) convertExpr(expr ast.Expr, selectorToPkg utils.SelectorToPkg) ast.Expr {
	if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		sel := selectorExpr.Sel
		selector := selectorExpr.X.(*ast.Ident).Name
		selectorPkg := selectorToPkg[selector]
		sel.Name = g.ctx.cs.ResolveIdentName(types.PkgID(selectorPkg.ID), sel.Name)
		return selectorExpr.Sel
	}

	if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		starExpr.X = g.convertExpr(starExpr.X, selectorToPkg)
		return starExpr
	}

	if structType, isStructType := expr.(*ast.StructType); isStructType {
		return g.convertStructTypeExpr(structType, selectorToPkg)
	}

	// the type spec is rename in the writeTypeDecl function
	// therefore any ident that was renamed should be rename here
	if identType, isIdentType := expr.(*ast.Ident); isIdentType {
		newName, isExist := g.ctx.typeSpecRenaming[identType.Name]
		if isExist {
			identType.Name = newName
		}
		return identType
	}

	return expr
}

func (g *GeneratorImpl) convertStructTypeExpr(structType *ast.StructType, selectorToPkg utils.SelectorToPkg) *ast.StructType {
	for _, field := range structType.Fields.List {
		if childStructType, isStructType := field.Type.(*ast.StructType); isStructType {
			g.convertStructTypeExpr(childStructType, selectorToPkg)
		} else {
			field.Type = g.convertExpr(field.Type, selectorToPkg)
		}
	}
	return structType
}
