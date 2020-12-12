package main

import (
	"fmt"
	"net/url"

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

var searchArgs = struct {
	skip int
	max  int
	typ  string
	all  bool
	ic   bool
}{}

var searchCommand = cobra.Command{
	Use:   "search ID [QUERIES...]",
	Short: "search for tokens and error patterns",
	RunE:  runSearch,
	Args:  cobra.MinimumNArgs(1),
}

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
			id, searchArgs.ic, searchArgs.max, skip,
			url.QueryEscape(searchArgs.typ))
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
