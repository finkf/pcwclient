package main

import (
	"bufio"
	"fmt"
	"log"
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
		getWord(c, bid, pid, lid, wid, len)
	case 4:
		getWord(c, bid, pid, lid, wid, -1)
	case 3:
		getLine(c, bid, pid, lid)
	case 2:
		getPage(c, bid, pid, mod)
	case 1:
		getPages(c, bid)
	default:
		log.Fatalf("invalid id: %s", id)
	}
	return nil
}

func getPages(c *api.Client, bid int) {
	pageid := 0
	for {
		next, _ := getPage(c, bid, pageid, 0)
		if next == pageid {
			break
		}
		pageid = next
	}
}

func getPage(c *api.Client, bid, pid, mod int) (int, int) {
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
	handle(get(c, url, &p), "cannot get page: %v")
	format(&p)
	return p.NextPageID, p.PrevPageID
}

func getLine(c *api.Client, bid, pid, lid int) {
	url := fmt.Sprintf("%s/books/%d/pages/%d/lines/%d", c.Host, bid, pid, lid)
	var line api.Line
	handle(get(c, url, &line), "cannot get line: %v")
	format(&line)
}

func getWord(c *api.Client, bid, pid, lid, wid, len int) {
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
	handle(get(c, url, &token), "cannot get word: %v")
	format(&token)
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
