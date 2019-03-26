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
	Args:  cobra.MinimumNArgs(1),
}

var correctLineCommand = cobra.Command{
	Use:   "line LINE",
	Short: "Correct lines",
	RunE:  doCorrectLine,
}

func doCorrectLine(cmd *cobra.Command, args []string) error {
	return correctLine(strings.Join(args, " "), os.Stdout)
}

func correctLine(line string, out io.Writer) error {
	if bookID == 0 || pageID == 0 || lineID == 0 {
		return fmt.Errorf("missing book, page or line id")
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: line}
		line, err := cmd.client.PostLine(bookID, pageID, lineID, cor)
		cmd.data = line
		return err
	})
	return cmd.output(func() error {
		return doPrintLine(out, *cmd.data.(*api.Line))
	})
}

var correctWordCommand = cobra.Command{
	Use:   "word WORD",
	Short: "Correct words",
	RunE:  doCorrectWord,
	Args:  cobra.MinimumNArgs(1),
}

func doCorrectWord(cmd *cobra.Command, args []string) error {
	return correctWord(strings.Join(args, " "), os.Stdout)
}

func correctWord(word string, out io.Writer) error {
	if bookID == 0 || pageID == 0 || lineID == 0 || wordID == 0 {
		return fmt.Errorf("missing book, page, line or word id")
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.Correction{Correction: word}
		word, err := cmd.client.PostToken(bookID, pageID, lineID, wordID, cor)
		cmd.data = word
		return err
	})
	return cmd.output(func() error {
		return doPrintWord(out, *cmd.data.(*api.Token))
	})
}
