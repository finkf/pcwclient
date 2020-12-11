package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var correctArgs = struct {
	typ   string
	stdin bool
}{}

func init() {
	correctCommand.Flags().StringVarP(&correctArgs.typ, "type", "t",
		"automatic", "set correction type")
	correctCommand.Flags().BoolVarP(&correctArgs.stdin, "stdin", "i",
		false, "read IDs and corrections from stdin")
}

var correctCommand = cobra.Command{
	Use:   "correct [ID CORRECTION]...",
	Short: "Correct lines or words",
	Args:  cobra.MinimumNArgs(0),
	RunE:  doCorrect,
}

func doCorrect(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	if !correctArgs.stdin {
		for i := 1; i < len(args); i += 2 {
			id := args[i-1]
			cor := args[i]
			if err := correct(c, id, cor, correctArgs.typ); err != nil {
				return fmt.Errorf("cannot correct: %v", err)
			}
		}
		return nil
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		pos := strings.Index(line, " ")
		if pos == -1 {
			return fmt.Errorf("cannot correct: invalid input line: %q", line)
		}
		id := line[:pos]
		cor := line[pos+1:]
		if err := correct(c, id, correctArgs.typ, cor); err != nil {
			return fmt.Errorf("cannot correct: %v", err)
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("cannot correct: %v", err)
	}
	return nil
}

func correct(c *api.Client, id, typ, correction string) error {
	cor, err := strconv.Unquote(`"` + correction + `"`)
	if err != nil {
		return fmt.Errorf("unqote %s: %v", correction, err)
	}
	var url string
	var resp interface{}
	var line api.Line
	var token api.Token
	var bid, pid, lid, wid, len int
	switch n := parseIDs(id, &bid, &pid, &lid, &wid, &len); n {
	case 3:
		url = c.URL("books/%d/pages/%d/lines/%d?t=%s",
			bid, pid, lid, typ)
		resp = &line
	case 4:
		url = c.URL("books/%d/pages/%d/lines/%d/tokens/%d?t=%s",
			bid, pid, lid, wid, typ)
		resp = &token
	case 5:
		url = c.URL("books/%d/pages/%d/lines/%d/tokens/%d?t=%s&len=%d",
			bid, pid, lid, wid, typ, len)
		resp = &token
	default:
		return fmt.Errorf("invalid id: %q", id)
	}
	err = c.Put(url, struct {
		Cor string `json:"correction"`
	}{cor}, resp)
	if err != nil {
		return err
	}
	format(resp)
	return nil
}
