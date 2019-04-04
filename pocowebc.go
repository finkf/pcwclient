package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mainCommand = &cobra.Command{
		Use:   "pocowebc",
		Short: "command line client for pocoweb",
	}
	debug        = false
	jsonOutput   = false
	formatString = ""
	authToken    = ""
	pocowebURL   = ""
	userID       = 0
	bookID       = 0
	pageID       = 0
	lineID       = 0
	wordID       = 0
)

func init() {
	mainCommand.AddCommand(&listCommand)
	mainCommand.AddCommand(&createCommand)
	mainCommand.AddCommand(&loginCommand)
	mainCommand.AddCommand(&printCommand)
	mainCommand.AddCommand(&versionCommand)
	mainCommand.AddCommand(&rawCommand)
	mainCommand.AddCommand(&searchCommand)
	mainCommand.AddCommand(&correctCommand)
	listCommand.AddCommand(&listUserCommand)
	listCommand.AddCommand(&listBookCommand)
	createCommand.AddCommand(&createUserCommand)
	createCommand.AddCommand(&createBookCommand)
	printCommand.AddCommand(&printPageCommand)
	printCommand.AddCommand(&printBookCommand)
	printCommand.AddCommand(&printLineCommand)
	printCommand.AddCommand(&printWordCommand)
	correctCommand.AddCommand(&correctLineCommand)
	correctCommand.AddCommand(&correctWordCommand)

	mainCommand.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output raw json")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "", false, "enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U", url(),
		"set pocoweb url (env: POCOWEBC_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatString, "format", "f", "", "set output format")
	mainCommand.PersistentFlags().StringVarP(&authToken, "auth", "a", auth(),
		"set auth token (env: POCOWEBC_AUTH)")
	mainCommand.PersistentFlags().IntVarP(&userID, "user-id", "u", 0, "set user id")
	mainCommand.PersistentFlags().IntVarP(&bookID, "book-id", "b", 0, "set book id")
	mainCommand.PersistentFlags().IntVarP(&pageID, "page-id", "p", 0, "set page id")
	mainCommand.PersistentFlags().IntVarP(&lineID, "line-id", "l", 0, "set line id")
	mainCommand.PersistentFlags().IntVarP(&wordID, "word-id", "w", 0, "set word id")
}

func url() string {
	if pocowebURL != "" {
		return pocowebURL
	}
	return os.Getenv("POCOWEBC_URL")
}

func auth() string {
	if authToken != "" {
		return authToken
	}
	return os.Getenv("POCOWEBC_AUTH")
}

type command struct {
	client *api.Client
	data   interface{}
	out    io.Writer
	err    error
}

func newCommand(out io.Writer) command {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	auth := auth()
	if auth == "" {
		return command{err: fmt.Errorf("missing login information: " +
			"use --auth or set POCOWEBC_AUTH envinronment variable")}
	}
	return command{client: api.Authenticate(url(), auth), out: out}
}

func (cmd *command) output(f func() error) error {
	if cmd.err != nil {
		return cmd.err
	}
	if jsonOutput {
		if err := json.NewEncoder(cmd.out).Encode(cmd.data); err != nil {
			return fmt.Errorf("error encoding to json: %v", err)
		}
		return nil
	}
	if formatString != "" {
		t, err := template.New("pocwebc").Parse(
			strings.Replace(formatString, "\\n", "\n", -1))
		if err != nil {
			return fmt.Errorf("invalid format string: %v", err)
		}
		if err = t.Execute(cmd.out, cmd.data); err != nil {
			return fmt.Errorf("error formatting string: %v", err)
		}
		return nil
	}
	return f()
}

func (cmd *command) do(f func() error) {
	if cmd.err != nil {
		return
	}
	cmd.err = f()
}
