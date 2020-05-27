package mage

import (
	"errors"
	"fmt"
	"os"

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

func Build() error {
	var err error

	target := os.Getenv("TARGET")
	version := os.Getenv("VERSION")

	if version == "" {
		version = "dev"
	}

	if target != "remote-control" && target != "rc" {
		return errors.New("must set the 'TARGET' environment variable to one of: 'rc' or 'remote-control'")
	}

	err = buildBinary(target, version)

	if err != nil {
		return err
	}

	// create SHA256 sums
	err = getBinarySha256Sum(target, version)

	if err != nil {
		return err
	}

	// create zip archives for each platform
	err = compressBinary(target, version)

	return err
}

func Clean() error {
	fmt.Println("cleaning up any prior builds")

	out, err := sh.OutCmd("rm", "-rf", "build/bin")()

	fmt.Println(out)

	return err
}
