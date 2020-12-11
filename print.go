package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func init() {
	printCommand.Flags().BoolVarP(&formatArgs.words, "words", "w", false,
		"print words not lines")
	printCommand.Flags().BoolVarP(&formatArgs.ocr, "ocr", "o", false,
		"print ocr lines")
	printCommand.Flags().BoolVarP(&formatArgs.noCor, "nocor", "c", false,
		"do not print corrected lines")
	printCommand.Flags().BoolVarP(&formatArgs.onlyManual, "manual", "m", false,
		"only print manual corrected lines/words")
}

var printCommand = cobra.Command{
	Use:   "print IDs...",
	Short: "Print books, pages, lines and/or words",
	RunE:  printIDs,
}

func printIDs(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for _, id := range args {
		if err := doPrintID(c, id); err != nil {
			return err
		}
	}
	if len(args) > 0 {
		return nil
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if err := doPrintID(c, s.Text()); err != nil {
			return err
		}
	}
	return s.Err()
}

func doPrintID(c *api.Client, id string) error {
	var bid, pid, lid, wid, len, mod int
	id, mod = getMod(id)
	switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
	case 5:
		return getWord(c, bid, pid, lid, wid, len)
	case 4:
		return getWord(c, bid, pid, lid, wid, -1)
	case 3:
		return getLine(c, bid, pid, lid)
	case 2:
		_, _, err := getPage(c, bid, pid, mod)
		return err
	case 1:
		return getPages(c, bid)
	default:
		return fmt.Errorf("invalid id: %s", id)
	}
}

func getPages(c *api.Client, bid int) error {
	pageid := 0
	for {
		next, _, err := getPage(c, bid, pageid, 0)
		if err != nil {
			return fmt.Errorf("get pages: %v", err)
		}
		if next == pageid {
			break
		}
		pageid = next
	}
	return nil
}

func getPage(c *api.Client, bid, pid, mod int) (int, int, error) {
	var url string
	switch pid {
	case 0:
		url = c.URL("books/%d/pages/first", bid)
	case -1:
		url = c.URL("books/%d/pages/last", bid)
	default:
		url = c.URL("books/%d/pages/%d", bid, pid)
		url = appendModToURL(url, mod)
	}
	var p api.Page
	if err := get(c, url, &p); err != nil {
		return 0, 0, fmt.Errorf("get page: %v", err)
	}
	format(&p)
	return p.NextPageID, p.PrevPageID, nil
}

func getLine(c *api.Client, bid, pid, lid int) error {
	url := fmt.Sprintf("%s/books/%d/pages/%d/lines/%d", c.Host, bid, pid, lid)
	var line api.Line
	if err := get(c, url, &line); err != nil {
		return fmt.Errorf("get line: %v", err)
	}
	format(&line)
	return nil
}

func getWord(c *api.Client, bid, pid, lid, wid, len int) error {
	var url string
	switch len {
	case -1:
		url = c.URL("books/%d/pages/%d/lines/%d/tokens/%d",
			bid, pid, lid, wid)
	default:
		url = c.URL("books/%d/pages/%d/lines/%d/tokens/%d?len=%d",
			c.Host, bid, pid, lid, wid, len)

	}
	var token api.Token
	if err := get(c, url, &token); err != nil {
		return fmt.Errorf("get word: %v")
	}
	format(&token)
	return nil
}

func getMod(id string) (string, int) {
	if pos := strings.Index(id, "/"); pos != -1 {
		mod, err := strconv.Atoi(id[pos+1:])
		if err != nil {
			return id, 0
		}
		return id[:pos], mod
	}
	return id, 0
}

func appendModToURL(url string, mod int) string {
	if mod > 0 {
		return url + fmt.Sprintf("/next/%d", mod)
	}
	if mod < 0 {
		return url + fmt.Sprintf("/prev/%d", -mod)
	}
	return url
}
