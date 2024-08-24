package main

import (
	internal_1 "gotypebundler/examples/multiple_package_star/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_star/internal/pkg_2"
)

type Animal struct {
	A internal_1.Internal_1_Animal
	B *internal_2.Internal_2_Animal
}
