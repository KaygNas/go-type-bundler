package main

import (
	internal_1 "gotypebundler/examples/multiple_package_same_name/internal/pkg_1"
	internal_2 "gotypebundler/examples/multiple_package_same_name/internal/pkg_2"
)

type InternalAnimals struct {
	A internal_1.InternalAnimal
	B internal_2.InternalAnimal
}
