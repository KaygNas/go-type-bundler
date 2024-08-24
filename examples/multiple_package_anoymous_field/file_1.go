package main

import (
	internal_1 "gotypebundler/examples/multiple_package_anoymous_field/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_anoymous_field/internal/pkg_2"
)

type Dog struct {
	DogName string
}
type Animal struct {
	A internal_2.Internal_2_Animal
	internal_1.Internal_1_Animal
	Dog
}
