package main

import (
	"fmt"
	"io"
	"os"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().BoolVarP(&searchErrorPattern, "error-pattern", "e", false,
		"search for error patterns")
}

var loginCommand = cobra.Command{
	Use:   "login EMAIL PASSWORD",
	Short: "login to pocoweb",
	RunE:  runLogin,
	Args:  cobra.ExactArgs(2),
}

func runLogin(cmd *cobra.Command, args []string) error {
	user, password := args[0], args[1]
	return login(os.Stdout, user, password)
}

func login(out io.Writer, user, password string) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	client, err := api.Login(url(), user, password)
	cmd := command{client: client, err: err, out: out}
	if err == nil {
		cmd.add(client.Session)
	}
	return cmd.print()
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
		cmd.add(version)
		return err
	})
	return cmd.print()
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
		cmd.add(res)
		return err
	})
	return cmd.print()
}
