package main

import (
	"time"
)

type Type_A = time.Duration
type Type_B time.Weekday

type Animal struct {
	A Type_A
	B Type_B
}
