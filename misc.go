package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().BoolVarP(&searchErrorPattern, "error-pattern", "e",
		false, "search for error patterns")
	splitCommand.Flags().BoolVarP(&splitRandom, "random", "r",
		false, "split random")
	splitCommand.Flags().IntVarP(&splitN, "number", "n",
		10, "number of splits")
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
	client, err := api.Login(host(), user, password)
	cmd := command{client: client, err: err, out: out}
	if err == nil {
		cmd.add(client.Session)
	}
	return cmd.print()
}

var logoutCommand = cobra.Command{
	Use:   "logout",
	Short: "logout from pocoweb",
	RunE:  runLogout,
	Args:  cobra.NoArgs, //ExactArgs(0),
}

func runLogout(cmd *cobra.Command, args []string) error {
	cmdx := newCommand(os.Stdout)
	cmdx.do(func() error {
		return cmdx.client.Logout()
	})
	return cmdx.err
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
	cmd := command{out: out, client: api.NewClient(host())}
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

var downloadCommand = cobra.Command{
	Use:   "download ID [OUTPUT-FILE]",
	Short: "login to pocoweb",
	RunE:  doDownload,
	Args:  cobra.RangeArgs(1, 2),
}

func doDownload(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return download(os.Stdout, args[0])
	}
	out, err := os.Create(args[1])
	if err != nil {
		return err
	}
	defer out.Close()
	return download(out, args[0])
}

func download(out io.Writer, id string) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		bid, ok := bookID(id)
		if !ok {
			return fmt.Errorf("invalid book id: %s", id)
		}
		r, err := cmd.client.Download(bid)
		if err != nil {
			return err
		}
		defer r.Close()
		n, err := io.Copy(cmd.out, r)
		log.Debugf("wrote %d bytes", n)
		return err
	})
	return cmd.err
}

var (
	splitRandom bool
	splitN      int
)

var splitCommand = cobra.Command{
	Use:   "split ID",
	Short: "split a book into multiple projects",
	RunE:  doSplit,
	Args:  cobra.ExactArgs(1),
}

func doSplit(cmd *cobra.Command, args []string) error {
	return split(os.Stdout, args[0])
}

func split(out io.Writer, id string) error {
	bid, ok := bookID(id)
	if !ok {
		return fmt.Errorf("invalid book ID: %s", id)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		books, err := cmd.client.Split(bid, splitN, splitRandom)
		cmd.add(books)
		return err
	})
	return cmd.print()
}

var assignCommand = cobra.Command{
	Use:   "assign BOOK-ID USER-ID",
	Short: "assign a book to another user",
	Args:  exactlyNIDs(2),
	RunE:  doAssign,
}

func doAssign(cmd *cobra.Command, args []string) error {
	bid, _ := strconv.Atoi(args[0])
	uid, _ := strconv.Atoi(args[1])
	return assign(os.Stdout, bid, uid)
}

func assign(out io.Writer, bid, uid int) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		return cmd.client.Assign(bid, uid)
	})
	return cmd.err
}

var finishCommand = cobra.Command{
	Use:   "finish ID",
	Short: "finish a book and reassign it to its original user",
	Args:  exactlyNIDs(1),
	RunE:  doFinish,
}

func doFinish(cmd *cobra.Command, args []string) error {
	bid, _ := strconv.Atoi(args[0])
	return finish(os.Stdout, bid)
}

func finish(out io.Writer, bid int) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		return cmd.client.Finish(bid)
	})
	return cmd.err
}

var deleteCommand = cobra.Command{
	Use:   "delete IDS...",
	Short: "delete a book, page or line",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doDelete,
}

func doDelete(_ *cobra.Command, args []string) error {
	cmd := newCommand(os.Stdout)
	for _, id := range args {
		if bid, pid, lid, ok := lineID(id); ok {
			cmd.do(func() error {
				return cmd.client.DeleteLine(bid, pid, lid)
			})
			continue
		}
		if bid, pid, ok := pageID(id); ok {
			cmd.do(func() error {
				return cmd.client.DeletePage(bid, pid)
			})
			continue
		}
		if _, ok := bookID(id); ok {
			return fmt.Errorf("cannot delete book: not implemented")
		}
		return fmt.Errorf("cannot delete: invalid id: %s", id)
	}
	return nil
}
