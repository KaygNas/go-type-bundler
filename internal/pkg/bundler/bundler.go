package bundler

import (
	"gotypebundler/internal/pkg/utils"
	"strings"

	"golang.org/x/tools/go/packages"
)

type BundlerImpl struct {
}

func (b *BundlerImpl) Bundle(pkgs []*packages.Package) (code string, err error) {

	str := strings.Builder{}
	cs := NewConflictResolver()
	generator := &GeneratorImpl{}

	cs.RegisterPkgs(pkgs)

	if len(pkgs) > 0 {
		str.WriteString(generator.GeneratePackageClause(pkgs[0]))
	}

	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		genCode, genErr := generator.Generate(pkg, cs)
		if genErr != nil {
			err = genErr
			return false
		}

		_, writeErr := str.WriteString(genCode)
		if writeErr != nil {
			err = writeErr
			return false
		}
		return true
	}, nil)

	formated, formatErr := utils.FormatCode(str.String())
	if formatErr != nil {
		err = formatErr
		return
	}

	code = formated

	return
}
