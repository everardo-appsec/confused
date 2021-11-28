package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// PipLookup represents a collection of python packages to be tested for dependency confusion.
type PipLookup struct {
	Packages []string
	Verbose  bool
}

// NewPipLookup constructs a `PipLookup` struct and returns it
func NewPipLookup(verbose bool) PackageResolver {
	return &PipLookup{Packages: []string{}, Verbose: verbose}
}

// ReadPackagesFromFile reads package information from a python `requirements.txt` file
//
// Returns any errors encountered
func (p *PipLookup) ReadPackagesFromFile(filename string) error {
	rawfile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	line := ""
	for _, l := range strings.Split(string(rawfile), "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			continue
		}
		if len(l) > 0 {
			// Support line continuation
			if strings.HasSuffix(l, "\\") {
				line += l[:len(l) - 1]
				continue
			}
			line += l
			pkgrow := strings.FieldsFunc(line, p.pipSplit)
			if len(pkgrow) > 0 {
				p.Packages = append(p.Packages, strings.TrimSpace(pkgrow[0]))
			}
			// reset the line variable
			line = ""
		}
	}
	return nil
}

// PackagesNotInPublic determines if a python package does not exist in the pypi package repository.
//
// Returns a slice of strings with any python packages not in the pypi package repository
func (p *PipLookup) PackagesNotInPublic() []string {
	notavail := []string{}
	for _, pkg := range p.Packages {
		if !p.isAvailableInPublic(pkg) {
			notavail = append(notavail, pkg)
		}
	}
	return notavail
}

func (p *PipLookup) pipSplit(r rune) bool {
	delims := []rune{
		'=',
		'<',
		'>',
		'!',
		' ',
		'~',
		'#',
		'[',
	}
	return inSlice(r, delims)
}

// isAvailableInPublic determines if a python package exists in the pypi package repository.
//
// Returns true if the package exists in the pypi package repository.
func (p *PipLookup) isAvailableInPublic(pkgname string) bool {
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
