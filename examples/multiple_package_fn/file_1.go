package main

import (
	internal_1 "gotypebundler/examples/multiple_package_fn/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_fn/internal/pkg_2"
)

type Animal struct {
	A func(internal_1.Internal_1_Animal)
	B func(func(param1 internal_1.Internal_1_Animal)) internal_2.Internal_2_Animal
}
