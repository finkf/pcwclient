package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func init() {
	printCommand.Flags().BoolVarP(&formatWords, "words", "w", false,
		"print words not lines")
	printCommand.Flags().BoolVarP(&formatOCR, "ocr", "o", false,
		"print ocr lines")
	printCommand.Flags().BoolVarP(&noFormatCor, "nocor", "c", false,
		"do not print corrected lines")
	printCommand.Flags().BoolVarP(&formatOnlyManual, "skip", "s", false,
		"skip non corrected lines/words")
}

var printCommand = cobra.Command{
	Use:   "print IDs...",
	Short: "print books, pages, lines and words",
	RunE:  printIDs,
}

func printIDs(_ *cobra.Command, args []string) error {
	c := newClient(os.Stdout)
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

func doPrintID(c *client, id string) error {
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
		return fmt.Errorf("invalid id: %s", id)
	}
	return nil
}

func getPages(c *client, bid int) {
	pageid := 0
	done := false
	var f formatter
	defer f.done()
	for !done {
		c.do(func(client *api.Client) (interface{}, error) {
			p, err := getPageImpl(client, bid, pageid)
			if err != nil {
				done = true
				return nil, err
			}
			if p.PageID == p.NextPageID {
				done = true
			}
			pageid = p.NextPageID
			f.format(p)
			return nil, nil
		})
	}
}

func getPage(c *client, bid, pid int) {
	var f formatter
	defer f.done()
	c.do(func(client *api.Client) (interface{}, error) {
		var p *api.Page
		var err error
		switch pid {
		case 0:
			p, err = client.GetFirstPage(bid)
		case -1:
			p, err = client.GetLastPage(bid)
		default:
			p, err = client.GetPage(bid, pid)
		}
		must(err, "cannot get page: %v")
		f.format(p)
		return nil, nil
	})
}

func getPageImpl(client *api.Client, bid, pid int) (*api.Page, error) {
	switch pid {
	case 0:
		return client.GetFirstPage(bid)
	case -1:
		return client.GetLastPage(bid)
	default:
		return client.GetPage(bid, pid)
	}
}

func getLine(c *client, bid, pid, lid int) {
	var f formatter
	defer f.done()
	c.do(func(client *api.Client) (interface{}, error) {
		l, err := c.client.GetLine(bid, pid, lid)
		must(err, "cannot get line: %v")
		f.format(l)
		return nil, nil
	})
}

func getWord(c *client, bid, pid, lid, wid, len int) {
	var f formatter
	defer f.done()
	c.do(func(client *api.Client) (interface{}, error) {
		var err error
		var t *api.Token
		switch len {
		case -1:
			t, err = c.client.GetToken(bid, pid, lid, wid)
		default:
			t, err = c.client.GetTokenLen(bid, pid, lid, wid, len)
		}
		must(err, "cannot get word: %v")
		f.format(t)
		return nil, nil
	})
}
