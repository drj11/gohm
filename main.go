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

	// List all top-level mailboxes.
	cmd, _ = imap.Wait(c.List("", "%"))

	fmt.Println("\nTop-level mailboxes:")
	for _, rsp = range cmd.Data {
		fmt.Println("|--", rsp.MailboxInfo())
	}

	// Check for new unilateral server data responses
	for _, rsp = range c.Data {
		fmt.Println("Unilateral:", rsp)
	}
	c.Data = nil

	// Check INBOX.
	c.Select("INBOX", true)
	fmt.Print("\nMailbox status:\n", c.Mailbox)

	// Fetch the headers of the 10 most recent messages
	set, _ := imap.NewSeqSet("")
	if c.Mailbox.Messages >= 10 {
		set.AddRange(c.Mailbox.Messages-9, c.Mailbox.Messages)
	} else {
		set.Add("1:*")
	}
	cmd, _ = c.Fetch(set, "RFC822.HEADER")

	// Process responses while the command is running
	fmt.Println("\nMost recent messages:")
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process command data
		for _, rsp = range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				fmt.Println("|--", msg.Header.Get("Subject"))
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
