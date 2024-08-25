package bundler_test

import (
	"gotypebundler/internal/pkg/bundler"
	"gotypebundler/internal/pkg/utils"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/tools/go/packages"
)

func TestSinglePackageSingleFile(t *testing.T) {
	runTestCase(t, "single_package_single_file")
}
func TestSinglePackageMultipleFile(t *testing.T) {
	runTestCase(t, "single_package_multiple_file")
}
func TestMultiplePackageDeep(t *testing.T) {
	runTestCase(t, "multiple_package_deep")
}
func TestMultiplePackageSameName(t *testing.T) {
	runTestCase(t, "multiple_package_same_name")
}
func TestMultiplePackageLitStruct(t *testing.T) {
	runTestCase(t, "multiple_package_lit_struct")
}
func TestMultiplePackageAlias(t *testing.T) {
	runTestCase(t, "multiple_package_alias")
}
func TestMultiplePackagePointer(t *testing.T) {
	runTestCase(t, "multiple_package_pointer")
}
func TestMultiplePackageAnoymousField(t *testing.T) {
	runTestCase(t, "multiple_package_anoymous_field")
}
func TestMultiplePackageArray(t *testing.T) {
	runTestCase(t, "multiple_package_array")
}
func TestMultiplePackageSlice(t *testing.T) {
	runTestCase(t, "multiple_package_slice")
}
func TestMultiplePackageMap(t *testing.T) {
	runTestCase(t, "multiple_package_map")
}
func TestMultiplePackageUnusedTypes(t *testing.T) {
	runTestCase(t, "multiple_package_unused_types", "Animal")
}
func TestMultiplePackageUnusedTypesIndirect(t *testing.T) {
	runTestCase(t, "multiple_package_unused_types_indirect", "Animal")
}
func TestMultiplePackageThirdParty(t *testing.T) {
	runTestCase(t, "multiple_package_third_party", "Animal")
}

func runTestCase(t *testing.T, exampleName string, entryTypes ...string) {
	t.Run(exampleName, func(t *testing.T) {
		root := "../../examples/"

		bundler := bundler.NewBundler()
		pkgs, pkgErr := packages.Load(&packages.Config{
			Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedDeps | packages.NeedImports,
		}, path.Join(root, exampleName))

		if pkgErr != nil {
			t.Errorf("Fail to load packages. Error: %v", pkgErr)
			return
		}

		code, bundleErr := bundler.Bundle(pkgs[0], entryTypes)
		if bundleErr != nil {
			t.Errorf("Failed to bundle. Error: %v", bundleErr)
			return
		}

		file, _ := filepath.Abs(root + exampleName + "/expected.code")
		expected, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Errorf("Failed to read file. Error: %v", readErr)
			return
		}

		formatedExpected, formatErr := utils.FormatCode(string(expected))
		if formatErr != nil {
			t.Errorf("Failed to format expected code. Error: %v", formatErr)
			return
		}

		utils.Debug("Expected:\n%v\n", formatedExpected)
		utils.Debug("Got:\n%v\n", code)

		if code != formatedExpected {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(formatedExpected),
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
