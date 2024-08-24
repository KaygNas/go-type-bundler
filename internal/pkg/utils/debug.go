package utils

import "fmt"

func Debug(format string, args ...any) (n int, err error) {
	return fmt.Printf("[debug] "+format+"\n", args...)
}

func Warn(format string, args ...any) (n int, err error) {
	return fmt.Printf("[warn] "+format+"\n", args...)
}
