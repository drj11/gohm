package gohm

import (
	"errors"
	"fmt"
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
	l := filepath.Join(PATH, "gohm.log")
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

// The directory on the filesystem for the specified folder.
func FolderDir(folder string) string {
	return filepath.Join(PATH, folder)
}

// The directory on the filesystem for the current folder.
func CurrentFolderDir() (string, error) {
	folder, err := CurrentFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(PATH, folder), nil
}

// The current message number.
func CurrentMessage() (current string) {
	current = "1"
	dir, err := CurrentFolderDir()
	if err != nil {
		return
	}
	file, err := os.Open(filepath.Join(dir, ".cur"))
	if err != nil {
		return
	}
	defer file.Close()
	var cur int
	_, err = fmt.Fscan(file, &cur)
	if err != nil {
		return
	}
	return fmt.Sprint(cur)
}

func SetCurrentMessage(c int) error {
	dir, err := CurrentFolderDir()
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(filepath.Join(dir, ".cur"),
		os.O_WRONLY|os.O_CREATE, 0666)
	if fp != nil {
		defer fp.Close()
		fmt.Fprint(fp, c)
	}
	return err
}
