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

var correctCommand = cobra.Command{
	Use:   "correct [ID CORRECTION...]",
	Short: "Correct lines or words",
	Args: func(c *cobra.Command, args []string) error {
		// zero or 2+ args
		if len(args) == 1 {
			return fmt.Errorf("expected exactly 0 or at least 2 args")
		}
		return nil
	},
	RunE: doCorrect,
}

func doCorrect(c *cobra.Command, args []string) error {
	if len(args) >= 2 {
		return correct(os.Stdout, args[0], strings.Join(args[1:], " "))
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		args := strings.Fields(s.Text())
		if len(args) < 2 {
			return fmt.Errorf("invalid input line: %q", s.Text())
		}
		if err := correct(os.Stdout, args[0], strings.Join(args[1:], " ")); err != nil {
			return err
		}
	}
	return s.Err()
}

func correct(out io.Writer, id, correction string) error {
	u, err := strconv.Unquote(`"` + correction + `"`)
	if err != nil {
		return err
	}
	var bid, pid, lid, wid, len int
	switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
	case 3:
		return correctLine(os.Stdout, bid, pid, lid, u)
	case 4:
		return correctWord(os.Stdout, bid, pid, lid, wid, -1, u)
	case 5:
		return correctWord(os.Stdout, bid, pid, lid, wid, len, u)
	default:
		return fmt.Errorf("invalid id: %q", id)
	}
}

func correctLine(out io.Writer, bid, pid, lid int, cor string) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.PutLine(bid, pid, lid, cor)
	})
	return c.done()
}

func correctWord(out io.Writer, bid, pid, lid, wid, len int, cor string) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		if len == -1 {
			return client.PutToken(bid, pid, lid, wid, cor)
		}
		return client.PutTokenLen(bid, pid, lid, wid, len, cor)
	})
	return c.done()
}
