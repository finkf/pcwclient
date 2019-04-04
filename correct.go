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
	Use:   "correct",
	Short: "Correct lines or words",
}

var correctLineCommand = cobra.Command{
	Use:   "line ID LINE",
	Short: "Correct lines",
	Args:  cobra.MinimumNArgs(2),
	RunE:  doCorrectLine,
}

func doCorrectLine(cmd *cobra.Command, args []string) error {
	return correctLine(os.Stdout, args[0], strings.Join(args[1:], " "))
}

func correctLine(out io.Writer, id, correction string) error {
	var bid, pid, lid int
	if err := scanf(id, "%d:%d:%d", &bid, &pid, &lid); err != nil {
		return fmt.Errorf("invalid line id: %s", id)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: correction}
		line, err := cmd.client.PostLine(bid, pid, lid, cor)
		cmd.data = line
		return err
	})
	return cmd.output(func() error {
		return cmd.print(out, cmd.data)
	})
}

var correctWordCommand = cobra.Command{
	Use:   "word ID WORD",
	Short: "Correct words",
	Args:  cobra.ExactArgs(2),
	RunE:  doCorrectWord,
}

func doCorrectWord(cmd *cobra.Command, args []string) error {
	return correctWord(os.Stdout, args[0], args[1])
}

func correctWord(out io.Writer, id, correction string) error {
	var bid, pid, lid, wid int
	if err := scanf(id, "%d:%d:%d:%d", &bid, &pid, &lid, &wid); err != nil {
		return fmt.Errorf("invalid line id: %s", id)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: correction}
		word, err := cmd.client.PostToken(bid, pid, lid, wid, cor)
		cmd.data = word
		return err
	})
	return cmd.output(func() error {
		return cmd.print(out, cmd.data)
	})
}
