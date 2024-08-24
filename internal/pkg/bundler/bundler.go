package bundler

import (
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"strings"

	"golang.org/x/tools/go/packages"
)

type bundlerImpl struct {
	Generator        types.Generator
	ConflictResolver types.ConflictResolver
}

func NewBundler() types.Bundler {
	return &bundlerImpl{
		Generator:        NewGenerator(),
		ConflictResolver: NewConflictResolver(),
	}
}

func (b *bundlerImpl) Bundle(pkg *packages.Package) (code string, err error) {
	str := strings.Builder{}
	pkgsWrapper := []*packages.Package{pkg}

	b.ConflictResolver.RegisterPkgs(pkgsWrapper)

	str.WriteString(b.Generator.GeneratePackageClause(pkg))

	packages.Visit(pkgsWrapper, func(pkg *packages.Package) bool {
		genCode, genErr := b.Generator.GenerateContent(pkg, b.ConflictResolver)
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
