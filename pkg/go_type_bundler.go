package gotypebundler

import (
	"gotypebundler/internal/pkg/bundler"
	"gotypebundler/internal/pkg/types"
	"io/fs"
	"os"

	"golang.org/x/tools/go/packages"
)

type Config struct {
	// Path to the directory containing the Go files to bundle
	Entry string
	// List of types to bundle
	Types []string
	// Path to the output file
	Output string
}

type GoTypeBundler struct {
	Bundler types.Bundler
	Config  *Config
}

func New(conf *Config) *GoTypeBundler {
	b := &GoTypeBundler{
		Bundler: bundler.NewBundler(),
		Config:  conf,
	}
	return b
}

func (b *GoTypeBundler) Bundle() (_ any, err error) {
	pkgs, loadErr := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
	}, b.Config.Entry)
	if loadErr != nil {
		err = loadErr
		return
	}

	code, bundleErr := b.Bundler.Bundle(pkgs[0], b.Config.Types)
	if bundleErr != nil {
		err = bundleErr
		return
	}

	os.WriteFile(b.Config.Output, []byte(code), fs.ModePerm)

	return

}
