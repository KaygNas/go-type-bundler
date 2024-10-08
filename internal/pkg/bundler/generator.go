package bundler

import (
	"bytes"
	"errors"
	"go/ast"
	"go/printer"
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"strings"

	"golang.org/x/tools/go/packages"
)

type generatorImpl struct {
	types.Generator
	ctx struct {
		pkg              *packages.Package
		cr               types.ConflictResolver
		s                strings.Builder
		typeSpecRenaming map[string]string
	}
}

func NewGenerator() types.Generator {
	return &generatorImpl{}
}

func (g *generatorImpl) GenerateContent(pkg *packages.Package, cs types.ConflictResolver) (code string, err error) {
	utils.Debug("Generating code for package %s %v", utils.NameOfPackage(pkg), pkg.ID)

	g.ctx.pkg = pkg
	g.ctx.cr = cs
	g.ctx.s = strings.Builder{}
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

func (g *generatorImpl) GeneratePackageClause(pkg *packages.Package) string {
	pkgName := utils.NameOfPackage(pkg)
	packageClaus := ""

	if pkgName != "" {
		packageClaus = "package " + pkgName + "\n\n"
	}

	return packageClaus
}

type TypeDeclStoreItem struct {
	selectorToPkg utils.SelectorToPkg
	typeDecls     []*ast.GenDecl
}

func (g *generatorImpl) writePkgTypeDecls(pkg *packages.Package) {
	store := make([]*TypeDeclStoreItem, 0)
	for _, astFile := range pkg.Syntax {
		selectorToPkg := utils.CreateSelectorToPkg(astFile, pkg)
		typeDecls := utils.CollectTypeDeclsFromAstFile(astFile)
		store = append(store, &TypeDeclStoreItem{
			selectorToPkg: selectorToPkg,
			typeDecls:     typeDecls,
		})
	}

	for _, item := range store {
		for _, decl := range item.typeDecls {
			g.setupTypeSpecRenaming(decl)
		}
	}

	for _, item := range store {
		for _, decl := range item.typeDecls {
			g.writeTypeDecl(decl, item.selectorToPkg)
		}
	}
}

func (g *generatorImpl) setupTypeSpecRenaming(typeDecl *ast.GenDecl) {
	typeSpecs := utils.CollectTypeSpecsFromDecl(typeDecl)
	for _, typeSpec := range typeSpecs {
		// setting up the renaming map
		oldName := typeSpec.Name.Name
		newName := g.ctx.cr.ResolveIdentName(types.PkgID(g.ctx.pkg.ID), oldName)
		if _, isExist := g.ctx.typeSpecRenaming[oldName]; isExist {
			utils.Warn("Type %s already renamed to %s", oldName, newName)
		} else {
			g.ctx.typeSpecRenaming[oldName] = newName
		}
	}
}

func (g *generatorImpl) writeTypeDecl(typeDecl *ast.GenDecl, selectorToPkg utils.SelectorToPkg) {
	typeSpecs := utils.CollectTypeSpecsFromDecl(typeDecl)

	for _, typeSpec := range typeSpecs {
		typeSpec.Name.Name = g.ctx.typeSpecRenaming[typeSpec.Name.Name]
		typeSpec.Type = g.convertExpr(typeSpec.Type, selectorToPkg)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, g.ctx.pkg.Fset, typeDecl)
	g.ctx.s.Write(buf.Bytes())
	g.ctx.s.Write([]byte("\n\n"))
}

func (g *generatorImpl) convertExpr(expr ast.Expr, selectorToPkg utils.SelectorToPkg) ast.Expr {
	if selectorExpr, isSelectorExpr := expr.(*ast.SelectorExpr); isSelectorExpr {
		sel := selectorExpr.Sel
		selector := selectorExpr.X.(*ast.Ident).Name
		selectorPkg, isSelectorPkgExist := selectorToPkg[selector]
		if !isSelectorPkgExist {
			utils.Warn("Selector package %s not found", selector)
			return expr
		}
		sel.Name = g.ctx.cr.ResolveIdentName(types.PkgID(selectorPkg.ID), sel.Name)
		return selectorExpr.Sel
	}

	if starExpr, isStarExpr := expr.(*ast.StarExpr); isStarExpr {
		starExpr.X = g.convertExpr(starExpr.X, selectorToPkg)
		return starExpr
	}

	if structType, isStructType := expr.(*ast.StructType); isStructType {
		return g.convertStructTypeExpr(structType, selectorToPkg)
	}

	if arrayType, isArrayType := expr.(*ast.ArrayType); isArrayType {
		arrayType.Elt = g.convertExpr(arrayType.Elt, selectorToPkg)
		return arrayType
	}

	if sliceType, isSliceType := expr.(*ast.SliceExpr); isSliceType {
		sliceType.X = g.convertExpr(sliceType.X, selectorToPkg)
		return sliceType
	}

	if mapType, isMapType := expr.(*ast.MapType); isMapType {
		mapType.Key = g.convertExpr(mapType.Key, selectorToPkg)
		mapType.Value = g.convertExpr(mapType.Value, selectorToPkg)
		return mapType
	}

	if funcType, isFuncType := expr.(*ast.FuncType); isFuncType {
		return g.convertFuncTypeExpr(funcType, selectorToPkg)
	}

	if chanType, isChanType := expr.(*ast.ChanType); isChanType {
		chanType.Value = g.convertExpr(chanType.Value, selectorToPkg)
		return chanType
	}

	if interfaceType, isInterfaceType := expr.(*ast.InterfaceType); isInterfaceType {
		return g.convertInerfaceTypeExpr(interfaceType, selectorToPkg)
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

func (g *generatorImpl) convertStructTypeExpr(structType *ast.StructType, selectorToPkg utils.SelectorToPkg) *ast.StructType {
	for _, field := range structType.Fields.List {
		if childStructType, isStructType := field.Type.(*ast.StructType); isStructType {
			g.convertStructTypeExpr(childStructType, selectorToPkg)
		} else {
			field.Type = g.convertExpr(field.Type, selectorToPkg)
		}
	}
	return structType
}

func (g *generatorImpl) convertFuncTypeExpr(funcType *ast.FuncType, selectorToPkg utils.SelectorToPkg) *ast.FuncType {
	if funcType.Params != nil {
		for _, field := range funcType.Params.List {
			field.Type = g.convertExpr(field.Type, selectorToPkg)
		}
	}
	if funcType.Results != nil {
		for _, field := range funcType.Results.List {
			field.Type = g.convertExpr(field.Type, selectorToPkg)
		}
	}
	return funcType
}

func (g *generatorImpl) convertInerfaceTypeExpr(interfaceType *ast.InterfaceType, selectorToPkg utils.SelectorToPkg) *ast.InterfaceType {
	for _, method := range interfaceType.Methods.List {
		method.Type = g.convertExpr(method.Type, selectorToPkg)
	}
	return interfaceType
}
