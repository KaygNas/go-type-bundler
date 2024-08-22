package bundler

import (
	"bytes"
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

	var s strings.Builder

	g.writePackageClause(pkg, &s)

	for _, astFile := range pkg.Syntax {
		ast.Inspect(astFile, func(n ast.Node) (gonext bool) {
			decl, isGenDecl := n.(*ast.GenDecl)

			if !isGenDecl || decl.Tok != token.TYPE {
				gonext = true
				return
			}

			var buf bytes.Buffer
			printer.Fprint(&buf, pkg.Fset, decl)
			s.Write(buf.Bytes())
			s.Write([]byte("\n\n"))

			gonext = false
			return
		})
	}

	formated, formatErr := g.newMethod(s.String())
	if formatErr != nil {
		err = formatErr
		return
	}

	code = formated

	return
}

func (*GeneratorImpl) newMethod(rawCode string) (string, error) {
	utils.Debug("Formating Raw code: \n%s", rawCode)

	formated, formatErr := format.Source([]byte(rawCode))
	if formatErr != nil {
		return "", formatErr
	}

	return string(formated), nil
}

func (g *GeneratorImpl) writePackageClause(pkg *packages.Package, s *strings.Builder) {
	firstFile := pkg.Syntax[0]
	if firstFile != nil {
		s.WriteString("package ")
		s.WriteString(pkg.Syntax[0].Name.Name)
		s.WriteString("\n\n")
	}
}
