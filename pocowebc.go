package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// various command line flags
var (
	debug        = false
	jsonOutput   = false
	formatString = ""
	authToken    = ""
	pocowebURL   = ""
	configpath   = ""
	noconfig     = false
)

var mainCommand = &cobra.Command{
	Use:   "pocowebc",
	Short: "command line client for pocoweb",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	mainCommand.AddCommand(&listCommand)
	mainCommand.AddCommand(&createCommand)
	mainCommand.AddCommand(&loginCommand)
	mainCommand.AddCommand(&logoutCommand)
	mainCommand.AddCommand(&printCommand)
	mainCommand.AddCommand(&versionCommand)
	mainCommand.AddCommand(&rawCommand)
	mainCommand.AddCommand(&searchCommand)
	mainCommand.AddCommand(&correctCommand)
	mainCommand.AddCommand(&downloadCommand)
	mainCommand.AddCommand(&splitCommand)
	mainCommand.AddCommand(&assignCommand)
	mainCommand.AddCommand(&finishCommand)
	mainCommand.AddCommand(&deleteCommand)
	listCommand.AddCommand(&listUserCommand)
	listCommand.AddCommand(&listBookCommand)
	listCommand.AddCommand(&listUsersCommand)
	listCommand.AddCommand(&listBooksCommand)
	createCommand.AddCommand(&createUserCommand)
	createCommand.AddCommand(&createBookCommand)

	mainCommand.SilenceUsage = true
	mainCommand.SilenceErrors = true
	mainCommand.PersistentFlags().BoolVarP(&jsonOutput, "json", "J", false,
		"output raw json")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "D", false,
		"enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U",
		getURL(), "set pocoweb url (env: POCOWEBC_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatString, "format", "F",
		"", "set output format")
	mainCommand.PersistentFlags().StringVarP(&authToken, "auth", "A",
		getAuth(), "set auth token (env: POCOWEBC_AUTH)")
	config, _ := userConfigDir()
	configpath = filepath.Join(config, "pocowebc/config.toml")
	mainCommand.PersistentFlags().StringVarP(&configpath, "config", "C",
		configpath, "set auth token (env: POCOWEBC_CONFIG)")
	mainCommand.PersistentFlags().BoolVarP(&noconfig, "noconfig", "N", false,
		"do not use configuration file")
}

func exactlyNIDs(n int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("expected exactly %d arguments", n)
		}
		for _, arg := range args {
			_, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("not an ID: %s", arg)
			}
		}
		return nil
	}
}

func scanf(str, format string, args ...interface{}) error {
	_, err := fmt.Sscanf(str, format, args...)
	return err
}

func wordID(id string) (bid, pid, lid, wid int, ok bool) {
	if err := scanf(id, "%d:%d:%d:%d", &bid, &pid, &lid, &wid); err != nil {
		return 0, 0, 0, 0, false
	}
	return bid, pid, lid, wid, true
}

func lineID(id string) (bid, pid, lid int, ok bool) {
	if err := scanf(id, "%d:%d:%d", &bid, &pid, &lid); err != nil {
		return 0, 0, 0, false
	}
	return bid, pid, lid, true
}

func pageID(id string) (bid, pid int, ok bool) {
	if err := scanf(id, "%d:%d", &bid, &pid); err != nil {
		return 0, 0, false
	}
	return bid, pid, true
}

func bookID(id string) (bid int, ok bool) {
	if err := scanf(id, "%d", &bid); err != nil {
		return 0, false
	}
	return bid, true
}

func userConfigDir() (string, error) {
	config := os.Getenv("XDG_DATA_HOME")
	if config != "" {
		return config, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

type command struct {
	client *api.Client
	data   []interface{}
	out    io.Writer
	err    error
}

func newCommand(out io.Writer) command {
	config := loadConfig()
	if config.Auth == "" || config.URL == "" {
		return command{err: fmt.Errorf("missing login information: see login sub command")}
	}
	return command{client: api.Authenticate(config.URL, config.Auth), out: out}
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
	err := f()
	if err != nil {
		cmd.err = err
	}
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
	return cmd.info("%d\t%s\t%s\t%s\t%s\n",
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
	return cmd.info("%d\t%s\t%s\t%s\t%t\n",
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

func toStrings(xs []int) []string {
	res := make([]string, len(xs))
	for i, x := range xs {
		res[i] = fmt.Sprintf("%d", x)
	}
	return res
}

func (cmd *command) printBook(book *api.Book) error {
	return cmd.info("%d\t%s\t%s\t%s\t%d\t%s\t%s\t%t\t%s\n",
		book.ProjectID, book.Author, book.Title, book.Description,
		book.Year, book.Language, book.ProfilerURL, book.IsBook,
		strings.Join(toStrings(book.PageIDs), ","))
}

func (cmd *command) info(format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "\t", " ", -1)
	_, err := fmt.Fprint(cmd.out, str)
	return err
}
