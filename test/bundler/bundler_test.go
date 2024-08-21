package test

import (
	"gotypebundler/internal/pkg/bundler"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestBundlerImpl(t *testing.T) {
	bundler := &bundler.BundlerImpl{}
	pkgs, pkgErr := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles,
	}, "gotypebundler/examples/bookstore")

	if pkgErr != nil {
		t.Errorf("Error: %v", pkgErr)
	}

	code, bundleErr := bundler.Bundle(pkgs)
	if bundleErr != nil {
		t.Errorf("Error: %v", bundleErr)
	}

	file, _ := filepath.Abs("../../examples/bookstore/expected.code")
	expected, readErr := os.ReadFile(file)
	if readErr != nil {
		t.Errorf("Error: %v", readErr)
	}

	if code != string(expected) {
		t.Errorf("Expected:\n%v\nGot:\n%v", string(expected), code)
	}
}
