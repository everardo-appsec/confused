package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"net/http"
)

// PipenvLookup represents a collection of python packages to be tested for dependency confusion.
type PipenvLookup struct {
	Packages []string
	Verbose  bool
}

// NewPipenvLookup constructs a `PipenvLookup` struct and returns it
func NewPipenvLookup(verbose bool) PackageResolver {
	return &PipenvLookup{Packages: []string{}, Verbose: verbose}
}

// ReadPackagesFromFile reads package information from a python `Pipenv` file.
// Only the `packages` and `dev-packages` sections are considered.
//
// Returns any errors encountered
func (p *PipenvLookup) ReadPackagesFromFile(filename string) error {
	rawfile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	config, err := toml.Load(string(rawfile))
	if err != nil {
		return err
	}
	packages := config.Get("packages")
	if packages != nil {
		p.Packages = append(p.Packages, packages.(*toml.Tree).Keys()...)
	}
	dev_packages := config.Get("dev-packages")
	if dev_packages != nil {
		p.Packages = append(p.Packages, dev_packages.(*toml.Tree).Keys()...)
	}
	return nil
}

// PackagesNotInPublic determines if a python package does not exist in the pypi package repository.
//
// Returns a slice of strings with any python packages not in the pypi package repository
func (p *PipenvLookup) PackagesNotInPublic() []string {
	notavail := []string{}
	for _, pkg := range p.Packages {
		if !p.isAvailableInPublic(pkg) {
			notavail = append(notavail, pkg)
		}
	}
	return notavail
}

// isAvailableInPublic determines if a python package exists in the pypi package repository.
//
// Returns true if the package exists in the pypi package repository.
func (p *PipenvLookup) isAvailableInPublic(pkgname string) bool {
	if p.Verbose {
		fmt.Print("Checking: https://pypi.org/project/" + pkgname + "/ : ")
	}
	resp, err := http.Get("https://pypi.org/project/" + pkgname + "/")
	if err != nil {
		fmt.Printf(" [W] Error when trying to request https://pypi.org/project/"+pkgname+"/ : %s\n", err)
		return false
	}
	if p.Verbose {
		fmt.Printf("%s\n", resp.Status)
	}
	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}
