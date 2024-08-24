package bundler

import (
	"gotypebundler/internal/pkg/types"

	"golang.org/x/tools/go/packages"
)

type collectorImpl struct {
}

func NewCollector() types.Collector {
	return &collectorImpl{}
}

func (c *collectorImpl) Collect(pkg *packages.Package, entryTypes []string) *packages.Package {
	return pkg
}
