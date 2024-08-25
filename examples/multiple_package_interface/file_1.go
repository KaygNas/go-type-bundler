package main

import (
	internal_1 "gotypebundler/examples/multiple_package_interface/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_interface/internal/pkg_2"
)

type Animal struct {
	A interface {
		Fn(internal_1.Internal_1_Animal)
	}
	B interface {
		interface {
			Fn(internal_1.Internal_1_Animal) internal_2.Internal_2_Animal
		}
	}
}
