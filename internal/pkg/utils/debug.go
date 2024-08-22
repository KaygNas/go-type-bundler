package utils

import "fmt"

func Debug(format string, args ...any) (n int, err error) {
	return fmt.Printf("[debug] "+format+"\n", args...)
}
