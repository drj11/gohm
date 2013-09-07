package main

import (
	"bytes"
	"code.google.com/p/go-imap/go1/imap"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path/filepath"
	"time"
)

// Derived from the example in the imap documentation.
//
// Note: most of error handling code is omitted for brevity
//
var (
	cmd *imap.Command
	rsp *imap.Response
)

var server = flag.String("server", "imap.gmail.com", "Server to check")
var user = flag.String("user", "gohm2013@gmail.com", "IMAP user")
var password = flag.String("password", "", "Password")

func main() {
	flag.Parse()
	// Connect to the server
	c, err := imap.DialTLS(*server, &tls.Config{})
	if err != nil {
		panic(err.Error())
	}

	// Remember to log out and close the connection when finished
	defer c.Logout(30 * time.Second)

	// Server greeting
	for _, thing := range c.Data {
		fmt.Println("hello:", thing)
	}
	c.Data = nil

	// Optionally enable encryption
	if c.Caps["STARTTLS"] {
		c.StartTLS(nil)
	}

	// Authenticate
	if c.State() == imap.Login {
		cmd, err = c.Login(*user, *password)
		if err != nil {
			panic(err.Error())
		}
	}

	// Check for new unilateral server data responses
	for _, rsp = range c.Data {
		fmt.Println("Unilateral:", rsp)
	}
	c.Data = nil

	mailbox := "inbox"
	incorporate(c, mailbox)
}

func incorporate(c *imap.Client, mailbox string) {
	cmd, err := c.Select(mailbox, true)
	if err != nil {
		panic(err.Error())
	}

	// Collect the UIDVALIDITY response. Protocol error
	// if not sent by server (RFC 3501 6.3.1).
	var uidValidity uint32
	for _, response := range cmd.Data {
		fmt.Println("response:", response)
		if len(response.Fields) >= 2 && response.Fields[0] == "UIDVALIDITY" {
			uidValidity = response.Fields[1].(uint32)
		}
	}
	fmt.Println("uidvalidity", uidValidity)

	err = ensureUIDValidity(mailbox, uidValidity)
	if err != nil {
		panic(err.Error())
	}

	// Fetch headers
	set, _ := imap.NewSeqSet("")
	set.Add("1:*")
	cmd, _ = c.Fetch(set, "RFC822.HEADER", "UID")

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process command data
		for _, rsp = range cmd.Data {
			info := rsp.MessageInfo()
			header := imap.AsBytes(info.Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				fmt.Println(info.UID, msg.Header.Get("Subject"))
			}
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp = range c.Data {
			fmt.Println("Unilateral:", rsp)
		}
		c.Data = nil
	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}
}

// Checks the uidVality (obtained from server) against the stored
// version (on the filesystem in the file ".uidvalidity"). If
// the stored validity does not match, the directory is emptied.
func ensureUIDValidity(mailbox string, uidValidity uint32) error {
	shortName := ".uidvalidity"
	longName := filepath.Join(mailbox, shortName)
	b, err := ioutil.ReadFile(longName)
	if err != nil {
		// Probably "no such file or directory", but even
		// if it's some other error we declare UIDValidity.
		return freshValidMailbox(mailbox, uidValidity)
	}
	fmt.Println("ensureUIDValidity", string(b))
	var cachedValidity uint32
	fmt.Sscan(string(b), &cachedValidity)
	if err != nil {
		panic(err.Error())
	}
	if cachedValidity == uidValidity {
		return nil
	}
	temp, err := ioutil.TempDir(".", mailbox)
	if err != nil {
		return err
	}
	err = os.Rename(mailbox, temp)
	if err != nil {
		return err
	}
	os.RemoveAll(temp)
	return freshValidMailbox(mailbox, uidValidity)
}

func freshValidMailbox(mailbox string, uidValidity uint32) error {
	longName := filepath.Join(mailbox, ".uidvalidity")
	_ = os.MkdirAll(mailbox, 0777)
	s := fmt.Sprint(uidValidity)
	err := ioutil.WriteFile(longName, []byte(s), 0666)
	if err != nil {
		return err
	}
	return nil
}
