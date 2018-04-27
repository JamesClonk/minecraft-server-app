package env

import (
	"fmt"
	"os"
)

func Get(key string, nvl string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return nvl
	}
	return value
}

func MustGet(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Required ENV variable [%s] is missing!", key))
	}
	return value
}
