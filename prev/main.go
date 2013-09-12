package main

import (
	"errors"
	"fmt"
	"github.com/drj11/gohm"
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
	err := prev()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	gohm.ShowCurrentMessage()
	return 0
}

func prev() error {
	dir, err := gohm.CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	current := gohm.CurrentMessage()
	var currentNum int
	n, _ := fmt.Sscan(current, &currentNum)
	if n < 1 {
		return errors.New("Bad Current")
	}
	for currentNum > 0 {
		currentNum -= 1
		fp, _ := os.Open(filepath.Join(dir, fmt.Sprint(currentNum)))
		if fp != nil {
			fp.Close()
			gohm.SetCurrentMessage(currentNum)
			return nil
		}
	}
	return errors.New("No prev message")
}
