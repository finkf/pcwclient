package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
)

var (
	formatWords      bool
	formatOCR        bool
	noFormatCor      bool
	formatOnlyManual bool
	formatJSON       bool
	formatTemplate   string
)

func handle(err error, args ...interface{}) {
	if err == nil {
		return
	}
	if len(args) > 0 {
		err = fmt.Errorf(args[0].(string), append(args[1:], err)...)
	}
	log.Fatalf("error: %v", err)
}

func format(data interface{}) {
	if formatMaybeSpecial(data) {
		return
	}
	switch t := data.(type) {
	case *api.Page:
		formatPage(t)
	case *api.Line:
		formatLine(t)
	case *api.Token:
		formatWord(t)
	case *api.SearchResults:
		formatSearchResults(t)
	case api.Suggestions:
		formatSuggestions(t)
	default:
		log.Fatalf("error: invalid type to print: %T", t)
	}
}

func printf(col *color.Color, format string, args ...interface{}) {
	if col == nil {
		_, err := fmt.Printf(format, args...)
		handle(err)
		return
	}
	_, err := col.Printf(format, args...)
	handle(err)
}

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
)

func colorForToken(t *api.Token) *color.Color {
	if t.IsMatch {
		return red
	}
	if t.IsManuallyCorrected {
		return green
	}
	if t.IsAutomaticallyCorrected {
		return yellow
	}
	return nil
}

func formatPage(page *api.Page) {
	for _, line := range page.Lines {
		formatLine(&line)
	}
}

func formatLine(line *api.Line) {
	if formatOnlyManual && !line.IsManuallyCorrected {
		return
	}
	if formatWords {
		formatLineWords(line)
		return
	}
	if !noFormatCor {
		printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			printf(nil, " ")
			printf(colorForToken(&w), w.Cor)
		}
		printf(nil, "\n")
	}
	if formatOCR {
		printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			printf(nil, " ")
			printf(nil, w.OCR)
		}
		printf(nil, "\n")
	}
}

func formatLineWords(line *api.Line) {
	if formatOnlyManual && !line.IsManuallyCorrected {
		return
	}
	for _, w := range line.Tokens {
		formatWord(&w)
	}
}

func formatWord(w *api.Token) {
	if formatOnlyManual && !w.IsManuallyCorrected {
		return
	}
	if !noFormatCor {
		printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		printf(colorForToken(w), w.Cor)
		printf(nil, "\n")
	}
	if formatOCR {
		printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		printf(nil, w.Cor)
		printf(nil, "\n")
	}
}

func formatSearchResults(res *api.SearchResults) {
	for _, m := range res.Matches {
		for _, line := range m.Lines {
			if formatWords {
				for _, w := range line.Tokens {
					// skip non matched tokens
					if !w.IsMatch {
						continue
					}
					// It does not make sense to mark each matched token in red.
					// Just print the normal token with its normal color.
					w.IsMatch = false
					formatWord(&w)
				}
			} else {
				formatLine(&line)
			}
		}
	}
}

func formatSuggestions(suggs api.Suggestions) {
	for _, sugg := range suggs.Suggestions {
		for _, s := range sugg {
			printf(nil, "%d %s %s %s %s %s %s %d %f %t\n",
				suggs.ProjectID, s.Token, s.Suggestion, s.Modern,
				patterns(s.HistPatterns), patterns(s.OCRPatterns),
				s.Dict, s.Distance, s.Weight, s.Top)
		}
	}
}

func formatMaybeSpecial(data interface{}) bool {
	if formatMaybeJSON(data) {
		return true
	}
	if formatMaybeTemplate(data) {
		return true
	}
	return false
}

func formatMaybeJSON(data interface{}) bool {
	if !formatJSON {
		return false
	}
	handle(json.NewEncoder(os.Stdout).Encode(data), "cannot encode json: %v")
	return true
}

func formatMaybeTemplate(data interface{}) bool {
	if formatTemplate == "" {
		return false
	}
	t, err := template.New("pocwebc").Parse(strings.Replace(formatTemplate, "\\n", "\n", -1))
	handle(err, "invalid format string: %v")
	err = t.Execute(os.Stdout, data)
	handle(err, "cannot format template: %v")
	return true
}

type patterns []string

func (ps patterns) String() string {
	if len(ps) == 0 {
		return "::"
	}
	return strings.Join(ps, ",")
}
