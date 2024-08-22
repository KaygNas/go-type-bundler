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

	formated, formatErr := g.formatCode(s.String())
	if formatErr != nil {
		err = formatErr
		return
	}

	code = formated

	return
}

func (*GeneratorImpl) formatCode(rawCode string) (string, error) {
	utils.Debug("Formating Raw code: \n%s", rawCode)

	formated, formatErr := format.Source([]byte(rawCode))
	if formatErr != nil {
		return "", formatErr
	}

	return string(formated), nil
}

func (g *GeneratorImpl) writePackageClause(pkg *packages.Package, s *strings.Builder) {
	if len(pkg.Syntax) > 0 {
		firstFile := pkg.Syntax[0]
		s.WriteString("package ")
		s.WriteString(firstFile.Name.Name)
		s.WriteString("\n\n")
	}
}
