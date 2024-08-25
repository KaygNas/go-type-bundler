package main

import (
	internal_1 "gotypebundler/examples/multiple_package_map/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_map/internal/pkg_2"
)

type LocalType_A = internal_1.Internal_1_Animal
type Animal struct {
	A map[string]internal_1.Internal_1_Animal
	B map[string]map[int]internal_1.Internal_1_Animal
	C map[internal_1.Internal_1_Animal]internal_2.Internal_2_Animal
	D map[string]map[internal_1.Internal_1_Animal]internal_2.Internal_2_Animal
	E map[string]LocalType_A
}
