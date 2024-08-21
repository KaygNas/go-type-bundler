package bundler

import (
	"gotypebundler/internal/pkg/types"

	"golang.org/x/tools/go/packages"
)

type GeneratorImpl struct {
}

func (g *GeneratorImpl) Generate(pkg *packages.Package, cs types.ConflictResolver) (code string, err error) {
	return
}
