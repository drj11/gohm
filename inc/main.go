package main

import (
	"code.google.com/p/go-imap/go1/imap"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/drj11/gohm"
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

var (
	GOHM_PATH, setupErr = gohm.Setup()
)

func main() {
	exitStatus := mainExitStatus()
	if exitStatus != 0 {
		os.Exit(exitStatus)
	}
}

func mainExitStatus() int {
	// Options
	server := flag.String("server", "imap.gmail.com", "Server to check")
	user := flag.String("user", "gohm2013@gmail.com", "IMAP user")
	password := flag.String("password", "", "Password")
	mailbox := flag.String("mailbox", "inbox", "IMAP mailbox to incorporate")
	flag.Parse()

	if setupErr != nil {
		fmt.Fprintln(os.Stderr, setupErr.Error())
		return 4
	}

	// Connect to the server
	log.Println("inc -server", *server, "-user", *user, "-mailbox", *mailbox)

	c, err := authenticatedClient(*server, *user, *password)
	if err != nil {
		panic(err.Error())
	}
	defer c.Logout(30 * time.Second)

	incorporate(c, *mailbox)
	return 0
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
	// :todo:(drj) Make mailbox the current folder.

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

	// Make list of UIDs already retrieved.
	got, _ := imap.NewSeqSet("")
	dir := gohm.FolderDir(mailbox)
	entries, _ := ioutil.ReadDir(dir)
	for _, f := range entries {
		var i_ int
		n, _ := fmt.Sscan(f.Name(), &i_)
		if n < 1 {
			continue
		}
		got.Add(f.Name())
	}
	log.Println("got", got)

	notgot, _ := imap.NewSeqSet("")
	// Fetch list of UIDs on server.
	all, _ := imap.NewSeqSet("")
	all.Add("1:*")
	cmd, _ = c.Fetch(all, "UID")
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			info := rsp.MessageInfo()
			if got.Contains(info.UID) {
				continue
			}
			notgot.AddNum(info.UID)
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp := range c.Data {
			log.Println("unilateral:", rsp)
		}
		c.Data = nil
	}
	log.Println("notgot", notgot)

	fetch(c, notgot, mailbox)
}

func fetch(c *imap.Client, set *imap.SeqSet, mailbox string) {
	if set.Empty() {
		return
	}
	// Fetch messages
	cmd, _ := c.Fetch(set, "RFC822", "INTERNALDATE", "UID")

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			info := rsp.MessageInfo()
			uid := fmt.Sprint(info.UID)
			fmt.Print(uid, " ")
			fn := filepath.Join(GOHM_PATH, mailbox, uid)
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
	longName := filepath.Join(GOHM_PATH, mailbox, shortName)
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
	dir := filepath.Join(GOHM_PATH, mailbox)
	v := filepath.Join(dir, ".uidvalidity")
	_ = os.MkdirAll(dir, 0777)
	s := fmt.Sprint(uidValidity)
	err := ioutil.WriteFile(v, []byte(s), 0666)
	if err != nil {
		return err
	}
	return nil
}
