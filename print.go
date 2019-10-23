package main

import (
	"fmt"
	"os"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var (
	printWords bool
)

func init() {
	printCommand.Flags().BoolVarP(&printWords, "words", "w", false,
		"print words not lines")
}

var printCommand = cobra.Command{
	Use:   "print IDs...",
	Short: "print books, pages, lines and words",
	Args:  cobra.MinimumNArgs(1),
	RunE:  printIDs,
}

func printIDs(_ *cobra.Command, args []string) error {
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
			page, err := getPageImpl(client, bid, pageid)
			if err != nil {
				done = true
				return nil, err
			}
			if page.PageID == page.NextPageID {
				done = true
			}
			pageid = page.NextPageID
			return page, nil
		})
	}
}

func getPage(c *client, bid, pid int) {
	c.do(func(client *api.Client) (interface{}, error) {
		switch pid {
		case 0:
			return client.GetFirstPage(bid)
		case -1:
			return client.GetLastPage(bid)
		default:
			return client.GetPage(bid, pid)
		}
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
		return c.client.GetLine(bid, pid, lid)
	})
}

func getWord(c *client, bid, pid, lid, wid, len int) {
	c.do(func(client *api.Client) (interface{}, error) {
		if len == -1 {
			return c.client.GetToken(bid, pid, lid, wid)
		}
		return c.client.GetTokenLen(bid, pid, lid, wid, len)
	})
}
