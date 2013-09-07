package gohm

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func Setup() (string, error) {
	GOHM_PATH := os.Getenv("GOHM_PATH")

	if GOHM_PATH == "" {
		return "", errors.New("GOHM_PATH must be set")
	}
	_ = os.MkdirAll(GOHM_PATH, 0777)

	// Logging
	l := filepath.Join(GOHM_PATH, "inc.log")
	logf, err := os.OpenFile(l, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err.Error())
		log.Println("Logging to stderr instead.")
	} else {
		log.SetOutput(logf)
	}
	return GOHM_PATH, nil
}
