package main

import (
	"errors"
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
	err := next()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	showCurrentMessage()
	return 0
}

func next() error {
	dir, err := gohm.CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err.Error())
	}
	var max = 0
	for _, f := range entries {
		var i int
		n, _ := fmt.Sscan(f.Name(), &i)
		if n > 0 {
			if i > max {
				max = i
			}
		}
	}
	current := gohm.CurrentMessage()
	var currentNum int
	n, _ := fmt.Sscan(current, &currentNum)
	if n < 1 {
		return errors.New("Bad Current")
	}
	for currentNum < max {
		currentNum += 1
		fp, _ := os.Open(filepath.Join(dir, fmt.Sprint(currentNum)))
		if fp != nil {
			fp.Close()
			gohm.SetCurrentMessage(currentNum)
			return nil
		}
	}
	return errors.New("No next message")
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
