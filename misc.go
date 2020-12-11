package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func init() {
	searchCommand.Flags().StringVarP(&searchArgs.typ, "type", "t",
		"token", "set search type (token|pattern|ac|regex)")
	searchCommand.Flags().BoolVarP(&formatArgs.ocr, "ocr", "o", false,
		"print ocr lines")
	searchCommand.Flags().BoolVarP(&formatArgs.noCor, "nocor", "c", false,
		"do not print corrected lines")
	searchCommand.Flags().BoolVarP(&formatArgs.words, "words", "w",
		false, "print out matched words")
	searchCommand.Flags().BoolVarP(&searchArgs.all, "all", "a",
		false, "search for all matches")
	searchCommand.Flags().BoolVarP(&searchArgs.ic, "ignore-case", "i",
		false, "ignore case for search")
	searchCommand.Flags().IntVarP(&searchArgs.max, "max", "m",
		50, "set max matches")
	searchCommand.Flags().IntVarP(&searchArgs.skip, "skip", "s",
		0, "set skip matches")
}

var loginCommand = cobra.Command{
	Use:   "login [EMAIL PASSWORD]",
	Short: "login to pocoweb or get logged in user",
	RunE:  runLogin,
	Args:  exactArgs(0, 2),
}

func runLogin(cmd *cobra.Command, args []string) error {
	if len(args) == 2 {
		user, password := args[0], args[1]
		return login(user, password)
	}
	return getLogin()
}

func login(user, password string) error {
	// if mainArgs.debug {
	// 	log.SetLevel(log.DebugLevel)
	// }
	url := getURL()
	if url == "" {
		return fmt.Errorf("login: missing url: use --url or PCWCLIENT_URL")
	}
	c, err := api.Login(url, user, password, mainArgs.skipVerify)
	if err != nil {
		return fmt.Errorf("login: %v", err)
	}
	format(c.Session)
	return nil
}

func getLogin() error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var session api.Session
	if err := get(c, c.URL("get login"), &session); err != nil {
		return fmt.Errorf("get login: %v", err)
	}
	format(session)
	return nil
}

var logoutCommand = cobra.Command{
	Use:   "logout",
	Short: "logout from pocoweb",
	RunE:  runLogout,
	Args:  cobra.NoArgs,
}

func runLogout(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	if err := get(c, c.URL("login"), nil); err != nil {
		return fmt.Errorf("logout: %v", err)
	}
	return nil
}

var versionCommand = cobra.Command{
	Use:   "version",
	Short: "get version information",
	RunE:  runVersion,
}

func runVersion(_ *cobra.Command, args []string) error {
	url := getURL()
	if url == "" {
		return fmt.Errorf("missing url: use --url, or set POCOWEBC_URL")
	}
	var version api.Version
	c := api.NewClient(url, mainArgs.skipVerify)
	if err := get(c, c.URL("api-version"), &version); err != nil {
		return fmt.Errorf("get api version: %v", err)
	}
	return nil
}

var searchCommand = cobra.Command{
	Use:   "search ID [QUERIES...]",
	Short: "search for tokens and error patterns",
	RunE:  runSearch,
	Args:  cobra.MinimumNArgs(1),
}

var searchArgs = struct {
	skip int
	max  int
	typ  string
	all  bool
	ic   bool
}{}

func runSearch(_ *cobra.Command, args []string) error {
	var id int
	if n := parseIDs(args[0], &id); n != 1 {
		return fmt.Errorf("search: invalid book id: %q", args[0])
	}
	return search(id, args[1:]...)
}

func hasAnyMatches(res *api.SearchResults) bool {
	for _, m := range res.Matches {
		if len(m.Lines) > 0 {
			return true
		}
	}
	return false
}

func search(id int, qs ...string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	skip := searchArgs.skip
	for {
		uri := c.URL("books/%d/search?i=%t&max=%d&skip=%d&type=%s",
			id, searchArgs.ic, searchArgs.max, skip, url.QueryEscape(searchArgs.typ))
		for _, q := range qs {
			uri += "&q=" + url.QueryEscape(q)
		}
		var results api.SearchResults
		if err := get(c, uri, &results); err != nil {
			return fmt.Errorf("search book %d: %v", id, err)
		}
		if !hasAnyMatches(&results) {
			break
		}
		format(&results)
		if !searchArgs.all {
			break
		}
		skip += searchArgs.max
	}
	return nil
}

var downloadCommand = cobra.Command{
	Use:   "download ID [FILE]",
	Short: "dowload archive of book ID",
	RunE:  doDownload,
	Args:  cobra.RangeArgs(1, 2),
}

func doDownload(_ *cobra.Command, args []string) error {
	if len(args) == 1 {
		return download(os.Stdout, args[0])
	}
	out, err := os.Create(args[1])
	if err != nil {
		return fmt.Errorf("download to %s: %v", args[1], err)
	}
	defer out.Close()
	return download(out, args[0])
}

func download(out io.Writer, id string) error {
	var bid int
	if n := parseIDs(id, &bid); n != 1 {
		return fmt.Errorf("download: invalid book id: %s", id)
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var ar struct {
		Archive string `json:"archive"`
	}
	if err := get(c, c.URL("books/%d/download", bid), &ar); err != nil {
		return fmt.Errorf("download: %v", err)
	}
	url := strings.TrimRight(c.Host, "/") + "/" + strings.TrimLeft(ar.Archive, "/")
	if err := downloadZIP(c, url, out); err != nil {
		return fmt.Errorf("download: %v", err)
	}
	return nil
}
