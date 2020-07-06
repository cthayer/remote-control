package mage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	OS "os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/magefile/mage/sh"
)

const (
	SHA256_CHECKSUM_FILE_PERMS     = 0644
	SHA256_CHECKSUM_FILE_EXTENSION = ".sha256.checksum"

	BUILD_OS_ARCHES_ARM_PATTERN = "^arm(64)?v\\d+$"
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

	// builds that require a specific version of GOARM to be set
	// (use "v" to separate the arch from the GOARM value)
	buildOsArchsArm []string = []string{
		"freebsd/armv5",
		"linux/armv5",
		"netbsd/armv5",
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

	if err != nil {
		return err
	}

	fmt.Println("building GOARM specific arm binaries: " + target)

	for i := 0; i < len(buildOsArchsArm); i++ {
		os, arch := splitOsArch(buildOsArchsArm[i])

		if os == "" || arch == "" {
			fmt.Println("skipping " + buildOsArchsArm[i])
			continue
		}

		if match, err := regexp.Match(BUILD_OS_ARCHES_ARM_PATTERN, []byte(arch)); !match || err != nil {
			fmt.Println("skipping " + buildOsArchsArm[i])
			continue
		}

		armArch, goArm := splitArmArch(arch)

		if armArch == "" || goArm == "" {
			fmt.Println("skipping " + buildOsArchsArm[i])
			continue
		}

		fmt.Println("building " + buildOsArchsArm[i])

		cmdArgs := []string{
			"-output", filepath.Join(buildRoot, "{{.OS}}_{{.Arch}}v" + goArm + "_"+version, "{{.Dir}}"),
			"-ldflags", "-X 'main.VERSION=" + version + "'",
			"-osarch", strings.Join([]string{os, armArch}, "/"),
			"./cmd/" + target,
		}

		env := make(map[string]string)

		env["GOARM"] = goArm

		out, err := sh.OutputWith(env,"gox", cmdArgs...)

		fmt.Println(out)

		if err != nil {
			return err
		}
	}

	return nil
}

func getBinarySha256Sum(target string, version string) error {
	var err *multierror.Error
	var errs []error

	fmt.Println("calculating Sha256 sums for binary: " + target)

	osArchs := append(buildOsArchs, buildOsArchsArm...)

	waitGroup := sync.WaitGroup{}
	outChan := make(chan string, len(osArchs))
	errChan := make(chan error, len(osArchs))

	for _, osarch := range osArchs {
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

	osArchs := append(buildOsArchs, buildOsArchsArm...)

	waitGroup := sync.WaitGroup{}
	outChan := make(chan string, len(osArchs))
	errChan := make(chan error, len(osArchs))

	for _, osarch := range osArchs {
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

	if len(osarchParts) != 2 {
		return "", ""
	}

	return osarchParts[0], osarchParts[1]
}

func splitArmArch(armArch string) (arch string, goarm string) {
	armArchParts := strings.Split(armArch, "v")

	if len(armArchParts) != 2 {
		return "", ""
	}

	return armArchParts[0], armArchParts[1]
}
