package main

import (
	internal_1 "gotypebundler/examples/multiple_package_lit_struct/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_lit_struct/internal/pkg_2"
)

type Animal struct {
	Internal struct {
		A internal_1.Internal_1_Animal
		B internal_2.Internal_2_Animal
	}
}
