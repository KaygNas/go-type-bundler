package bundler

import (
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"strconv"

	"golang.org/x/tools/go/packages"
)

type ConflictResolverImpl struct {
	pkgNameSpaces     map[types.PkgID]types.PkgNameSpace
	pkgNameSpaceStore map[types.PkgNameSpace]map[types.PkgID]int
}

func NewConflictResolver() *ConflictResolverImpl {
	return &ConflictResolverImpl{
		pkgNameSpaces:     make(map[types.PkgID]types.PkgNameSpace),
		pkgNameSpaceStore: make(map[types.PkgNameSpace]map[types.PkgID]int),
	}
}

func (cr *ConflictResolverImpl) RegisterPkgs(pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		importPkgs := utils.CollectMapValues(pkg.Imports)
		cr.RegisterPkgs(importPkgs)
		cr.RegisterPkgNameSpace(types.PkgID(pkg.ID), types.PkgNameSpace(utils.NameOfPackage(pkg)))
	}
}

func (cr *ConflictResolverImpl) RegisterPkgNameSpace(pkgId types.PkgID, pkgNs types.PkgNameSpace) {
	ns, isExist := cr.pkgNameSpaces[pkgId]

	if !isExist {
		ns = pkgNs
		cr.pkgNameSpaceStore[ns] = make(map[types.PkgID]int)
		cr.pkgNameSpaces[pkgId] = ns
	}
	store := cr.pkgNameSpaceStore[ns]
	_, isExistInStore := store[pkgId]
	if !isExistInStore {
		store[pkgId] = len(store)
	}
}

func (cr *ConflictResolverImpl) ResolveIdentName(pkgId types.PkgID, name string) (newName string) {
	ns, isNsExist := cr.pkgNameSpaces[pkgId]
	newName = name

	if !isNsExist {
		return
	}

	store, isStoreExist := cr.pkgNameSpaceStore[ns]
	if !isStoreExist {
		return
	}

	idx, isExistInStore := store[pkgId]
	if !isExistInStore {
		return
	}

	newName = string(ns)
	if idx > 0 {
		newName += strconv.Itoa(idx)
	}
	newName += "_" + name
	return
}
