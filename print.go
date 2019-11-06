package main

import (
	"fmt"
	"os"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var (
	printWords bool
	printOCR   bool
	printCor   bool
	skipNonCor bool
)

func init() {
	printCommand.Flags().BoolVarP(&printWords, "words", "w", false,
		"print words not lines")
	printCommand.Flags().BoolVarP(&printOCR, "ocr", "o", false,
		"print ocr lines instead of cor")
	printCommand.Flags().BoolVarP(&printCor, "cor", "c", false,
		"print corrected lines (set to print ocr and cor lines)")
	printCommand.Flags().BoolVarP(&skipNonCor, "skip", "s", false,
		"skip non corrected lines/words")
}

var printCommand = cobra.Command{
	Use:   "print IDs...",
	Short: "print books, pages, lines and words",
	Args:  cobra.MinimumNArgs(1),
	RunE:  printIDs,
}

func printIDs(_ *cobra.Command, args []string) error {
	if !printOCR && !printCor {
		printCor = true
	}
	c := newClient(os.Stdout)
	for _, id := range args {
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
	}
	return c.done()
}

func getPages(c *client, bid int) {
	pageid := 0
	done := false
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
			return pageF{
				page:  p,
				cor:   printCor,
				ocr:   printOCR,
				words: printWords,
				skip:  skipNonCor,
			}, nil
		})
	}
}

func getPage(c *client, bid, pid int) {
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
		return pageF{
			page:  p,
			cor:   printCor,
			ocr:   printOCR,
			words: printWords,
			skip:  skipNonCor,
		}, err
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
	c.do(func(client *api.Client) (interface{}, error) {
		l, err := c.client.GetLine(bid, pid, lid)
		return lineF{
			line:  l,
			cor:   printCor,
			ocr:   printOCR,
			words: printWords,
			skip:  skipNonCor,
		}, err
	})
}

func getWord(c *client, bid, pid, lid, wid, len int) {
	c.do(func(client *api.Client) (interface{}, error) {
		var err error
		var t *api.Token
		switch len {
		case -1:
			t, err = c.client.GetToken(bid, pid, lid, wid)
		default:
			t, err = c.client.GetTokenLen(bid, pid, lid, wid, len)
		}
		return tokenF{
			token: t,
			cor:   printCor,
			ocr:   printOCR,
			skip:  skipNonCor,
		}, err
	})
}
