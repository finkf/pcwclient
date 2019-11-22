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

type formatter struct {
	data []interface{}
}

func (f *formatter) format(data interface{}) {
	if formatJSON || formatTemplate != "" {
		f.data = append(f.data, data)
		return
	}
	switch t := data.(type) {
	case *api.Page:
		f.formatPage(t)
	case *api.Line:
		f.formatLine(t)
	case *api.Token:
		f.formatWord(t)
	case *api.SearchResults:
		f.formatSearchResults(t)
	case api.Suggestions:
		f.formatSuggestions(t)
	default:
		log.Fatalf("error: invalid type to print: %T", t)
	}
}

func (f *formatter) done() {
	if !formatJSON && formatTemplate == "" {
		return
	}
	var toPrint interface{}
	switch len(f.data) {
	case 0:
		return
	case 1:
		toPrint = f.data[0]
	default:
		toPrint = f.data
	}
	if formatJSON {
		f.formatJSON(toPrint)
	} else {
		f.formatTemplate(toPrint)
	}
}

func (f *formatter) printf(col *color.Color, format string, args ...interface{}) {
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

func (f *formatter) color(t *api.Token) *color.Color {
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

func (f *formatter) formatPage(page *api.Page) {
	for _, line := range page.Lines {
		f.formatLine(&line)
	}
}

func (f *formatter) formatLine(line *api.Line) {
	if formatOnlyManual && !line.IsManuallyCorrected {
		return
	}
	if formatWords {
		f.formatLineWords(line)
		return
	}
	if !noFormatCor {
		f.printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			f.printf(nil, " ")
			f.printf(f.color(&w), w.Cor)
		}
		f.printf(nil, "\n")
	}
	if formatOCR {
		f.printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			f.printf(nil, " ")
			f.printf(nil, w.OCR)
		}
		f.printf(nil, "\n")
	}
}

func (f *formatter) formatLineWords(line *api.Line) {
	if formatOnlyManual && !line.IsManuallyCorrected {
		return
	}
	for _, w := range line.Tokens {
		f.formatWord(&w)
	}
}

func (f *formatter) formatWord(w *api.Token) {
	if formatOnlyManual && !w.IsManuallyCorrected {
		return
	}
	if !noFormatCor {
		f.printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		f.printf(f.color(w), w.Cor)
		f.printf(nil, "\n")
	}
	if formatOCR {
		f.printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		f.printf(nil, w.Cor)
		f.printf(nil, "\n")
	}
}

func (f *formatter) formatSearchResults(res *api.SearchResults) {
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
					f.formatWord(&w)
				}
			} else {
				f.formatLine(&line)
			}
		}
	}
}

func (f *formatter) formatSuggestions(suggs api.Suggestions) {
	for _, sugg := range suggs.Suggestions {
		for _, s := range sugg {
			f.printf(nil, "%d %s %s %s %s %s %s %d %f %t\n",
				suggs.ProjectID, s.Token, s.Suggestion, s.Modern,
				patterns(s.HistPatterns), patterns(s.OCRPatterns),
				s.Dict, s.Distance, s.Weight, s.Top)
		}
	}
}

func (f *formatter) formatJSON(data interface{}) {
	handle(json.NewEncoder(os.Stdout).Encode(data), "cannot encode json: %v")
}

func (f *formatter) formatTemplate(data interface{}) error {
	t, err := template.New("pocwebc").Parse(strings.Replace(formatTemplate, "\\n", "\n", -1))
	handle(err, "invalid format string: %v")
	err = t.Execute(os.Stdout, data)
	handle(err, "cannot format template: %v")
	return nil
}

type patterns []string

func (ps patterns) String() string {
	if len(ps) == 0 {
		return "::"
	}
	return strings.Join(ps, ",")
}
