package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var corType string

func init() {
	correctCommand.Flags().StringVarP(&corType, "type", "t",
		"automatic", "set correction type")
}

var correctCommand = cobra.Command{
	Use:   "correct [ID CORRECTION]...",
	Short: "Correct lines or words",
	Args: func(c *cobra.Command, args []string) error {
		// zero or 2+ args
		if len(args)%2 != 0 {
			return fmt.Errorf("expected an even number of arugments")
		}
		return nil
	},
	RunE: doCorrect,
}

func doCorrect(c *cobra.Command, args []string) error {
	for i := 1; i < len(args); i += 2 {
		if err := correct(os.Stdout, args[i-1], args[i],
			api.CorType(corType)); err != nil {
			return err
		}
	}
	if len(args) > 0 {
		return nil
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		args := strings.Fields(s.Text())
		if len(args) < 2 {
			return fmt.Errorf("invalid input line: %q", s.Text())
		}
		cor := strings.Join(args[1:], " ")
		if err := correct(os.Stdout, args[0], cor, api.CorType(corType)); err != nil {
			return err
		}
	}
	return s.Err()
}

func correct(out io.Writer, id, correction string, typ api.CorType) error {
	cor, err := strconv.Unquote(`"` + correction + `"`)
	if err != nil {
		return err
	}
	var bid, pid, lid, wid, len int
	switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
	case 3:
		line := api.Line{ProjectID: bid, PageID: pid, LineID: lid}
		return correctLine(os.Stdout, &line, api.CorType(typ), cor)
	case 4:
		token := api.Token{ProjectID: bid, PageID: pid, LineID: lid, TokenID: wid}
		return correctWord(os.Stdout, &token, -1, api.CorType(typ), cor)
	case 5:
		token := api.Token{ProjectID: bid, PageID: pid, LineID: lid, TokenID: wid}
		return correctWord(os.Stdout, &token, len, api.CorType(typ), cor)
	default:
		return fmt.Errorf("invalid id: %q", id)
	}
}

func correctLine(out io.Writer, line *api.Line, typ api.CorType, cor string) error {
	c := newClient(out)
	var f formatter
	c.do(func(client *api.Client) (interface{}, error) {
		must(client.PutLineX(line, typ, cor), "cannot correct line: %v")
		f.format(line)
		return nil, nil
	})
	return c.done()
}

func correctWord(out io.Writer, token *api.Token, len int, typ api.CorType, cor string) error {
	c := newClient(out)
	var f formatter
	c.do(func(client *api.Client) (interface{}, error) {
		if len == -1 {
			must(client.PutTokenX(token, typ, cor), "cannot correct token: %v")
			f.format(token)
			return nil, nil
		}
		must(client.PutTokenLenX(token, len, typ, cor), "cannot correct token: %v")
		f.format(token)
		return nil, nil
	})
	return c.done()
}
