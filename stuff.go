package gohm

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

var (
    PATH = os.Getenv("GOHM_PATH")
)

func Setup() (string, error) {
	if PATH == "" {
		return "", errors.New("GOHM_PATH must be set")
	}
	_ = os.MkdirAll(PATH, 0777)

	// Logging
	l := filepath.Join(PATH, "inc.log")
	logf, err := os.OpenFile(l, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err.Error())
		log.Println("Logging to stderr instead.")
	} else {
		log.SetOutput(logf)
	}
	return PATH, nil
}

func CurrentFolder() (string, error) {
    return "inbox", nil
}
