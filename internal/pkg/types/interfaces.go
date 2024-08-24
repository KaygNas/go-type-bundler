package types

import "golang.org/x/tools/go/packages"

type Bundler interface {
	Bundle(pkg *packages.Package) (code string, err error)
}

type PkgID string
type PkgNameSpace string
type ConflictResolver interface {
	RegisterPkgs(pkgs []*packages.Package)
	// Resolve conflict names in the package
	ResolveIdentName(pkgId PkgID, name string) (newName string)
}

type Generator interface {
	GeneratePackageClause(pkg *packages.Package) string
	GenerateContent(pkg *packages.Package, cs ConflictResolver) (code string, err error)
}
