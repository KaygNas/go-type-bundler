package main

import (
	"time"
)

type Type_A = time.Time
type Type_B time.Duration

type Animal struct {
	A Type_A
	B Type_B
}
