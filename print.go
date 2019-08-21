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
	cmd := newCommand(os.Stdout)
	for _, id := range args {
		var bid, pid, lid, wid, len int
		switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
		case 5:
			getWord(&cmd, bid, pid, lid, wid, len)
		case 4:
			getWord(&cmd, bid, pid, lid, wid, -1)
		case 3:
			getLine(&cmd, bid, pid, lid)
		case 2:
			getPage(&cmd, bid, pid)
		case 1:
			getPages(&cmd, bid)
		default:
			return fmt.Errorf("invalid id: %s", id)
		}
	}
	return cmd.done()
}

func getPages(cmd *command, bid int) {
	var book *api.Book
	cmd.do(func(client *api.Client) (interface{}, error) {
		var err error
		book, err = client.GetBook(bid)
		return nil, err
	})
	for _, pid := range book.PageIDs {
		getPage(cmd, book.ProjectID, pid)
	}
}

func getPage(cmd *command, bid, pid int) {
	cmd.do(func(client *api.Client) (interface{}, error) {
		return client.GetPage(bid, pid)
	})
}

func getLine(cmd *command, bid, pid, lid int) {
	cmd.do(func(client *api.Client) (interface{}, error) {
		return cmd.client.GetLine(bid, pid, lid)
	})
}

func getWord(cmd *command, bid, pid, lid, wid, len int) {
	cmd.do(func(client *api.Client) (interface{}, error) {
		if len == -1 {
			return cmd.client.GetToken(bid, pid, lid, wid)
		}
		return cmd.client.GetTokenLen(bid, pid, lid, wid, len)
	})
}
