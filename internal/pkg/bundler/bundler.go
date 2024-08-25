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
	Collector        types.Collector
}

func NewBundler() types.Bundler {
	return &bundlerImpl{
		Generator:        NewGenerator(),
		ConflictResolver: NewConflictResolver(),
		Collector:        NewCollector(),
	}
}

func (b *bundlerImpl) Bundle(pkg *packages.Package, entryTypes []string) (code string, err error) {
	pkgs := b.Collector.Collect(pkg, entryTypes)
	str := strings.Builder{}

	b.ConflictResolver.RegisterPkgs(pkgs)

	str.WriteString(b.Generator.GeneratePackageClause(pkg))

	packages.Visit(pkgs, func(pkg *packages.Package) bool {
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
