package main

import (
	"fmt"
	"github.com/drj11/gohm"
	"os"
)

var (
	GOHM_PATH, setupErr = gohm.Setup()
)

func main() {
	if setupErr != nil {
		fmt.Fprintln(os.Stderr, setupErr.Error())
		os.Exit(4)
	}
}
