package main

import (
	"io"
	"os"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var correctLineCommand = cobra.Command{
	Use:   "correct LINE",
	Short: "Correct lines",
	RunE:  doCorrectLine,
	Args:  cobra.MinimumNArgs(1),
}

func doCorrectLine(cmd *cobra.Command, args []string) error {
	return correctLine(strings.Join(args, " "), os.Stdout)
}

func correctLine(line string, out io.Writer) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		cor := api.CorrectLineRequest{Correction: line}
		line, err := cmd.client.PostLine(bookID, pageID, lineID, cor)
		cmd.data = line
		return err
	})
	return cmd.output(func() error {
		return doPrintLine(out, *cmd.data.(*api.Line))
	})
}
