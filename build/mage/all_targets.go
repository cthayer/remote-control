package mage

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

func Lint() error {
	fmt.Println("running go lint checks")

	out, err := sh.OutCmd("go", "fmt", "./...")()

	if err == nil {
		fmt.Println(out)
	}

	return err
}

func Fmt() error {
	return Lint()
}

func Test() error {
	fmt.Println("running go test with coverage")

	out, err := sh.OutCmd("go", "test", "./...", "-cover")()

	fmt.Println(out)

	return err
}

func Tidy() error {
	fmt.Println("running go mod tidy")

	out, err := sh.OutCmd("go", "mod", "tidy")()

	fmt.Println(out)

	return err
}
