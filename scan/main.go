package main

import (
	"fmt"
	"github.com/drj11/gohm"
	"io/ioutil"
	"net/mail"
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
	dir, err := gohm.CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	current := gohm.CurrentMessage()
	for _, f := range entries {
		// Flags for current message and replied.
		// :todo:(drj) implement.
		var curr, repl string
		var i_ int
		n, _ := fmt.Sscan(f.Name(), &i_)
		if current == f.Name() {
			curr = "+"
		}
		if n == 1 {
			fullName := filepath.Join(dir, f.Name())
			fp, err := os.Open(fullName)
			if err != nil {
				panic(err.Error())
			}
			defer fp.Close()
			msg, _ := mail.ReadMessage(fp)
			var subj = "??"
			if msg != nil {
				subj = msg.Header.Get("Subject")
			}
			fmt.Printf("%4s%1s%1sMM-DD  %-40.40s\n", f.Name(), curr, repl, subj)
		}
	}
	return 0
}
