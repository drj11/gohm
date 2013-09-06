package main

import (
	"bytes"
	"code.google.com/p/go-imap/go1/imap"
	"crypto/tls"
	"flag"
	"fmt"
	"net/mail"
	"time"
)

// Derived from the example in the imap documentation.
//
// Note: most of error handling code is omitted for brevity
//
var (
	c   *imap.Client
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

	// Check INBOX.
	_, err = c.Select("INBOX", true)
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
