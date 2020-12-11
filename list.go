package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/finkf/gofiler"
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
	Short: "list various informations",
}

var listUsersCommand = cobra.Command{
	Use:   "users [IDs...]",
	Short: "list user information",
	RunE:  doListUsers,
	Long: `
List user information for the users with the given IDs.  If no IDs are
given, information about all users is listed.`,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	if len(args) == 0 {
		return listAllUsers(c)
	}
	return listUsers(c, args...)
}

func listUsers(c *api.Client, ids ...string) error {
	for _, id := range ids {
		var uid int
		if n := parseIDs(id, &uid); n != 1 {
			return fmt.Errorf("list user: invalid user id: %q", id)
		}
		var user api.User
		if err := get(c, c.URL("users/%d", uid), &user); err != nil {
			return fmt.Errorf("list user %d: %v", uid, err)
		}
		format(&user)
	}
	return nil
}

func listAllUsers(c *api.Client) error {
	var users api.Users
	if err := get(c, c.URL("users"), &users); err != nil {
		return fmt.Errorf("list users: %v", err)
	}
	format(&users)
	return nil
}

var listBooksCommand = cobra.Command{
	Use:   "books [IDs...]",
	Short: "list book information",
	RunE:  doListBooks,
	Long: `
List book information for the books and/or packages with the given
IDs.  If no IDs are given, information about all available books
and/or packages is listed.`,
}

func doListBooks(cmd *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	if len(args) == 0 {
		return listAllBooks(c)
	}
	return listBooks(c, args...)
}

func listBooks(c *api.Client, ids ...string) error {
	for _, id := range ids {
		var bid int
		if n := parseIDs(id, &bid); n != 1 {
			return fmt.Errorf("list book: invalid book id: %q", id)
		}
		var book api.Book
		if err := get(c, c.URL("books/%d", bid), &book); err != nil {
			return fmt.Errorf("list book %d: %v", bid, err)
		}
		format(&book)
	}
	return nil
}

func listAllBooks(c *api.Client) error {
	var books api.Books
	if err := get(c, c.URL("books"), &books); err != nil {
		return fmt.Errorf("list books: %v", err)
	}
	format(&books)
	return nil
}

var listPatternsCommand = cobra.Command{
	Use:   "patterns ID [QUERY...]",
	Short: "list patterns for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListPatterns,
}

func doListPatterns(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("list patterns: invalid book id: %q", args[0])
	}
	u := unescape(args...)
	switch len(u) {
	case 1:
		return listPatterns(c, bid)
	default:
		return listPatterns(c, bid, u[1:]...)
	}
}

func listPatterns(c *api.Client, id int, qs ...string) error {
	uri := c.URL("profile/patterns/books/%d?ocr=%t", id, !histPatterns)
	for _, q := range qs {
		uri += "&q=" + url.QueryEscape(q)
	}
	var counts api.PatternCounts
	if err := get(c, uri, &counts); err != nil {
		return fmt.Errorf("list patterns for book %d: %v", id, err)
	}
	format(&counts)
	return nil
}

var listSuggestionsCommand = cobra.Command{
	Use:   "suggestions ID [QUERY...]",
	Short: "list profiler suggestions for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListSuggestions,
}

func doListSuggestions(cmd *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("list suggestions: invalid book id: %q", args[0])
	}
	u := unescape(args...)
	switch len(u) {
	case 1:
		return listSuggestions(c, bid)
	default:
		return listSuggestions(c, bid, u[1:]...)
	}
}

func listSuggestions(c *api.Client, id int, qs ...string) error {
	uri := c.URL("profile/books/%d", id)
	pre := "?"
	for _, q := range qs {
		uri += pre + "q=" + url.QueryEscape(q)
		pre = "&"
	}
	if len(qs) == 0 {
		var profile gofiler.Profile
		if err := get(c, uri, &profile); err != nil {
			return fmt.Errorf("list suggestions for book %d: %v", id, err)
		}
		format(profile)
		return nil
	}
	var suggs api.Suggestions
	if err := get(c, uri, &suggs); err != nil {
		return fmt.Errorf("list suggestions for book %d: %v", id, err)
	}
	format(suggs)
	return nil

}

var listSuspiciousCommand = cobra.Command{
	Use:   "suspicious ID [IDs...]",
	Short: "list suspicous words for the given books",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListSuspicious,
}

func doListSuspicious(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for i := range args {
		var bid int
		if n := parseIDs(args[i], &bid); n != 1 {
			return fmt.Errorf("list suspicious: invalid book id: %q", args[i])
		}
		url := c.URL("profile/suspicious/books/%d", bid)
		var counts api.SuggestionCounts
		if err := get(c, url, &counts); err != nil {
			return fmt.Errorf("list suspicious for %d: %v", bid, err)
		}
		format(&counts)
	}
	return nil
}

var listAdaptiveCommand = cobra.Command{
	Use:   "adaptive ID [IDs...]",
	Short: "list adaptive tokens for the given books",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListAdaptive,
}

func doListAdaptive(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for i := range args {
		var bid int
		if n := parseIDs(args[i], &bid); n != 1 {
			return fmt.Errorf("list adaptive tokens: invalid book id: %q", args[i])
		}
		url := c.URL("profile/adaptive/books/%d", bid)
		var tokens api.AdaptiveTokens
		if err := get(c, url, &tokens); err != nil {
			return fmt.Errorf("list adaptive tokens for book %d: %v", bid, err)
		}
		format(&tokens)
	}
	return nil
}

var listELCommand = cobra.Command{
	Use:   "el [ID...]",
	Short: "list extended lexicon tokens for the given books",
	RunE:  doListEL,
}

func doListEL(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for i := range args {
		var bid int
		if n := parseIDs(args[i], &bid); n != 1 {
			return fmt.Errorf("list extended lexicon entries: invalid book id: %q", args[i])
		}
		url := c.URL("postcorrect/le/books/%d", bid)
		var el api.ExtendedLexicon
		if err := get(c, url, &el); err != nil {
			return fmt.Errorf("list extended lexicon entries for book %d: %v", bid, err)
		}
		format(&el)
	}
	return nil
}

var listRRDMCommand = cobra.Command{
	Use:   "rrdm [ID...]",
	Short: "list post correction for the given book",
	RunE:  doListRRDM,
}

func doListRRDM(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for i := range args {
		var bid int
		if n := parseIDs(args[i], &bid); n != 1 {
			return fmt.Errorf("list post corrections: invalid book id: %q", args[i])
		}
		url := c.URL("postcorrect/books/%d", bid)
		var pc api.PostCorrection
		if err := get(c, url, &pc); err != nil {
			return fmt.Errorf("list post corrections for book %d: %v", bid, err)
		}
		format(&pc)
	}
	return nil
}

var listCharsCommand = cobra.Command{
	Use:   "chars ID",
	Short: "list frequency list of characters in book ID",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListChars,
}

func doListChars(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for i := range args {
		var bid int
		if n := parseIDs(args[i], &bid); n != 1 {
			return fmt.Errorf("list chars: invalid book id: %q", args[i])
		}
		url := c.URL("books/%d/charmap?filter=%s",
			bid, url.QueryEscape(charFilter()))
		var chars api.CharMap
		if err := get(c, url, &chars); err != nil {
			return fmt.Errorf("list chars for book %d: %v", bid, err)
		}
		format(&chars)
	}
	return nil
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
