package gohm

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
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

func ShowCurrentMessage() {
	dir, err := CurrentFolderDir()
	if err != nil {
		panic(err.Error())
	}
	current := CurrentMessage()
	fullName := filepath.Join(dir, current)
	show(fullName)
}
func show(path string) {
	r, err := os.Open(path)
	if err != nil {
		panic(err.Error())
	}
	msg, err := mail.ReadMessage(r)
	if err != nil {
		panic(err.Error())
	}
	contentType, isMime := msg.Header["Content-Type"]
	if isMime {
		mediaType, params, err := mime.ParseMediaType(contentType[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		if mediaType != "multipart/alternative" {
			fmt.Println("I'm refusing to handle", mediaType)
			return
		}
		showMultipart(msg, params)
	} else {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err.Error())
		}
		os.Stdout.Write(content)
	}
}

func showMultipart(message *mail.Message, params map[string]string) {
	boundary := params["boundary"]
	reader := multipart.NewReader(message.Body, boundary)
	partN := 0
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		partTypeHeader := part.Header["Content-Type"][0]
		log.Println("Part", partN, " type", partTypeHeader)
		partType, _, err := mime.ParseMediaType(partTypeHeader)
		if partType == "text/plain" {
			showTextPlain(message, part)
			return
		}
		partN += 1
	}
	fmt.Fprintln(os.Stderr, "Didn't find any text/plain part")
}

func showTextPlain(message *mail.Message, part *multipart.Part) {
	fmt.Fprintln(os.Stdout, "Subject:", message.Header["Subject"][0])
	fmt.Fprintln(os.Stdout, "Date:", message.Header["Date"][0])
	fmt.Fprintln(os.Stdout, "From:", message.Header["From"][0])
	fmt.Fprintln(os.Stdout)
	partBody, err := ioutil.ReadAll(part)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	os.Stdout.Write(partBody)
}
