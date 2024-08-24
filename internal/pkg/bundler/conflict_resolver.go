package bundler

import (
	"gotypebundler/internal/pkg/types"
	"gotypebundler/internal/pkg/utils"
	"strconv"

	"golang.org/x/tools/go/packages"
)

type conflictResolverImpl struct {
	pkgNameSpaces     map[types.PkgID]types.PkgNameSpace
	pkgNameSpaceStore map[types.PkgNameSpace]map[types.PkgID]int
}

func NewConflictResolver() types.ConflictResolver {
	return &conflictResolverImpl{
		pkgNameSpaces:     make(map[types.PkgID]types.PkgNameSpace),
		pkgNameSpaceStore: make(map[types.PkgNameSpace]map[types.PkgID]int),
	}
}

func (cr *conflictResolverImpl) RegisterPkgs(pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		importPkgs := utils.CollectMapValues(pkg.Imports)
		cr.RegisterPkgs(importPkgs)
		cr.registerPkgNameSpace(types.PkgID(pkg.ID), types.PkgNameSpace(utils.NameOfPackage(pkg)))
	}
}

func (cr *conflictResolverImpl) registerPkgNameSpace(pkgId types.PkgID, pkgNs types.PkgNameSpace) {
	ns, isExist := cr.pkgNameSpaces[pkgId]

	if !isExist {
		if pkgNs == "" {
			ns = types.PkgNameSpace(pkgId)
		} else {
			ns = pkgNs
		}
		cr.pkgNameSpaces[pkgId] = ns
	}

	store, isStoreExist := cr.pkgNameSpaceStore[ns]
	if !isStoreExist {
		store = make(map[types.PkgID]int)
		cr.pkgNameSpaceStore[ns] = store
	}

	_, isExistInStore := store[pkgId]
	if !isExistInStore {
		store[pkgId] = len(store)
	}
}

func (cr *conflictResolverImpl) ResolveIdentName(pkgId types.PkgID, name string) (newName string) {
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

	newName = name + "_" + string(ns)
	if idx > 0 {
		newName += strconv.Itoa(idx)
	}
	return
}
