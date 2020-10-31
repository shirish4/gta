/*
Copyright 2016 The gta AUTHORS. All rights reserved.

Use of this source code is governed by the Apache 2 license that can be found
in the LICENSE file.
*/
package gta

import (
	"encoding/json"
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

var _ Differ = &testDiffer{}

type testDiffer struct {
	diff map[string]Directory
}

func (t *testDiffer) Diff() (map[string]Directory, error) {
	return t.diff, nil
}

func (t *testDiffer) DiffFiles() (map[string]bool, error) {
	panic("not implemented")
}

// func TestGTA(t *testing.T) {
// 	// A depends on B depends on C
// 	// dirC is dirty, we expect them all to be marked
// 	difr := &testDiffer{
// 		diff: map[string]Directory{
// 			"dirC": Directory{Exists: true},
// 		},
// 	}

// 	graph := &Graph{
// 		graph: map[string]map[string]bool{
// 			"C": map[string]bool{
// 				"B": true,
// 			},
// 			"B": map[string]bool{
// 				"A": true,
// 			},
// 		},
// 	}

// 	want := []*build.Package{
// 		&build.Package{ImportPath: "A"},
// 		&build.Package{ImportPath: "B"},
// 		&build.Package{ImportPath: "C"},
// 	}

// 	gta, err := New(SetDiffer(difr))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pkgs, err := gta.ChangedPackages()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got := pkgs.AllChanges

// 	if !reflect.DeepEqual(want, got) {
// 		t.Errorf("want: %v", want)
// 		t.Errorf(" got: %v", got)
// 		t.Fatal("expected want and got to be equal")
// 	}
// }

// func TestGTA_ChangedPackages(t *testing.T) {
// 	// A depends on B depends on C
// 	// D depends on B
// 	// E depends on F depends on G

// 	difr := &testDiffer{
// 		diff: map[string]Directory{
// 			"dirC": Directory{Exists: true},
// 			"dirH": Directory{Exists: true},
// 		},
// 	}

// 	graph := &Graph{
// 		graph: map[string]map[string]bool{
// 			"C": map[string]bool{
// 				"B": true,
// 			},
// 			"B": map[string]bool{
// 				"A": true,
// 				"D": true,
// 			},
// 			"G": map[string]bool{
// 				"F": true,
// 			},
// 			"F": map[string]bool{
// 				"E": true,
// 			},
// 		},
// 	}

// 	want := &Packages{
// 		Dependencies: map[string][]*build.Package{
// 			"C": []*build.Package{
// 				{ImportPath: "A"},
// 				{ImportPath: "B"},
// 				{ImportPath: "D"},
// 			},
// 			"G": []*build.Package{
// 				{ImportPath: "E"},
// 				{ImportPath: "F"},
// 			},
// 		},
// 		Changes: []*build.Package{
// 			{ImportPath: "C"},
// 			{ImportPath: "G"},
// 		},
// 		AllChanges: []*build.Package{
// 			{ImportPath: "A"},
// 			{ImportPath: "B"},
// 			{ImportPath: "C"},
// 			{ImportPath: "D"},
// 			{ImportPath: "E"},
// 			{ImportPath: "F"},
// 			{ImportPath: "G"},
// 		},
// 	}

// 	gta, err := New(SetDiffer(difr), SetPackager(pkgr))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got, err := gta.ChangedPackages()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := cmp.Diff(got, want); diff != "" {
// 		t.Errorf("(-want, +got)\n%s", diff)
// 	}
// }

// func TestGTA_Prefix(t *testing.T) {
// 	// A depends on B and foo
// 	// B depends on C and bar
// 	// C depends on qux
// 	difr := &testDiffer{
// 		diff: map[string]Directory{
// 			"dirB":   Directory{Exists: true},
// 			"dirC":   Directory{Exists: true},
// 			"dirFoo": Directory{Exists: true},
// 		},
// 	}

// 	graph := &Graph{
// 		graph: map[string]map[string]bool{
// 			"C": map[string]bool{
// 				"B": true,
// 			},
// 			"B": map[string]bool{
// 				"A": true,
// 			},
// 			"foo": map[string]bool{
// 				"A": true,
// 			},
// 			"bar": map[string]bool{
// 				"B": true,
// 			},
// 			"qux": map[string]bool{
// 				"C": true,
// 			},
// 		},
// 	}

// 	pkgr := &testPackager{
// 		dirs2Imports: map[string]string{
// 			"dirA":   "A",
// 			"dirB":   "B",
// 			"dirC":   "C",
// 			"dirFoo": "foo",
// 			"dirBar": "bar",
// 			"dirQux": "qux",
// 		},
// 		graph: graph,
// 		errs:  make(map[string]error),
// 	}
// 	want := []*build.Package{
// 		&build.Package{ImportPath: "C"},
// 		&build.Package{ImportPath: "foo"},
// 	}

// 	gta, err := New(SetDiffer(difr), SetPackager(pkgr), SetPrefixes("foo", "C"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pkgs, err := gta.ChangedPackages()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got := pkgs.AllChanges

// 	if !reflect.DeepEqual(want, got) {
// 		t.Errorf("want: %+v", want)
// 		t.Errorf(" got: %+v", got)
// 		t.Fatal("expected want and got to be equal")
// 	}
// }

// func TestNoBuildableGoFiles(t *testing.T) {
// 	// we have changes but they don't belong to any dirty golang files, so no dirty packages
// 	const dir = "docs"
// 	difr := &testDiffer{
// 		diff: map[string]Directory{
// 			dir: Directory{},
// 		},
// 	}

// 	pkgr := &testPackager{
// 		errs: map[string]error{
// 			dir: &build.NoGoError{
// 				Dir: dir,
// 			},
// 		},
// 	}

// 	var want []*build.Package

// 	gta, err := New(SetDiffer(difr))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pkgs, err := gta.ChangedPackages()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got := pkgs.AllChanges

// 	if !reflect.DeepEqual(want, got) {
// 		t.Errorf("want: %#v", want)
// 		t.Errorf(" got: %#v", got)
// 		t.Fatal("expected want and got to be equal")
// 	}
// }

// func TestSpecialCaseDirectory(t *testing.T) {
// 	// We want to ignore the special case directory "testdata"
// 	const (
// 		special1 = "specia/case/testdata"
// 		special2 = "specia/case/testdata/multi"
// 	)
// 	difr := &testDiffer{
// 		diff: map[string]Directory{
// 			special1: Directory{Exists: true}, // this
// 			special2: Directory{Exists: true},
// 			"dirC":   Directory{Exists: true},
// 		},
// 	}
// 	graph := &Graph{
// 		graph: map[string]map[string]bool{
// 			"C": map[string]bool{
// 				"B": true,
// 			},
// 			"B": map[string]bool{
// 				"A": true,
// 			},
// 		},
// 	}

// 	want := []*build.Package{
// 		&build.Package{ImportPath: "A"},
// 		&build.Package{ImportPath: "B"},
// 		&build.Package{ImportPath: "C"},
// 	}

// 	gta, err := New(SetDiffer(difr))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pkgs, err := gta.ChangedPackages()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got := pkgs.AllChanges

// 	if !reflect.DeepEqual(want, got) {
// 		t.Errorf("want: %v", want)
// 		t.Errorf(" got: %v", got)
// 		t.Fatal("expected want and got to be equal")
// 	}
// }

func TestUnmarshalJSON(t *testing.T) {
	want := &Packages{
		Dependencies: map[string][]*Package{
			"do/tools/build/gta": []*Package{
				{
					Package: &packages.Package{
						PkgPath: "do/tools/build/gta/cmd/gta",
					},
				},
				{
					Package: &packages.Package{
						PkgPath: "do/tools/build/gtartifacts",
					},
				},
			},
		},
		Changes: []*Package{
			{
				Package: &packages.Package{
					PkgPath: "do/teams/compute/octopus",
				},
			},
		},
		AllChanges: []*Package{
			{
				Package: &packages.Package{
					PkgPath: "do/teams/compute/octopus",
				},
			},
		},
	}
	in := []byte(`{"dependencies":{"do/tools/build/gta":["do/tools/build/gta/cmd/gta","do/tools/build/gtartifacts"]},"changes":["do/teams/compute/octopus"],"all_changes":["do/teams/compute/octopus"]}`)

	got := new(Packages)
	err := json.Unmarshal(in, got)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestJSONRoundtrip(t *testing.T) {
	want := &Packages{
		Dependencies: map[string][]*Package{
			"do/tools/build/gta": []*Package{
				{
					Package: &packages.Package{
						PkgPath: "do/tools/build/gta/cmd/gta",
					},
				},
				{
					Package: &packages.Package{
						PkgPath: "do/tools/build/gtartifacts",
					},
				},
			},
		},
		Changes: []*Package{
			{
				Package: &packages.Package{
					PkgPath: "do/teams/compute/octopus",
				},
			},
		},
		AllChanges: []*Package{
			{
				Package: &packages.Package{
					PkgPath: "do/teams/compute/octopus",
				},
			},
		},
	}

	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	got := new(Packages)
	err = json.Unmarshal(b, got)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}
