package main

import (
	internal_1 "gotypebundler/examples/multiple_package_chan/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_chan/internal/pkg_2"
)

type Animal struct {
	A <-chan internal_1.Internal_1_Animal
	B <-chan chan internal_2.Internal_2_Animal
}
