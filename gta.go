/*
Copyright 2016 The gta AUTHORS. All rights reserved.

Use of this source code is governed by the Apache 2 license that can be found
in the LICENSE file.
*/
package gta

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	// ErrNoDiffer is returned when there is no differ set on the GTA.
	ErrNoDiffer = errors.New("there is no differ set")
)

// Packages contains various detailed information about the structure of
// packages GTA has detected.
type Packages struct {
	// Dependencies contains a map of changed packages to their dependencies
	Dependencies map[string][]*Package

	// Changes represents the changed files
	Changes []*Package

	// AllChanges represents all packages that are dirty including the initial
	// changed packages.
	AllChanges []*Package
}

type packagesJSON struct {
	Dependencies map[string][]string `json:"dependencies,omitempty"`
	Changes      []string            `json:"changes,omitempty"`
	AllChanges   []string            `json:"all_changes,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface.
func (p *Packages) MarshalJSON() ([]byte, error) {
	s := packagesJSON{
		Dependencies: mapify(p.Dependencies),
		Changes:      UniquePackagePaths(p.Changes),
		AllChanges:   UniquePackagePaths(p.AllChanges),
	}
	return json.Marshal(s)
}

// UnmarshalJSON used by gtartifacts when providing a changed package list
// see `useChangedPackagesFrom()`
func (p *Packages) UnmarshalJSON(b []byte) error {
	s := new(packagesJSON)

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	p.Dependencies = make(map[string][]*Package)
	for k, v := range s.Dependencies {
		for _, vv := range v {
			p.Dependencies[k] = append(p.Dependencies[k], &Package{Package: &packages.Package{PkgPath: vv}})
		}
	}

	for _, v := range s.Changes {
		p.Changes = append(p.Changes, &Package{Package: &packages.Package{PkgPath: v}})
	}

	for _, v := range s.AllChanges {
		p.AllChanges = append(p.AllChanges, &Package{Package: &packages.Package{PkgPath: v}})
	}

	return nil
}

// A GTA provides a method of building dirty packages, and their dependent
// packages.
type GTA struct {
	prefixes        []string
	tags            []string
	differ          Differ
	dependencyGraph *DependencyGraph
}

// New returns a new GTA with various options passed to New. Options will be
// applied in order so that later options can override earlier options.
func New(opts ...Option) (*GTA, error) {
	gta := &GTA{
		differ: NewGitDiffer(),
	}

	for _, opt := range opts {
		err := opt(gta)
		if err != nil {
			return nil, err
		}
	}

	var err error
	gta.dependencyGraph, err = BuildDependencyGraph(gta.prefixes, gta.tags)
	if err != nil {
		return nil, err
	}
	return gta, nil
}

// ChangedPackages uses the differ and packager to build a map of changed root
// packages to their dependent packages where dependent is defined as "changed"
// as well due to their dependency to the changed packages. It returns the
// dependency graph, the changes differ detected and a set of all unique
// packages (including the changes).
//
// As an example: package "foo" is imported by packages "bar" and "qux". If
// "foo" has changed, it has two dependent packages, "bar" and "qux". The
// result would be then:
//
//   Dependencies = {"foo": ["bar", "qux"]}
//   Changes      = ["foo"]
//   AllChanges   = ["foo", "bar", "qux]
//
// Note that two different changed package might have the same dependent
// package. Below you see that both "foo" and "foo2" has changed. Each have
// "bar" because "bar" imports both "foo" and "foo2", i.e:
//
//   Dependencies = {"foo": ["bar", "qux"], "foo2" : ["afa", "bar", "qux"]}
//   Changes      = ["foo", "foo2"]
//   AllChanges   = ["foo", "foo2", "afa", "bar", "qux]
func (g *GTA) ChangedPackages() (*Packages, error) {
	diffDirs, err := g.differ.Diff()
	if err != nil {
		return nil, fmt.Errorf("determining diff: %w", err)
	}

	var diffFilePaths []string
	for path, d := range diffDirs {
		for _, file := range d.Files {
			diffFilePaths = append(diffFilePaths, fmt.Sprintf("%s/%s", path, file))
		}
	}

	direct, transitive, err := g.dependencyGraph.AffectedPackages(diffFilePaths...)
	if err != nil {
		return nil, fmt.Errorf("determining affected packages: %w", err)
	}

	cp := &Packages{
		Dependencies: make(map[string][]*Package),
		Changes:      direct,
		AllChanges:   transitive,
	}
	for _, pkg := range direct {
		deps, err := g.dependencyGraph.TransitiveDependents(pkg.PkgPath)
		if err != nil {
			return nil, fmt.Errorf("building dependency map: %w", err)
		}
		cp.Dependencies[pkg.PkgPath] = deps
	}
	return cp, nil
}

type byPackageImportPath []*Package

func (b byPackageImportPath) Len() int               { return len(b) }
func (b byPackageImportPath) Less(i int, j int) bool { return b[i].PkgPath < b[j].PkgPath }
func (b byPackageImportPath) Swap(i int, j int)      { b[i], b[j] = b[j], b[i] }

// UniquePackagePaths returns the set of unique package paths for a set of
// packages.
func UniquePackagePaths(pkgs []*Package) []string {
	pkgPathMap := make(map[string]struct{})
	for _, pkg := range pkgs {
		// _test packages and .test binaries should be treated like the actual
		// package.
		if strings.HasSuffix(pkg.PkgPath, "_test") {
			pkgPathMap[strings.TrimSuffix(pkg.PkgPath, "_test")] = struct{}{}
		} else if strings.HasSuffix(pkg.PkgPath, ".test") {
			pkgPathMap[strings.TrimSuffix(pkg.PkgPath, ".test")] = struct{}{}
		} else {
			pkgPathMap[pkg.PkgPath] = struct{}{}
		}
	}

	pkgPaths := make([]string, 0, len(pkgPathMap))
	for path := range pkgPathMap {
		pkgPaths = append(pkgPaths, path)
	}

	sort.Slice(pkgPaths, func(i, j int) bool {
		return pkgPaths[i] < pkgPaths[j]
	})
	return pkgPaths
}

func mapify(pkgs map[string][]*Package) map[string][]string {
	out := map[string][]string{}
	for key, pkgs := range pkgs {
		out[key] = UniquePackagePaths(pkgs)
	}
	return out
}
