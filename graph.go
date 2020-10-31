/*
Copyright 2016 The gta AUTHORS. All rights reserved.

Use of this source code is governed by the Apache 2 license that can be found
in the LICENSE file.
*/
package gta

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Package represents a package and is a node in a dependency graph.
type Package struct {
	*packages.Package

	Dependencies map[string]*Package
	Dependents   map[string]*Package
}

// DependencyGraph represents a set of packages that are related to each other.
type DependencyGraph struct {
	// packageMap maps an ID to the package it uniquely identifies.
	packageMap map[string]*Package
	// packagePathMap maps a package path to the packages for that path.
	packagePathMap map[string][]*Package
	// fileMap maps a file path to a package that the file belongs to.
	// Note that not all file types are considered to be part of a package.
	fileMap map[string]*Package
	// dirMap maps a directory path to a map of id to package. This is used as a
	// fallback when identifying packages for files that are not considered to be
	// part of a package.
	dirMap map[string]map[string]*Package
}

// BuildDependencyGraph constructs a dependency graph.
func BuildDependencyGraph(includePkgs, tags []string) (*DependencyGraph, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedModule,
		BuildFlags: []string{
			fmt.Sprintf(`-tags=%s`, strings.Join(tags, ",")),
		},
		Tests: true,
	}

	pkgs := make([]string, len(includePkgs))
	for i, pkg := range includePkgs {
		pkgs[i] = fmt.Sprintf("%s...", pkg)
	}

	loadedPackages, err := packages.Load(cfg, pkgs...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	dg := &DependencyGraph{
		packageMap:     make(map[string]*Package),
		packagePathMap: make(map[string][]*Package),
		fileMap:        make(map[string]*Package),
		dirMap:         make(map[string]map[string]*Package),
	}

	var addPackage func(pkg *packages.Package) (*Package, error)
	addPackage = func(pkg *packages.Package) (*Package, error) {
		if existingPackage, exists := dg.packageMap[pkg.ID]; exists {
			return existingPackage, nil
		}

		// if len(pkg.Errors) > 0 {
		// 	errs := make([]string, len(pkg.Errors))
		// 	for i, err := range pkg.Errors {
		// 		fmt.Printf("%#v\n", err.Kind)
		// 		errs[i] = err.Error()
		// 	}
		// 	return nil, fmt.Errorf("loading package %s: %s", pkg.ID, strings.Join(errs, ", "))
		// }

		newPackage := &Package{
			Package:      pkg,
			Dependencies: make(map[string]*Package),
			Dependents:   make(map[string]*Package),
		}

		dg.packageMap[pkg.ID] = newPackage
		dg.packagePathMap[pkg.PkgPath] = append(dg.packagePathMap[pkg.PkgPath], newPackage)
		for _, files := range [][]string{
			pkg.GoFiles,
			pkg.CompiledGoFiles,
			pkg.IgnoredFiles,
			pkg.OtherFiles,
		} {
			for _, f := range files {
				dg.fileMap[f] = newPackage
				dm, exists := dg.dirMap[filepath.Dir(f)]
				if !exists {
					dm = make(map[string]*Package)
					dg.dirMap[filepath.Dir(f)] = dm
				}
				dm[newPackage.ID] = newPackage
			}
		}

		for _, importedPkg := range pkg.Imports {
			dependency, err := addPackage(importedPkg)
			if err != nil {
				return nil, err
			}

			dependency.Dependents[pkg.ID] = newPackage
			newPackage.Dependencies[importedPkg.ID] = dependency
		}

		return newPackage, nil
	}

	for _, pkg := range loadedPackages {
		_, err := addPackage(pkg)
		if err != nil {
			return nil, err
		}

	}
	return dg, nil
}

// ErrUnknownPackage is returned for an unknown package path in a dependency graph.
var ErrUnknownPackage = errors.New("unknown package in dependency graph")

// Dependencies returns the direct dependencies of the package.
func (g *DependencyGraph) Dependencies(pkgPath string) ([]*Package, error) {
	deps, err := g.dependencies(pkgPath, false)
	if err != nil {
		return nil, err
	}
	return packageMapToArray(deps), nil
}

// Dependencies returns the full transitive dependencies of the package.
func (g *DependencyGraph) TransitiveDependencies(pkgPath string) ([]*Package, error) {
	deps, err := g.dependencies(pkgPath, true)
	if err != nil {
		return nil, err
	}
	return packageMapToArray(deps), nil
}

// Dependencies returns the direct dependents of the package.
func (g *DependencyGraph) Dependents(pkgPath string) ([]*Package, error) {
	deps, err := g.dependents(pkgPath, false)
	if err != nil {
		return nil, err
	}
	return packageMapToArray(deps), nil
}

// Dependencies returns the full transitive dependents of the package.
func (g *DependencyGraph) TransitiveDependents(pkgPath string) ([]*Package, error) {
	deps, err := g.dependents(pkgPath, true)
	if err != nil {
		return nil, err
	}
	return packageMapToArray(deps), nil
}

// AffectedPackages returns the full transitive set of dependent packages
// affected by the the provides files.
func (g *DependencyGraph) AffectedPackages(files ...string) (direct []*Package, transitive []*Package, err error) {
	directlyAffected := make(map[string]*Package)
	for _, f := range files {
		pkg, exists := g.fileMap[f]
		if exists {
			directlyAffected[pkg.ID] = pkg
			continue
		}
		// File is not directly part of a package, attempt to match to the nearest
		// package based on directory.
		dirPkgs, exists := g.dirMap[filepath.Dir(f)]
		if exists {
			for id, pkg := range dirPkgs {
				directlyAffected[id] = pkg
			}
		}
	}

	allAffected := make(map[string]*Package)
	for id, pkg := range directlyAffected {
		direct = append(direct, pkg)
		allAffected[id] = pkg

		deps, err := g.dependents(pkg.PkgPath, true)
		if err != nil {
			return nil, nil, err
		}

		for id, pkg := range deps {
			allAffected[id] = pkg
		}
	}
	return direct, packageMapToArray(allAffected), nil
}

func (g *DependencyGraph) dependencies(pkgPath string, recursive bool) (map[string]*Package, error) {
	pkgs, exists := g.packagePathMap[pkgPath]
	if !exists {
		return nil, ErrUnknownPackage
	}

	dependencies := make(map[string]*Package)
	for _, pkg := range pkgs {
		for _, depPkg := range pkg.Dependencies {
			dependencies[depPkg.ID] = depPkg

			if !recursive {
				continue
			}

			deps, err := g.dependencies(depPkg.PkgPath, true)
			if err != nil {
				return nil, err
			}
			for id, depPkg := range deps {
				dependencies[id] = depPkg
			}
		}
	}
	return dependencies, nil
}

func (g *DependencyGraph) dependents(pkgPath string, recursive bool) (map[string]*Package, error) {
	pkgs, exists := g.packagePathMap[pkgPath]
	if !exists {
		return nil, ErrUnknownPackage
	}

	dependents := make(map[string]*Package)
	for _, pkg := range pkgs {
		for _, depPkg := range pkg.Dependents {
			dependents[depPkg.ID] = depPkg

			if !recursive {
				continue
			}

			deps, err := g.dependents(depPkg.PkgPath, true)
			if err != nil {
				return nil, err
			}
			for id, depPkg := range deps {
				dependents[id] = depPkg
			}
		}
	}
	return dependents, nil
}

func packageMapToArray(pkgMap map[string]*Package) []*Package {
	pkgs := make([]*Package, 0, len(pkgMap))
	for _, pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}
