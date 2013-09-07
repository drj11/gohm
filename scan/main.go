package main

import (
	"fmt"
	"github.com/drj11/gohm"
	"io/ioutil"
	"os"
)

var (
	_, setupErr = gohm.Setup()
)

func main() {
	exitStatus := mainExitStatus()
	if exitStatus != 0 {
		os.Exit(exitStatus)
	}
}
func mainExitStatus() int {
	if setupErr != nil {
		fmt.Fprintln(os.Stderr, setupErr.Error())
		return 4
	}
	dir, err := gohm.CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	for _, f := range entries {
		var i int
		n, _ := fmt.Sscan(f.Name(), &i)
		if n == 1 {
			fmt.Printf("%4s\n", f.Name())
		}
	}
	return 0
}
