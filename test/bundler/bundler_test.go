package test

import (
	"gotypebundler/internal/pkg/bundler"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/tools/go/packages"
)

func TestBundlerImpl(t *testing.T) {
	runTestCase(t, "single_package_single_file")
	runTestCase(t, "single_package_multiple_file")
}

func runTestCase(t *testing.T, exampleName string) {
	t.Run(exampleName, func(t *testing.T) {
		bundler := &bundler.BundlerImpl{}
		pkgs, pkgErr := packages.Load(&packages.Config{
			Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
		}, "gotypebundler/examples/"+exampleName)

		if pkgErr != nil {
			t.Errorf("Fail to load packages. Error: %v", pkgErr)
		}

		code, bundleErr := bundler.Bundle(pkgs)
		if bundleErr != nil {
			t.Errorf("Failed to bundle. Error: %v", bundleErr)
		}

		file, _ := filepath.Abs("../../examples/" + exampleName + "/expected.code")
		expected, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Errorf("Failed to read file. Error: %v", readErr)
		}

		if code != string(expected) {

			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(expected)),
				B:        difflib.SplitLines(code),
				FromFile: "Expected",
				ToFile:   "Got",
				Context:  3,
			}
			diffStr, _ := difflib.GetUnifiedDiffString(diff)

			t.Error("\n" + diffStr)
		}
	})
}
