package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/finkf/pcwgo/api"
)

type formatter interface {
	format(io.Writer) error
	underlying() interface{}
}

func formatSlice(out io.Writer, prefix, str string, col *color.Color) error {
	if _, err := fmt.Fprint(out, prefix); err != nil {
		return err
	}
	return formatMaybeColored(out, col, str)
}

func getCol(t *api.Token) *color.Color {
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

func formatMaybeColored(out io.Writer, col *color.Color, strs ...interface{}) error {
	if col == nil {
		_, err := fmt.Fprint(out, strs...)
		return err
	}
	_, err := col.Fprint(out, strs...)
	return err
}

type lineF struct {
	line                  *api.Line
	cor, ocr, skip, words bool
}

func (f lineF) underlying() interface{} {
	return f.line
}

func (f lineF) format(out io.Writer) error {
	if f.words {
		return f.formatWords(out)
	}
	return f.formatLine(out)
}

func (f lineF) formatLine(out io.Writer) error {
	if f.skip && !f.line.IsManuallyCorrected {
		return nil
	}
	if f.cor {
		if _, err := fmt.Fprintf(out, "%d:%d:%d",
			f.line.ProjectID, f.line.PageID, f.line.LineID); err != nil {
			return err
		}
		for _, w := range f.line.Tokens {
			if err := formatSlice(out, " ", w.Cor, getCol(&w)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}

	}
	if f.ocr {
		if _, err := fmt.Fprintf(out, "%d:%d:%d",
			f.line.ProjectID, f.line.PageID, f.line.LineID); err != nil {
			return err
		}
		for _, w := range f.line.Tokens {
			if err := formatSlice(out, " ", w.OCR, nil); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}

	}
	return nil
}

func (f lineF) formatWords(out io.Writer) error {
	tf := tokenF{
		cor:  f.cor,
		ocr:  f.ocr,
		skip: f.skip,
	}
	for _, w := range f.line.Tokens {
		tf.token = &w
		if err := tf.format(out); err != nil {
			return err
		}
	}
	return nil
}

type tokenF struct {
	token          *api.Token
	cor, ocr, skip bool
}

func (f tokenF) underlying() interface{} {
	return f.token
}

func (f tokenF) format(out io.Writer) error {
	if f.skip && !f.token.IsManuallyCorrected {
		return nil
	}
	id := fmt.Sprintf("%d:%d:%d:%d ",
		f.token.ProjectID, f.token.PageID, f.token.LineID, f.token.TokenID)
	if f.cor {
		if err := formatSlice(out, id, f.token.Cor, getCol(f.token)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	if f.ocr {
		if err := formatSlice(out, id, f.token.OCR, nil); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	return nil
}

type pageF struct {
	page                  *api.Page
	cor, ocr, skip, words bool
}

func (f pageF) underlying() interface{} {
	return f.page
}

func (f pageF) format(out io.Writer) error {
	lf := lineF{
		cor:   f.cor,
		ocr:   f.ocr,
		skip:  f.skip,
		words: f.words,
	}
	for _, l := range f.page.Lines {
		lf.line = &l
		if err := lf.format(out); err != nil {
			return err
		}
	}
	return nil
}

type searchF struct {
	results *api.SearchResults
	words   bool
}

func (f searchF) underlying() interface{} {
	return f.results
}

func (f searchF) format(out io.Writer) error {
	cb := f.formatLine
	if f.words {
		cb = f.formatWords
	}
	for _, m := range f.results.Matches {
		for _, line := range m.Lines {
			if err := cb(out, &line); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f searchF) formatLine(out io.Writer, line *api.Line) error {
	if _, err := fmt.Fprintf(out, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID); err != nil {
		return err
	}
	var epos int
	for _, t := range line.Tokens {
		if epos == 0 || t.Offset != epos {
			if _, err := fmt.Fprint(out, " "); err != nil {
				return err
			}
			if err := formatMaybeColored(out, getCol(&t), t.Cor); err != nil {
				return err
			}
			corlen := len([]rune(t.Cor))
			ocrlen := len([]rune(t.OCR))
			maxlen := corlen
			if maxlen < ocrlen {
				maxlen = ocrlen
			}
			epos = t.Offset + maxlen
		}
	}
	_, err := fmt.Fprintln(out)
	return err
}

func (f searchF) formatWords(out io.Writer, line *api.Line) error {
	for _, t := range line.Tokens {
		if !t.IsMatch {
			continue
		}
		if _, err := fmt.Fprintf(out, "%d:%d:%d:%d %s\n",
			t.ProjectID, t.PageID, t.LineID, t.TokenID, t.Cor); err != nil {
			return err
		}
	}
	return nil
}

type suggestionsF struct {
	suggestions api.Suggestions
}

func (f suggestionsF) underlying() interface{} {
	return f.suggestions
}

func (f suggestionsF) format(out io.Writer) error {
	for _, xs := range f.suggestions.Suggestions {
		for _, s := range xs {
			_, err := fmt.Fprintf(out, "%d %s %s %s %s %s %s %d %f %t\n",
				f.suggestions.ProjectID, s.Token, s.Suggestion, s.Modern,
				formatPatterns(s.HistPatterns), formatPatterns(s.OCRPatterns),
				s.Dict, s.Distance, s.Weight, s.Top)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func formatPatterns(patterns []string) string {
	if len(patterns) == 0 {
		return "::"
	}
	return strings.Join(patterns, ",")
}
