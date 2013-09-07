package main

import (
	"code.google.com/p/go-imap/go1/imap"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Derived from the example in the imap documentation.
//
// Note: most of error handling code is omitted for brevity
//

func main() {
	// Options
	server := flag.String("server", "imap.gmail.com", "Server to check")
	user := flag.String("user", "gohm2013@gmail.com", "IMAP user")
	password := flag.String("password", "", "Password")
	mailbox := flag.String("mailbox", "inbox", "IMAP mailbox to incorporate")
	flag.Parse()

	// Logging
	logf, err := os.OpenFile("inc.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err.Error())
		log.Println("Logging to stderr instead.")
	} else {
		log.SetOutput(logf)
	}

	// Connect to the server
	log.Println("inc -server", *server, "-user", *user, "-mailbox", *mailbox)

	c, err := authenticatedClient(*server, *user, *password)
	if err != nil {
		panic(err.Error())
	}
	defer c.Logout(30 * time.Second)

	incorporate(c, *mailbox)
}

func authenticatedClient(server, user, password string) (*imap.Client, error) {
	c, err := imap.DialTLS(server, &tls.Config{})
	if err != nil {
		return nil, err
	}

	// Server greeting
	for _, response := range c.Data {
		log.Println("unilateral:", response)
	}
	c.Data = nil

	// Optionally enable encryption
	if c.Caps["STARTTLS"] {
		c.StartTLS(nil)
	}

	// Authenticate
	if c.State() == imap.Login {
		cmd, err := c.Login(user, password)
		if err != nil {
			panic(err.Error())
		}
		// Generally, there don't seem to be any responses.
		for _, response := range cmd.Data {
			log.Println("Login response:", response)
		}
	}

	// Check for new unilateral server data responses.
	for _, rsp := range c.Data {
		log.Println("unilateral:", rsp)
	}
	c.Data = nil
	return c, nil
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
		log.Println("Select response:", response)
		if len(response.Fields) >= 2 && response.Fields[0] == "UIDVALIDITY" {
			uidValidity = response.Fields[1].(uint32)
		}
	}

	err = ensureUIDValidity(mailbox, uidValidity)
	if err != nil {
		panic(err.Error())
	}

	// Fetch headers
	set, _ := imap.NewSeqSet("")
	set.Add("1:*")
	cmd, _ = c.Fetch(set, "RFC822", "INTERNALDATE", "UID")

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			info := rsp.MessageInfo()
			uid := fmt.Sprint(info.UID)
			fmt.Print(uid, " ")
			fn := filepath.Join(mailbox, uid)
			err := ioutil.WriteFile(fn, imap.AsBytes(info.Attrs["RFC822"]), 0666)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			} else {
				_ = os.Chtimes(fn, info.InternalDate, info.InternalDate)
			}
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp := range c.Data {
			log.Println("unilateral:", rsp)
		}
		c.Data = nil
	}
	fmt.Println()

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
