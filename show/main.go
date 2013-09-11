package main

import (
	"fmt"
	"github.com/drj11/gohm"
	"io/ioutil"
	"os"
	"path/filepath"
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
        showCurrentMessage()
        return 0
}

func showCurrentMessage() {
	dir, err := gohm.CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	current := gohm.CurrentMessage()
        fullName := filepath.Join(dir, current)
        content, err := ioutil.ReadFile(fullName)
        if err != nil {
                panic(err.Error())
        }
        os.Stdout.Write(content)
}
