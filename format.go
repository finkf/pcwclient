package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/finkf/gofiler"
	"github.com/finkf/pcwgo/api"
)

var formatArgs = struct {
	template   string
	words      bool
	ocr        bool
	noCor      bool
	onlyManual bool
	json       bool
}{}

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
	case *api.CharMap:
		formatCharMap(t)
	case *api.PostCorrection:
		formatPostCorrection(t)
	case api.Suggestions:
		formatSuggestions(t)
	case *api.SuggestionCounts:
		formatSuggestionCounts(t)
	case *api.PatternCounts:
		formatPatternCounts(t)
	case *api.AdaptiveTokens:
		formatAdaptiveTokens(t)
	case *api.ExtendedLexicon:
		formatExtendedLexicon(t)
	case gofiler.Profile:
		formatProfile(t)
	case api.Session:
		formatSession(t)
	case *api.Users:
		formatUsers(t)
	case *api.User:
		formatUser(t)
	case *api.Books:
		formatBooks(t)
	case *api.Book:
		formatBook(t)
	default:
		log.Fatalf("error: invalid type to print: %T", t)
	}
}

func printf(col *color.Color, format string, args ...interface{}) {
	if col == nil {
		_, err := fmt.Printf(format, args...)
		chk(err)
		return
	}
	_, err := col.Printf(format, args...)
	chk(err)
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
	if formatArgs.onlyManual && !line.IsManuallyCorrected {
		return
	}
	if formatArgs.words {
		formatLineWords(line)
		return
	}
	if !formatArgs.noCor {
		printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			printf(nil, " ")
			printf(colorForToken(&w), w.Cor)
		}
		printf(nil, "\n")
	}
	if formatArgs.ocr {
		printf(nil, "%d:%d:%d", line.ProjectID, line.PageID, line.LineID)
		for _, w := range line.Tokens {
			printf(nil, " ")
			printf(nil, w.OCR)
		}
		printf(nil, "\n")
	}
}

func formatLineWords(line *api.Line) {
	if formatArgs.onlyManual && !line.IsManuallyCorrected {
		return
	}
	for _, w := range line.Tokens {
		formatWord(&w)
	}
}

func formatWord(w *api.Token) {
	if formatArgs.onlyManual && !w.IsManuallyCorrected {
		return
	}
	if !formatArgs.noCor {
		printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		printf(colorForToken(w), w.Cor)
		printf(nil, "\n")
	}
	if formatArgs.ocr {
		printf(nil, "%d:%d:%d:%d ", w.ProjectID, w.PageID, w.LineID, w.TokenID)
		printf(nil, w.Cor)
		printf(nil, "\n")
	}
}

func formatCharMap(chars *api.CharMap) {
	for char, n := range chars.CharMap {
		fmt.Printf("%d %d %s %d\n", chars.BookID, chars.ProjectID, char, n)
	}
}

func formatPostCorrection(pcs *api.PostCorrection) {
	for _, pc := range pcs.Corrections {
		fmt.Printf("%d:%d:%d:%d %s %s %f %t\n",
			pcs.BookID, pc.PageID, pc.LineID, pc.TokenID,
			pc.OCR, pc.Cor, pc.Confidence, pc.Taken)
	}
}

func formatSearchResults(res *api.SearchResults) {
	for _, m := range res.Matches {
		for _, line := range m.Lines {
			if formatArgs.words {
				for _, w := range line.Tokens {
					// Skip not matched tokens
					if !w.IsMatch {
						continue
					}
					// It does not make sense to
					// mark each matched token in
					// red.  Just print the normal
					// token with its normal
					// color.
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

func formatSuggestionCounts(counts *api.SuggestionCounts) {
	for k, v := range counts.Counts {
		printf(nil, "%d %d %s %d\n", counts.BookID, counts.ProjectID, k, v)
	}
}

func formatPatternCounts(counts *api.PatternCounts) {
	for k, v := range counts.Counts {
		printf(nil, "%d %d %s %d %t\n", counts.BookID, counts.ProjectID, k, v, counts.OCR)
	}
}

func formatAdaptiveTokens(tokens *api.AdaptiveTokens) {
	for _, token := range tokens.AdaptiveTokens {
		printf(nil, "%d %d %s\n", tokens.BookID, tokens.ProjectID, token)
	}
}

func formatExtendedLexicon(lex *api.ExtendedLexicon) {
	for entry, n := range lex.Yes {
		printf(nil, "%d %d %s %d %t\n", lex.BookID, lex.ProjectID, entry, n, true)
	}
	for entry, n := range lex.No {
		printf(nil, "%d %d %s %d %t\n", lex.BookID, lex.ProjectID, entry, n, false)
	}
}

func formatProfile(profile gofiler.Profile) {
	pats := func(pats []gofiler.Pattern) []string {
		var ret []string
		for _, pat := range pats {
			ret = append(ret, fmt.Sprintf("%s:%s:%d",
				pat.Left, pat.Right, pat.Pos))
		}
		return ret
	}
	for k, v := range profile {
		top := true
		for _, c := range v.Candidates {
			printf(nil, "%s %s %s %s %s %s %d %f %t\n",
				k, c.Suggestion, c.Modern,
				strings.Join(pats(c.HistPatterns), ","),
				strings.Join(pats(c.OCRPatterns), ","),
				c.Dict, c.Distance, c.Weight, top)
			top = false
		}
	}
}

func formatSession(s api.Session) {
	printf(nil, "%d %s %s %s %s\n",
		s.User.ID, s.User.Email, s.User.Name, s.Auth,
		time.Unix(s.Expires, 0).Format(time.RFC3339))
}

func formatUsers(users *api.Users) {
	for i := range users.Users {
		formatUser(&users.Users[i])
	}
}

func formatUser(user *api.User) {
	printf(nil, "%d %s %s %s %t\n",
		user.ID, user.Name, user.Email, user.Institute, user.Admin)
}

func formatBooks(books *api.Books) {
	for i := range books.Books {
		formatBook(&books.Books[i])
	}
}

func formatBook(book *api.Book) {
	var typ = "B"
	if !book.IsBook {
		typ = "P"
	}
	printf(nil, "%d %d %s %s %d %s %s %d %s %s %s\n",
		book.BookID, book.ProjectID, book.Author, book.Title,
		len(book.PageIDs), typ, bookStatusString(book),
		book.Year, book.Language, book.ProfilerURL, book.Description)
}

func bookStatusString(book *api.Book) string {
	res := []byte("---")
	if book.Status["profiled"] {
		res[0] = 'p'
	}
	if book.Status["extended-lexicon"] {
		res[1] = 'e'
	}
	if book.Status["post-corrected"] {
		res[2] = 'c'
	}
	return string(res)
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
	if !formatArgs.json {
		return false
	}
	chk(json.NewEncoder(os.Stdout).Encode(data))
	return true
}

func formatMaybeTemplate(data interface{}) bool {
	if formatArgs.template == "" {
		return false
	}
	t, err := template.New("pocwebc").Parse(strings.Replace(formatArgs.template, "\\n", "\n", -1))
	chk(err)
	err = t.Execute(os.Stdout, data)
	chk(err)
	return true
}

type patterns []string

func (ps patterns) String() string {
	if len(ps) == 0 {
		return "::"
	}
	return strings.Join(ps, ",")
}
