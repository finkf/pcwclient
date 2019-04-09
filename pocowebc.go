package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
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
	correctCommand.AddCommand(&correctLineCommand)
	correctCommand.AddCommand(&correctWordCommand)

	mainCommand.PersistentFlags().BoolVarP(&jsonOutput, "json", "J", false,
		"output raw json")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "D", false,
		"enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U", url(),
		"set pocoweb url (env: POCWEBC_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatString, "format", "F", "",
		"set output format")
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
	data   []interface{}
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

func (cmd *command) add(x interface{}) {
	if cmd.err != nil {
		return
	}
	cmd.data = append(cmd.data, x)
}

func (cmd *command) do(f func() error) {
	if cmd.err != nil {
		return
	}
	cmd.err = f()
}

func (cmd *command) print() error {
	if cmd.err != nil {
		return cmd.err
	}
	if jsonOutput {
		return cmd.printJSON()
	}
	if formatString != "" {
		return cmd.printTemplate(formatString)
	}
	return cmd.printWithIDs(cmd.data)
}

func (cmd *command) printJSON() error {
	var data interface{} = cmd.data
	if len(cmd.data) == 1 {
		data = cmd.data[0]
	}
	if err := json.NewEncoder(cmd.out).Encode(data); err != nil {
		return fmt.Errorf("error encoding to json: %v", err)
	}
	return nil
}

func (cmd *command) printTemplate(tmpl string) error {
	var data interface{} = cmd.data
	if len(cmd.data) == 1 {
		data = cmd.data[0]
	}
	t, err := template.New("pocwebc").Parse(
		strings.Replace(tmpl, "\\n", "\n", -1))
	if err != nil {
		return fmt.Errorf("invalid format string: %v", err)
	}
	if err = t.Execute(cmd.out, data); err != nil {
		return fmt.Errorf("error formatting string: %v", err)
	}
	return nil
}

func (cmd *command) printWithIDs(what interface{}) error {
	switch t := what.(type) {
	case *api.Page:
		return cmd.printPage(t)
	case *api.Line:
		return cmd.printLine(t)
	case *api.Token:
		cmd.err = cmd.printWord(t)
	case []api.Token:
		return cmd.printWords(t)
	case []interface{}:
		return cmd.printArray(t)
	case api.Session:
		return cmd.printSession(t)
	case api.Version:
		return cmd.printVersion(t)
	case *api.Book:
		return cmd.printBook(t)
	case *api.Books:
		return cmd.printBooks(t)
	case api.User:
		return cmd.printUser(t)
	case api.Users:
		return cmd.printUsers(t)
	case *api.SearchResults:
		return cmd.printSearchResults(t)
	}
	panic("invalid type to print")
}

func (cmd *command) printPage(page *api.Page) error {
	for _, line := range page.Lines {
		if err := cmd.printLine(&line); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) printLine(line *api.Line) error {
	if !printWords {
		_, err := fmt.Fprintf(cmd.out, "%d:%d:%d %s\n",
			line.ProjectID, line.PageID, line.LineID, line.Cor)
		return err
	}
	words, err := cmd.client.GetTokens(line.ProjectID, line.PageID, line.LineID)
	if err != nil {
		return err
	}
	return cmd.printWords(words.Tokens)
}

func (cmd *command) printWords(words []api.Token) error {
	for _, word := range words {
		if err := cmd.printWord(&word); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) printWord(word *api.Token) error {
	_, err := fmt.Fprintf(cmd.out, "%d:%d:%d:%d %s\n",
		word.ProjectID, word.PageID, word.LineID, word.TokenID, word.Cor)
	return err
}

func (cmd *command) printArray(xs []interface{}) error {
	for _, x := range xs {
		if err := cmd.printWithIDs(x); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) printSession(s api.Session) error {
	t := time.Unix(s.Expires, 0).Format(time.RFC3339)
	return info(cmd.out, "%d\t%s\t%s\t%s\t%s\n",
		s.User.ID, s.User.Email, s.Auth, t, s.User.Name)
}

func (cmd *command) printVersion(v api.Version) error {
	_, err := fmt.Fprintln(cmd.out, v.Version)
	return err
}

func (cmd *command) printSearchResults(res *api.SearchResults) error {
	for _, match := range res.Matches {
		for _, token := range match.Tokens {
			if err := printColoredSearchMatch(cmd.out, match.Line, token); err != nil {
				return err
			}
		}
	}
	return nil
}

func printColoredSearchMatch(out io.Writer, line api.Line, token api.Token) error {
	c := color.New(color.FgRed)
	o := strings.Index(line.Cor[token.Offset:], token.Cor) + token.Offset
	e := o + len(token.Cor)
	prefix, match, suffix := line.Cor[0:o], line.Cor[o:e], line.Cor[e:]
	_, err := fmt.Fprintf(out, "%d:%d:%d:%d %s",
		line.ProjectID, line.PageID, line.LineID, token.TokenID, prefix)
	if err != nil {
		return err
	}
	_, err = c.Fprint(out, match)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, suffix)
	return err
}

func (cmd *command) printUsers(users api.Users) error {
	for _, user := range users.Users {
		if err := cmd.printUser(user); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) printUser(user api.User) error {
	return info(cmd.out, "%d\t%s\t%s\t%s\t%t\n",
		user.ID, user.Name, user.Email, user.Institute, user.Admin)
}

func (cmd *command) printBooks(books *api.Books) error {
	for _, book := range books.Books {
		if err := cmd.printBook(&book); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *command) printBook(book *api.Book) error {
	return info(cmd.out, "%d\t%s\t%s\t%s\t%d\t%s\t%s\t%t\n",
		book.ProjectID, book.Author, book.Title, book.Description,
		book.Year, book.Language, book.ProfilerURL, book.IsBook)
}

func info(out io.Writer, format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "\t", " ", -1)
	_, err := fmt.Fprint(out, str)
	return err
}
