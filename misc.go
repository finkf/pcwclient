package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().StringVarP(&searchType, "type", "t",
		"token", "set search type (token|pattern|ac)")
	searchCommand.Flags().BoolVarP(&formatOCR, "ocr", "o", false,
		"print ocr lines")
	searchCommand.Flags().BoolVarP(&noFormatCor, "nocor", "c", false,
		"do not print corrected lines")
	searchCommand.Flags().BoolVarP(&formatWords, "words", "w",
		false, "print out matched words")
	searchCommand.Flags().BoolVarP(&searchAll, "all", "a",
		false, "search for all matches")
	searchCommand.Flags().BoolVarP(&searchIC, "ignore-case", "i",
		false, "ignore case for search")
	searchCommand.Flags().IntVarP(&searchMax, "max", "m",
		50, "set max matches")
	searchCommand.Flags().IntVarP(&searchSkip, "skip", "s",
		0, "set skip matches")
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
		return fmt.Errorf("missing url: use --url or PCWCLIENT_URL")
	}
	login, err := api.Login(url, user, password, skipVerify)
	if err != nil {
		return fmt.Errorf("cannot login: %v", err)
	}
	c := client{client: login, err: err, out: out}
	c.do(func(client *api.Client) (interface{}, error) {
		return client.Session, nil
	})
	return c.done()
}

func getLogin(out io.Writer) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.GetLogin()
	})
	return c.done()
}

var logoutCommand = cobra.Command{
	Use:   "logout",
	Short: "logout from pocoweb",
	RunE:  runLogout,
	Args:  cobra.NoArgs, //ExactArgs(0),
}

func runLogout(_ *cobra.Command, args []string) error {
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return nil, client.Logout()
	})
	return c.done()
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
	c := client{out: out, client: api.NewClient(url, skipVerify)}
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetAPIVersion()
	})
	return c.done()
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
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return nil, c.client.Raw(fmt.Sprintf(format, args...), out)
	})
	return c.done()
}

var searchCommand = cobra.Command{
	Use:   "search ID [QUERIES...]",
	Short: "search for tokens and error patterns",
	RunE:  runSearch,
	Args:  cobra.MinimumNArgs(1),
}

var (
	searchSkip int
	searchMax  int
	searchType string
	searchAll  bool
	searchIC   bool
)

func searchTypeFromString(typ string) (api.SearchType, error) {
	switch strings.ToLower(typ) {
	case "token":
		return api.SearchToken, nil
	case "pattern":
		return api.SearchPattern, nil
	case "ac":
		return api.SearchAC, nil
	default:
		return "", fmt.Errorf("invalid search type: %s", typ)
	}
}

func runSearch(cmd *cobra.Command, args []string) error {
	var id int
	if n := parseIDs(args[0], &id); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	typ, err := searchTypeFromString(searchType)
	if err != nil {
		return err
	}
	return search(os.Stdout, id, typ, args[1:]...)
}

func search(out io.Writer, id int, typ api.SearchType, qs ...string) error {
	c := newClient(out)
	var done bool
	var f formatter
	defer f.done()
	for !done && c.err == nil {
		c.do(func(client *api.Client) (interface{}, error) {
			s := api.Search{
				Client: *client,
				Skip:   searchSkip,
				Max:    searchMax,
				IC:     searchIC,
				Type:   typ,
			}
			ret, err := s.Search(id, qs...)
			must(err, "cannot search: %v")
			f.format(ret)
			if len(ret.Matches) == 0 {
				done = true
				return nil, nil
			}
			if !searchAll {
				done = true
			}
			return nil, nil
		})
	}
	return c.done()
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
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		var bid int
		if n := parseIDs(id, &bid); n != 1 {
			return nil, fmt.Errorf("invalid book id: %s", id)
		}
		r, err := c.client.Download(bid)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		n, err := io.Copy(c.out, r)
		log.Debugf("wrote %d bytes", n)
		return nil, err
	})
	return c.err
}

var (
	splitRandom bool
)

var assignCommand = cobra.Command{
	Use:   "assign ID [USERID]",
	Short: "Assign the package ID to the user USERID",
	Long: `
Assign the package ID to a user.  If USERID is omitted, the package is
assigned back to its original owner.  Otherwise it is assigned to the
user with the given USERID.`,
	RunE: doAssign,
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
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return nil, c.client.AssignTo(bid, uid)
	})
	return c.done()
}

func assignBack(out io.Writer, bid int) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return nil, c.client.AssignBack(bid)
	})
	return c.done()
}

var pkgCommand = cobra.Command{
	Use:   "pkg",
	Short: "Assign and take back packages.",
}

var takeBackCommand = cobra.Command{
	Use:   "takeback ID",
	Short: "Take back packages of project ID",
	Long: `
Take back all packages of the project ID.  All packages of the project
that are owned by different users are reassigned to the owner of the
project.`,
	RunE: doTakeBack,
	Args: cobra.ExactArgs(1),
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
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return nil, c.client.TakeBack(pid)
	})
	return c.done()
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
	c := newClient(os.Stdout)
	for _, id := range args {
		var bid, pid, lid int
		switch n := parseIDs(id, &bid, &pid, &lid); n {
		case 3:
			c.do(func(client *api.Client) (interface{}, error) {
				return nil, client.DeleteLine(bid, pid, lid)
			})
			continue
		case 2:
			c.do(func(client *api.Client) (interface{}, error) {
				return nil, client.DeletePage(bid, pid)
			})
			continue
		case 1:
			c.do(func(client *api.Client) (interface{}, error) {
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
	c := newClient(os.Stdout)
	for _, id := range args {
		var uid int
		if n := parseIDs(id, &uid); n != 1 {
			return fmt.Errorf("cannot delete user: invalid user id: %s", id)
		}
		c.do(func(client *api.Client) (interface{}, error) {
			return nil, client.DeleteUser(int64(uid))
		})
	}
	return nil
}
