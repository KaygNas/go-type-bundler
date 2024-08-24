package types

import "golang.org/x/tools/go/packages"

type Bundler interface {
	Bundle(pkgs []*packages.Package) (code string, err error)
}

type PkgID string
type PkgNameSpace string
type ConflictResolver interface {
	RegisterPkgNameSpace(pkgId PkgID, ns PkgNameSpace)
	// Resolve conflict names in the package
	ResolveIdentName(pkgId PkgID, name string) (newName string, err error)
}

type Generator interface {
	Generate(pkg *packages.Package, cs ConflictResolver) (code string, err error)
}
