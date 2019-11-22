package main

import (
	"bufio"
	"os"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	printCommand.Flags().BoolVarP(&formatWords, "words", "w", false,
		"print words not lines")
	printCommand.Flags().BoolVarP(&formatOCR, "ocr", "o", false,
		"print ocr lines")
	printCommand.Flags().BoolVarP(&noFormatCor, "nocor", "c", false,
		"do not print corrected lines")
	printCommand.Flags().BoolVarP(&formatOnlyManual, "manual", "m", false,
		"only print manual corrected lines/words")
}

var printCommand = cobra.Command{
	Use:   "print IDs...",
	Short: "print books, pages, lines and words",
	RunE:  printIDs,
}

func printIDs(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), skipVerify)
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
	var bid, pid, lid, wid, len int
	switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
	case 5:
		getWord(c, bid, pid, lid, wid, len)
	case 4:
		getWord(c, bid, pid, lid, wid, -1)
	case 3:
		getLine(c, bid, pid, lid)
	case 2:
		getPage(c, bid, pid)
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
		p, err := getPageImpl(c, bid, pageid)
		handle(err, "cannot get page %d: %v", pageid)
		if p.PageID == p.NextPageID {
			break
		}
		pageid = p.NextPageID
		format(p)
	}
}

func getPage(c *api.Client, bid, pid int) {
	var p *api.Page
	var err error
	switch pid {
	case 0:
		p, err = c.GetFirstPage(bid)
	case -1:
		p, err = c.GetLastPage(bid)
	default:
		p, err = c.GetPage(bid, pid)
	}
	handle(err, "cannot get page: %v")
	format(p)
}

func getPageImpl(c *api.Client, bid, pid int) (*api.Page, error) {
	switch pid {
	case 0:
		return c.GetFirstPage(bid)
	case -1:
		return c.GetLastPage(bid)
	default:
		return c.GetPage(bid, pid)
	}
}

func getLine(c *api.Client, bid, pid, lid int) {
	l, err := c.GetLine(bid, pid, lid)
	handle(err, "cannot get line: %v")
	format(l)
}

func getWord(c *api.Client, bid, pid, lid, wid, len int) {
	var err error
	var t *api.Token
	switch len {
	case -1:
		t, err = c.GetToken(bid, pid, lid, wid)
	default:
		t, err = c.GetTokenLen(bid, pid, lid, wid, len)
	}
	handle(err, "cannot get word: %v")
	format(t)
}
