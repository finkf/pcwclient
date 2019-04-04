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
	listCommand.AddCommand(&listUsersCommand)
	listCommand.AddCommand(&listBooksCommand)
	createCommand.AddCommand(&createUserCommand)
	createCommand.AddCommand(&createBookCommand)
	printCommand.AddCommand(&printPageCommand)
	printCommand.AddCommand(&printBookCommand)
	printCommand.AddCommand(&printLineCommand)
	printCommand.AddCommand(&printWordCommand)
	correctCommand.AddCommand(&correctLineCommand)
	correctCommand.AddCommand(&correctWordCommand)

	mainCommand.PersistentFlags().BoolVarP(&jsonOutput, "json", "J", false, "output raw json")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U", url(),
		"set pocoweb url (env: POCWEBC_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatString, "format", "F", "", "set output format")
	mainCommand.PersistentFlags().StringVarP(&authToken, "auth", "A", auth(),
		"set auth token (env: POCWEBC_AUTH)")
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

func scanf(str, format string, args ...interface{}) error {
	_, err := fmt.Sscanf(str, format, args...)
	return err
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
			"use --auth or set POCOWEBC_AUTH environment variable")}
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
