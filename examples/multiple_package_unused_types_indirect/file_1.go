package main

import (
	internal_1 "gotypebundler/examples/multiple_package_unused_types_indirect/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_unused_types_indirect/internal/pkg_2"
)

type Type_A = internal_1.Internal_1_Animal
type Type_B internal_2.Internal_2_Animal

type Animal struct {
	A Type_A
	B Type_B
}
