/*
Copyright 2016 The gta AUTHORS. All rights reserved.

Use of this source code is governed by the Apache 2 license that can be found
in the LICENSE file.
*/

// Command gta uses git to find the subset of code changes from a branch
// and then builds the list of go packages that have changed as a result,
// including all dependent go packages.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/digitalocean/gta"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	flagBase := flag.String("base", "origin/master", "base, branch to diff against")
	flagMerge := flag.Bool("merge", false, "diff using the latest merge commit")
	flagChangedFiles := flag.String("changed-files", "", "path to a file containing a newline separated list of files that have changed")
	flagInclude := flag.String("include", "", "define changes to be filtered with a set of comma separated prefixes")
	flagTags := flag.String("tags", "", "a list of build tags to consider")
	// TODO(nan) need to figure out if the go/packages equivalent is skipping packages with no go files.
	//flagBuildableOnly := flag.Bool("buildable-only", true, "keep buildable changed packages only")
	flagJSON := flag.Bool("json", false, "output list of changes as json")
	flag.Parse()

	if *flagMerge && len(*flagChangedFiles) > 0 {
		fmt.Fprintln(os.Stderr, "changed files must not be provided when using the latest merge commit")
		os.Exit(1)
	}

	options := []gta.Option{
		gta.SetPrefixes(parseStringSlice(*flagInclude)...),
		gta.SetTags(parseStringSlice(*flagTags)...),
	}

	if len(*flagChangedFiles) == 0 {
		// override the differ to use the git differ instead.
		gitDifferOptions := []gta.GitDifferOption{
			gta.SetBaseBranch(*flagBase),
			gta.SetUseMergeCommit(*flagMerge),
		}
		options = append(options, gta.SetDiffer(gta.NewGitDiffer(gitDifferOptions...)))
	} else {
		sl, err := changedFiles(*flagChangedFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not read changed file list: %s", err)
			os.Exit(1)
		}
		options = append(options, gta.SetDiffer(gta.NewFileDiffer(sl)))
	}

	gt, err := gta.New(options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't prepare gta: %s", err)
		os.Exit(1)
	}

	packages, err := gt.ChangedPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't list dirty packages: %s", err)
		os.Exit(1)
	}

	if *flagJSON {
		err = json.NewEncoder(os.Stdout).Encode(packages)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	packagePaths := gta.UniquePackagePaths(packages.AllChanges)

	if terminal.IsTerminal(syscall.Stdin) {
		for _, pkg := range packagePaths {
			fmt.Println(pkg)
		}
		return
	}
	fmt.Println(strings.Join(packagePaths, " "))
}

func changedFiles(fn string) ([]string, error) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	sl := strings.Split(string(b), "\n")
	n := 0
	for _, s := range sl {
		if !keepChangedFile(s) {
			continue
		}

		if !filepath.IsAbs(s) {
			return nil, errors.New("all changed files paths must be absolute paths")
		}

		sl[n] = s
		n++
	}

	return sl[:n], nil
}

func keepChangedFile(s string) bool {
	// Trim spaces, especially in case the newlines were CRLF instead of LF.
	s = strings.TrimSpace(s)
	return len(s) > 0
}

func parseStringSlice(strValue string) []string {
	var values []string
	for _, s := range strings.Split(strValue, ",") {
		v := strings.TrimSpace(s)
		if v == "" {
			continue
		}
		values = append(values, v)
	}
	return values
}
