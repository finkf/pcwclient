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
	searchCommand.Flags().BoolVarP(&searchPattern, "pattern", "e",
		false, "search for error patterns")
	searchCommand.Flags().BoolVarP(&searchAll, "all", "a",
		false, "search for all matches")
	searchCommand.Flags().IntVarP(&searchMax, "max", "m",
		50, "set max matches")
	searchCommand.Flags().IntVarP(&searchSkip, "skip", "s",
		0, "set skip matches")
	splitCommand.Flags().BoolVarP(&splitRandom, "random", "r",
		false, "split random")
}

var loginCommand = cobra.Command{
	Use:   "login [EMAIL PASSWORD]",
	Short: "login to pocoweb or get logged in user",
	RunE:  runLogin,
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 0 && len(args) != 2 {
			return fmt.Errorf("accepts 0 or 2 arg(s)")
		}
		return nil
	},
}

func runLogin(cmd *cobra.Command, args []string) error {
	if len(args) == 2 {
		user, password := args[0], args[1]
		return login(os.Stdout, user, password)
	}
	return getLogin(os.Stdout)
}

func login(out io.Writer, user, password string) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	url := getURL()
	if url == "" {
		return fmt.Errorf("missing url: use --url, or POCOWEBC_URL")
	}
	client, err := api.Login(url, user, password)
	if err != nil {
		return fmt.Errorf("cannot login: %v", err)
	}
	cmd := command{client: client, err: err, out: out}
	session := client.Session
	cmd.do(func(client *api.Client) (interface{}, error) {
		return session, nil
	})
	return cmd.done()
}

func getLogin(out io.Writer) error {
	cmd := newCommand(out)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return client.GetLogin()
	})
	return cmd.done()
}

var logoutCommand = cobra.Command{
	Use:   "logout",
	Short: "logout from pocoweb",
	RunE:  runLogout,
	Args:  cobra.NoArgs, //ExactArgs(0),
}

func runLogout(_ *cobra.Command, args []string) error {
	cmd := newCommand(os.Stdout)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return nil, client.Logout()
	})
	return cmd.done()
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
	url := getURL()
	if url == "" {
		return fmt.Errorf("missing url: use --url, or set POCOWEBC_URL")
	}
	cmd := command{out: out, client: api.NewClient(url)}
	cmd.do(func(client *api.Client) (interface{}, error) {
		return cmd.client.GetAPIVersion()
	})
	return cmd.done()
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
	cmd.do(func(client *api.Client) (interface{}, error) {
		return nil, cmd.client.Raw(fmt.Sprintf(format, args...), out)
	})
	return cmd.done()
}

var searchCommand = cobra.Command{
	Use:   "search ID QUERY [QUERIES...]",
	Short: "search for tokens and error patterns",
	RunE:  runSearch,
	Args:  cobra.MinimumNArgs(2),
}

var (
	searchSkip    int
	searchMax     int
	searchPattern bool
	searchAll     bool
)

func runSearch(cmd *cobra.Command, args []string) error {
	var id int
	if n := parseIDs(args[0], &id); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	switch len(args) {
	case 2:
		return search(os.Stdout, id, searchPattern, args[1])
	default:
		return search(os.Stdout, id, searchPattern, args[1], args[2:]...)
	}
}

func search(out io.Writer, id int, ep bool, q string, qs ...string) error {
	cmd := newCommand(out)
	cmd.client.Skip = searchSkip
	cmd.client.Max = searchMax
	var done bool
	for !done {
		done = true
		cmd.do(func(client *api.Client) (interface{}, error) {
			done = false
			var res *api.SearchResults
			var err error
			if ep {
				res, err = cmd.client.SearchErrorPatterns(id, q, qs...)
			} else {
				res, err = cmd.client.Search(id, q, qs...)
			}
			if !searchAll || len(res.Matches) == 0 {
				done = true
				return nil, nil
			}
			cmd.client.Skip += cmd.client.Max
			return res, err
		})
	}
	return cmd.done()
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
	cmd.do(func(client *api.Client) (interface{}, error) {
		var bid int
		if n := parseIDs(id, &bid); n != 1 {
			return nil, fmt.Errorf("invalid book id: %s", id)
		}
		r, err := cmd.client.Download(bid)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		n, err := io.Copy(cmd.out, r)
		log.Debugf("wrote %d bytes", n)
		return nil, err
	})
	return cmd.err
}

var (
	splitRandom bool
)

var splitCommand = cobra.Command{
	Use:   "split ID USERID [USERID...]",
	Short: "split the project ID into multiple packages",
	RunE:  doSplit,
	Args:  cobra.MinimumNArgs(2),
}

func doSplit(cmd *cobra.Command, args []string) error {
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("split: invalid id: %s", arg)
		}
		ids = append(ids, id)
	}
	if err := split(os.Stdout, ids[0], ids[1:]); err != nil {
		return fmt.Errorf("split: %v", err)
	}
	return nil
}

func split(out io.Writer, bid int, userids []int) error {
	cmd := newCommand(out)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return cmd.client.Split(bid, splitRandom, userids[0], userids[1:]...)
	})
	return cmd.done()
}

var assignCommand = cobra.Command{
	Use:   "assign ID [USERID]",
	Short: "assign the package ID to the user USERID",
	RunE:  doAssign,
	Args: func(_ *cobra.Command, args []string) error {
		switch len(args) {
		case 1, 2:
			return nil
		default:
			return fmt.Errorf("requires one or two args")
		}
	},
}

func doAssign(cmd *cobra.Command, args []string) error {
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("assign: invalid id: %s", arg)
		}
		ids = append(ids, id)
	}
	var err error
	switch len(ids) {
	case 2:
		err = assignTo(os.Stdout, ids[0], ids[1])
	default:
		err = assignBack(os.Stdout, ids[0])
	}
	if err != nil {
		return fmt.Errorf("assign: %v", err)
	}
	return nil
}

func assignTo(out io.Writer, bid, uid int) error {
	cmd := newCommand(out)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return nil, cmd.client.AssignTo(bid, uid)
	})
	return cmd.done()
}

func assignBack(out io.Writer, bid int) error {
	cmd := newCommand(out)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return nil, cmd.client.AssignBack(bid)
	})
	return cmd.done()
}

var takeBackCommand = cobra.Command{
	Use:   "takeback ID",
	Short: "take back packages of project ID",
	RunE:  doTakeBack,
	Args:  cobra.ExactArgs(1),
}

func doTakeBack(cmd *cobra.Command, args []string) error {
	var pid int
	if n := parseIDs(args[0], &pid); n != 1 {
		return fmt.Errorf("takeback: invalid id: %s", args[0])
	}
	if err := takeBack(os.Stdout, pid); err != nil {
		return fmt.Errorf("takeback: %v", err)
	}
	return nil
}

func takeBack(out io.Writer, pid int) error {
	cmd := newCommand(out)
	cmd.do(func(client *api.Client) (interface{}, error) {
		return nil, cmd.client.TakeBack(pid)
	})
	return cmd.done()
}

var deleteCommand = cobra.Command{
	Use:   "delete",
	Short: "delete users or books",
}

var deleteBooksCommand = cobra.Command{
	Use:   "books IDS...",
	Short: "delete a books, pages or lines",
	Args:  cobra.MinimumNArgs(1),
	RunE:  deleteBooks,
}

func deleteBooks(_ *cobra.Command, args []string) error {
	cmd := newCommand(os.Stdout)
	for _, id := range args {
		var bid, pid, lid int
		switch n := parseIDs(id, &bid, &pid, &lid); n {
		case 3:
			cmd.do(func(client *api.Client) (interface{}, error) {
				return nil, client.DeleteLine(bid, pid, lid)
			})
			continue
		case 2:
			cmd.do(func(client *api.Client) (interface{}, error) {
				return nil, client.DeletePage(bid, pid)
			})
			continue
		case 1:
			cmd.do(func(client *api.Client) (interface{}, error) {
				return nil, client.DeleteBook(bid)
			})
			continue
		default:
			return fmt.Errorf("cannot delete: invalid id: %q", id)
		}
	}
	return nil
}

var deleteUsersCommand = cobra.Command{
	Use:   "users IDS...",
	Short: "delete users",
	Args:  cobra.MinimumNArgs(1),
	RunE:  deleteUsers,
}

func deleteUsers(_ *cobra.Command, args []string) error {
	cmd := newCommand(os.Stdout)
	for _, id := range args {
		var uid int
		if n := parseIDs(id, &uid); n != 1 {
			return fmt.Errorf("cannot delete user: invalid user id: %s", id)
		}
		cmd.do(func(client *api.Client) (interface{}, error) {
			return nil, client.DeleteUser(int64(uid))
		})
	}
	return nil
}
