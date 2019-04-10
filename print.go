package main

import (
	"fmt"
	"os"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var printWords bool

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
		getByID(&cmd, id)
	}
	return cmd.print()
}

func getByID(cmd *command, id string) {
	cmd.do(func() error {
		if bid, pid, lid, wid, ok := wordID(id); ok {
			getWord(cmd, bid, pid, lid, wid)
			return nil
		}
		if bid, pid, lid, ok := lineID(id); ok {
			getLine(cmd, bid, pid, lid)
			return nil
		}
		if bid, pid, ok := pageID(id); ok {
			getPage(cmd, bid, pid)
			return nil
		}
		if bid, ok := bookID(id); ok {
			getPages(cmd, bid)
			return nil
		}
		return fmt.Errorf("invalid id: %s", id)
	})
}

func getPages(cmd *command, bid int) {
	var book *api.Book
	cmd.do(func() error {
		var err error
		book, err = cmd.client.GetBook(bid)
		return err
	})
	cmd.do(func() error {
		for _, pid := range book.PageIDs {
			page, err := cmd.client.GetPage(book.ProjectID, pid)
			if err != nil {
				return err
			}
			cmd.add(page)
		}
		return nil
	})
}

func getPage(cmd *command, bid, pid int) {
	cmd.do(func() error {
		page, err := cmd.client.GetPage(bid, pid)
		cmd.add(page)
		return err
	})
}

func getLine(cmd *command, bid, pid, lid int) {
	cmd.do(func() error {
		line, err := cmd.client.GetLine(bid, pid, lid)
		cmd.add(line)
		return err
	})
}

func getWord(cmd *command, bid, pid, lid, wid int) {
	cmd.do(func() error {
		tokens, err := cmd.client.GetTokens(bid, pid, lid)
		var found bool
		for _, word := range tokens.Tokens {
			if word.TokenID == wid {
				cmd.add(&word)
				found = true
				break
			}
		}
		if !found && err == nil {
			return fmt.Errorf("cannot find %d:%d:%d:%d", bid, pid, lid, wid)
		}
		return err
	})
}
