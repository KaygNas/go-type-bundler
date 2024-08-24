package bundler

import (
	"errors"
	"gotypebundler/internal/pkg/types"
	"strconv"
)

type ConflictResolverImpl struct {
	pkgNameSpaces     map[types.PkgID]types.PkgNameSpace
	pkgNameSpaceStore map[types.PkgNameSpace]map[types.PkgID]int
}

func (cr *ConflictResolverImpl) RegisterPkgNameSpace(pkgId types.PkgID, pkgNs types.PkgNameSpace) {
	ns, isExist := cr.pkgNameSpaces[pkgId]

	if !isExist {
		cr.pkgNameSpaceStore[ns] = make(map[types.PkgID]int)
	}
	store := cr.pkgNameSpaceStore[ns]
	_, isExistInStore := store[pkgId]
	if !isExistInStore {
		store[pkgId] = len(store)
	}
}

func (cr *ConflictResolverImpl) ResolveIdentName(pkgId types.PkgID, name string) (newName string, err error) {
	ns, isNsExist := cr.pkgNameSpaces[pkgId]

	if !isNsExist {
		err = errors.New("namespace not found")
		return
	}

	store, isStoreExist := cr.pkgNameSpaceStore[ns]
	if !isStoreExist {
		err = errors.New("store not found")
		return
	}

	idx, isExistInStore := store[pkgId]
	if !isExistInStore {
		err = errors.New("index not found")
		return
	}

	newName = string(ns)
	if idx > 0 {
		newName += strconv.Itoa(idx)
	}
	newName += "_" + name
	return
}
