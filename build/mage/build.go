package mage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	OS "os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/magefile/mage/sh"
)

const (
	SHA256_CHECKSUM_FILE_PERMS     = 0644
	SHA256_CHECKSUM_FILE_EXTENSION = ".sha256.checksum"
)

var (
	buildOsArchs []string = []string{
		"darwin/386",
		"darwin/amd64",
		"freebsd/386",
		"freebsd/amd64",
		"freebsd/arm",
		"linux/386",
		"linux/amd64",
		"linux/arm",
		"linux/arm64",
		"linux/mips64",
		"linux/mips64le",
		"linux/mips",
		"linux/mipsle",
		"linux/s390x",
		"netbsd/386",
		"netbsd/amd64",
		"netbsd/arm",
		"openbsd/386",
		"openbsd/amd64",
		"windows/386",
		"windows/amd64",
	}

	buildRoot string = filepath.Join("build", "bin")
)

func buildBinary(target string, version string) error {
	fmt.Println("building binary: " + target)

	cmdArgs := []string{
		"-output", filepath.Join(buildRoot, "{{.OS}}_{{.Arch}}_"+version, "{{.Dir}}"),
		"-ldflags", "-X 'main.VERSION=" + version + "'",
		"-osarch", strings.Join(buildOsArchs, " "),
		"./cmd/" + target,
	}

	out, err := sh.OutCmd("gox", cmdArgs...)()

	fmt.Println(out)

	return err
}

func getBinarySha256Sum(target string, version string) error {
	var err *multierror.Error
	var errs []error

	fmt.Println("calculating Sha256 sums for binary: " + target)

	waitGroup := sync.WaitGroup{}
	outChan := make(chan string, len(buildOsArchs))
	errChan := make(chan error, len(buildOsArchs))

	for _, osarch := range buildOsArchs {
		os, arch := splitOsArch(osarch)

		if os == "" || arch == "" {
			errChan <- errors.New("Cannot split osarch: " + osarch)
			continue
		}

		binaryName := filepath.Join(buildRoot, strings.Join([]string{os, arch, version}, "_"), target)
		shaFileName := filepath.Join(buildRoot, strings.Join([]string{target, version, os, arch}, "_")+SHA256_CHECKSUM_FILE_EXTENSION)

		if os == "windows" {
			binaryName = binaryName + ".exe"
		}

		waitGroup.Add(1)
		go func(fn string, sfn string) {
			defer waitGroup.Done()

			f, err := OS.Open(fn)

			if err != nil {
				errChan <- err
				return
			}

			defer f.Close()

			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				errChan <- err
				return
			}

			shaSum := fmt.Sprintf("%x", h.Sum(nil))

			err = ioutil.WriteFile(sfn, []byte(shaSum), SHA256_CHECKSUM_FILE_PERMS)

			errChan <- err
			outChan <- "Sha256 Sum for " + fn + ": " + string(shaSum)

			//fmt.Printf("%x", h.Sum(nil))
		}(binaryName, shaFileName)
	}

	// wait for go routines to finish
	waitGroup.Wait()

	// close channels
	close(outChan)
	close(errChan)

	// print output
	for o := range outChan {
		fmt.Println(o)
	}

	// process errors
	for e := range errChan {
		if e == nil {
			continue
		}

		errs = append(errs, e)
	}

	if len(errs) < 1 {
		return nil
	}

	err = multierror.Append(err, errs...)

	return err
}

func compressBinary(target string, version string) error {
	var err *multierror.Error
	var errs []error

	fmt.Println("compressing binary: " + target)

	waitGroup := sync.WaitGroup{}
	outChan := make(chan string, len(buildOsArchs))
	errChan := make(chan error, len(buildOsArchs))

	for _, osarch := range buildOsArchs {
		os, arch := splitOsArch(osarch)

		if os == "" || arch == "" {
			outChan <- ""
			errChan <- errors.New("Cannot split osarch: " + osarch)
			continue
		}

		zipFile := filepath.Join(buildRoot, strings.Join([]string{target, version, os, arch}, "_")+".zip")
		zipContent := filepath.Join(buildRoot, strings.Join([]string{os, arch, version}, "_"), target)

		if os == "windows" {
			zipContent = zipContent + ".exe"
		}

		waitGroup.Add(1)
		go func(zf string, zc string) {
			defer waitGroup.Done()

			cmdArgs := []string{
				"-j",
				zf,
				zc,
			}

			out, err := sh.OutCmd("zip", cmdArgs...)()

			outChan <- out
			errChan <- err
		}(zipFile, zipContent)
	}

	// wait for go routines to finish
	waitGroup.Wait()

	// close channels
	close(outChan)
	close(errChan)

	// print output
	for o := range outChan {
		fmt.Println(o)
	}

	// process errors
	for e := range errChan {
		if e == nil {
			continue
		}

		errs = append(errs, e)
	}

	if len(errs) < 1 {
		return nil
	}

	err = multierror.Append(err, errs...)

	return err
}

func splitOsArch(osarch string) (os string, arch string) {
	osarchParts := strings.Split(osarch, "/")

	return osarchParts[0], osarchParts[1]
}
