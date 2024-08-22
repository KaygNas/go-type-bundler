package bundler

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
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

	g.writePkgTypes(pkg, &s, cs)

	formated, formatErr := g.formatCode(s.String())
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

func (g *GeneratorImpl) writePkgTypes(pkg *packages.Package, s *strings.Builder, cs types.ConflictResolver) {
	//TODO: rename types which import external types
	for _, pkg := range pkg.Imports {
		g.writePkgTypes(pkg, s, cs)
	}

	for _, astFile := range pkg.Syntax {
		ast.Inspect(astFile, func(n ast.Node) (gonext bool) {
			decl, isGenDecl := n.(*ast.GenDecl)

			if !isGenDecl || decl.Tok != token.TYPE {
				gonext = true
				return
			}

			g.writeTypeDecl(decl, s, cs)

			gonext = false
			return
		})
	}
}

func (g *GeneratorImpl) writeTypeDecl(typeDecl *ast.GenDecl, s *strings.Builder, cs types.ConflictResolver) {
	for _, spec := range typeDecl.Specs {
		typeSpec, isTypeSpec := spec.(*ast.TypeSpec)
		if !isTypeSpec {
			continue
		}

		structType, isStructType := typeSpec.Type.(*ast.StructType)
		if !isStructType {
			continue
		}

		removePackageSelector(structType)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), typeDecl)
	s.Write(buf.Bytes())
	s.Write([]byte("\n\n"))
}

// remove the package selector
func removePackageSelector(ts *ast.StructType) {
	for _, field := range ts.Fields.List {
		structType, isStructType := field.Type.(*ast.StructType)
		if isStructType {
			removePackageSelector(structType)
		} else {
			selector, isSelector := field.Type.(*ast.SelectorExpr)
			if isSelector {
				field.Type = selector.Sel
			}
		}
	}
}

func (*GeneratorImpl) formatCode(rawCode string) (string, error) {
	formated, formatErr := format.Source([]byte(rawCode))
	if formatErr != nil {
		return "", formatErr
	}

	utils.Debug("Raw code formated: \n%s", formated)

	return string(formated), nil
}
