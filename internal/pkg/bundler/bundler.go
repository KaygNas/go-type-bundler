package bundler

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

type BundlerImpl struct {
}

func (b *BundlerImpl) Bundle(pkgs []*packages.Package) (code string, err error) {

	str := strings.Builder{}
	cs := &ConflictResolverImpl{}
	generator := &GeneratorImpl{}

	for _, pkg := range pkgs {
		genCode, genErr := generator.Generate(pkg, cs)
		if genErr != nil {
			err = genErr
			return
		}

		_, writeErr := str.WriteString(genCode)
		if writeErr != nil {
			err = writeErr
			return
		}
	}

	code = str.String()

	return
}
