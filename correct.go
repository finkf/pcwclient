package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var correctCommand = cobra.Command{
	Use:   "correct ID CORRECTION...",
	Short: "Correct lines or words",
	Args:  cobra.MinimumNArgs(2),
	RunE:  doCorrect,
}

func doCorrect(cmd *cobra.Command, args []string) error {
	id, correction := args[0], strings.Join(args[1:], " ")
	if bid, pid, lid, wid, ok := wordID(id); ok {
		return correctWord(os.Stdout, bid, pid, lid, wid, correction)
	}
	if bid, pid, lid, ok := lineID(id); ok {
		return correctLine(os.Stdout, bid, pid, lid, correction)
	}
	return fmt.Errorf("invalid id: %q", id)
}

func correctLine(out io.Writer, bid, pid, lid int, correction string) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: correction}
		line, err := cmd.client.PostLine(bid, pid, lid, cor)
		cmd.add(line)
		return err
	})
	return cmd.print()
}

func correctWord(out io.Writer, bid, pid, lid, wid int, correction string) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: correction}
		word, err := cmd.client.PostToken(bid, pid, lid, wid, cor)
		cmd.add(word)
		return err
	})
	return cmd.print()
}
