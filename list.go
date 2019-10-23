package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var (
	histPatterns    bool
	listCharsFilter string
)

func init() {
	listPatternsCommand.Flags().BoolVarP(&histPatterns, "hist", "H", false,
		"list historical rewrite patterns")
	listCharsCommand.Flags().StringVarP(&listCharsFilter,
		"filter", "f", "A-Za-z0-9", "set filter characters")

}

var listCommand = cobra.Command{
	Use:   "list",
	Short: "List various informations",
}

var listUsersCommand = cobra.Command{
	Use:   "users",
	Short: "List user information",
	RunE:  doListUsers,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listAllUsers(os.Stdout)
	}
	return listUsers(os.Stdout, args...)
}

func listUsers(out io.Writer, ids ...string) error {
	c := newClient(out)
	for _, id := range ids {
		var uid int
		if n := parseIDs(id, &uid); n != 1 {
			return fmt.Errorf("invalid user id: %q", id)
		}
		c.do(func(client *api.Client) (interface{}, error) {
			return client.GetUser(int64(uid))
		})
	}
	return c.done()
}

func listAllUsers(out io.Writer) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.GetUsers()
	})
	return c.done()
}

var listBooksCommand = cobra.Command{
	Use:   "books [ID [IDS...]]",
	Short: "List book information",
	RunE:  doListBooks,
}

func doListBooks(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listAllBooks(os.Stdout)
	}
	return listBooks(os.Stdout, args...)
}

func listBooks(out io.Writer, ids ...string) error {
	c := newClient(out)
	for _, id := range ids {
		var bid int
		if n := parseIDs(id, &bid); n != 1 {
			return fmt.Errorf("invalid book id: %q", id)
		}
		c.do(func(client *api.Client) (interface{}, error) {
			return client.GetBook(bid)
		})
	}
	return c.done()
}

func listAllBooks(out io.Writer) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.GetBooks()
	})
	return c.done()
}

var listPatternsCommand = cobra.Command{
	Use:   "patterns ID [QUERY [QUERY...]]",
	Short: "List patterns for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListPatterns,
}

func doListPatterns(cmd *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	switch len(args) {
	case 1:
		return listAllPatterns(os.Stdout, bid)
	case 2:
		return listPatterns(os.Stdout, bid, args[1])
	default:
		return listPatterns(os.Stdout, bid, args[1], args[2:]...)
	}
}

func listAllPatterns(out io.Writer, id int) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetPatterns(id, !histPatterns)
	})
	return c.done()
}

func listPatterns(out io.Writer, id int, q string, qs ...string) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.QueryPatterns(id, !histPatterns, q, qs...)
	})
	return c.done()
}

var listSuggestionsCommand = cobra.Command{
	Use:   "suggestions ID [QUERY [QUERY...]]",
	Short: "List profiler suggestions for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListSuggestions,
}

func doListSuggestions(cmd *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	u := unescape(args[1:]...)
	switch len(u) {
	case 0:
		return listAllSuggestions(os.Stdout, bid)
	case 1:
		return listSuggestions(os.Stdout, bid, u[0])
	default:
		return listSuggestions(os.Stdout, bid, u[0], u[1:]...)
	}
}

func listAllSuggestions(out io.Writer, id int) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetProfile(id)
	})
	return c.done()
}

func listSuggestions(out io.Writer, id int, q string, qs ...string) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.QueryProfile(id, q, qs...)
	})
	return c.done()
}

var listSuspiciousCommand = cobra.Command{
	Use:   "suspicious ID",
	Short: "List suspicous words for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListSuspicious,
}

func doListSuspicious(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetSuspicious(bid)
	})
	return c.done()
}

var listAdaptiveCommand = cobra.Command{
	Use:   "adaptive ID",
	Short: "List adaptive tokens for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListAdaptive,
}

func doListAdaptive(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetAdaptiveTokens(bid)
	})
	return c.done()
}

var listELCommand = cobra.Command{
	Use:   "el ID",
	Short: "List extended lexicon tokens for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListEL,
}

func doListEL(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetExtendedLexicon(bid)
	})
	return c.done()
}

var listRRDMCommand = cobra.Command{
	Use:   "rrdm ID",
	Short: "List post correction for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListRRDM,
}

func doListRRDM(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetPostCorrection(bid)
	})
	return c.done()
}

var listOCRModelsCommand = cobra.Command{
	Use:   "ocr ID",
	Short: "List available ocr models",
	Args:  exactlyNIDs(1),
	RunE:  doListOCRModels,
}

func doListOCRModels(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetOCRModels(bid)
	})
	return c.done()
}

var listCharsCommand = cobra.Command{
	Use:   "chars ID",
	Short: "List frequency list of characters in book ID",
	Args:  exactlyNIDs(1),
	RunE:  doListChars,
}

func doListChars(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book id: %q", args[0])
	}
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.GetCharMap(bid, charFilter())
	})
	return c.done()
}

func charFilter() string {
	var str strings.Builder
	wstr := []rune(listCharsFilter)
	for i := 0; i < len(wstr); {
		if i+1 < len(wstr) && i+2 < len(wstr) && wstr[i+1] == '-' {
			if wstr[i] < wstr[i+2] {
				for r := wstr[i]; r <= wstr[i+2]; r++ {
					str.WriteRune(r)
				}
				i += 3
				continue
			}
			for j := 0; j < 3; j++ {
				str.WriteRune(wstr[i+j])
			}
			i += 3
			continue
		}
		str.WriteRune(wstr[i])
		i++
	}
	return str.String()
}
