package main

import (
	internal_1 "gotypebundler/examples/multiple_package_array/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_array/internal/pkg_2"
)

type LocalType_A = internal_1.Internal_1_Animal

type Animal struct {
	A [5]internal_1.Internal_1_Animal
	B [5][5]internal_2.Internal_2_Animal
	C [5]LocalType_A
}
