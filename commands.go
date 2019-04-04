package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().BoolVarP(&searchErrorPattern, "error-pattern", "e", false, "search for error patterns")
}

var loginCommand = cobra.Command{
	Use:   "login EMAIL PASSWORD",
	Short: "login to pocoweb",
	RunE:  runLogin,
	Args:  cobra.ExactArgs(2),
}

func runLogin(cmd *cobra.Command, args []string) error {
	user, password := args[0], args[1]
	return login(user, password, os.Stdout)
}

func login(user, password string, out io.Writer) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	client, err := api.Login(url(), user, password)
	cmd := command{client: client, err: err, out: out}
	if cmd.err == nil {
		cmd.data = cmd.client.Session
	}
	return cmd.output(func() error {
		s := cmd.data.(api.Session)
		t := time.Unix(s.Expires, 0).Format(time.RFC3339)
		_, err := fmt.Fprintf(out, "%d\t%s\t%s\t%s\n",
			s.User.ID, s.User.Email, s.Auth, t)
		return err
	})
}

var versionCommand = cobra.Command{
	Use:   "version",
	Short: "get version information",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	return version(os.Stdout)
}

func version(out io.Writer) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		version, err := cmd.client.GetAPIVersion()
		cmd.data = version
		return err
	})
	return cmd.output(func() error {
		_, err := fmt.Fprintln(out, cmd.data.(api.Version).Version)
		return err
	})
}

var rawCommand = cobra.Command{
	Use:   "raw FORMAT [ARGS...]",
	Short: "send raw get requests",
	RunE:  runRaw,
	Args:  cobra.MinimumNArgs(1),
}

func runRaw(cmd *cobra.Command, args []string) error {
	iargs := make([]interface{}, len(args)-1)
	for i := 1; i < len(args); i++ {
		iargs[i] = args[i]
	}
	return raw(os.Stdout, args[0], iargs...)
}

func raw(out io.Writer, format string, args ...interface{}) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		return cmd.client.Raw(fmt.Sprintf(format, args...), out)
	})
	return cmd.err
}

var searchCommand = cobra.Command{
	Use:   "search ID QUERY",
	Short: "search for tokens and error patterns",
	RunE:  runSearch,
	Args:  cobra.ExactArgs(2),
}

var searchErrorPattern bool

func runSearch(cmd *cobra.Command, args []string) error {
	iargs := make([]interface{}, len(args)-1)
	for i := 1; i < len(args); i++ {
		iargs[i-1] = args[i]
	}
	return search(os.Stdout, args[0], args[1], searchErrorPattern)
}

func search(out io.Writer, id, query string, errorPattern bool) error {
	var bid int
	if err := scanf(id, "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		res, err := cmd.client.Search(bid, query, errorPattern)
		cmd.data = res
		return err
	})
	return cmd.output(func() error {
		for _, match := range cmd.data.(*api.SearchResults).Matches {
			for _, token := range match.Tokens {
				if err := printSearchMatch(out, match.Line, token); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func printSearchMatch(out io.Writer, line api.Line, token api.Token) error {
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
