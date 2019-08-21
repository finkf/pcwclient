package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/finkf/gofiler"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// various command line flags
var (
	debug        = false
	jsonOutput   = false
	formatString = ""
	authToken    = ""
	pocowebURL   = ""
	green        = color.New(color.FgGreen)
	red          = color.New(color.FgRed)
	yellow       = color.New(color.FgYellow)
)

var mainCommand = &cobra.Command{
	Use:   "pocowebc",
	Short: "command line client for pocoweb",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	mainCommand.AddCommand(&listCommand)
	mainCommand.AddCommand(&createCommand)
	mainCommand.AddCommand(&loginCommand)
	mainCommand.AddCommand(&logoutCommand)
	mainCommand.AddCommand(&printCommand)
	mainCommand.AddCommand(&versionCommand)
	mainCommand.AddCommand(&rawCommand)
	mainCommand.AddCommand(&searchCommand)
	mainCommand.AddCommand(&correctCommand)
	mainCommand.AddCommand(&downloadCommand)
	mainCommand.AddCommand(&splitCommand)
	mainCommand.AddCommand(&assignCommand)
	mainCommand.AddCommand(&takeBackCommand)
	mainCommand.AddCommand(&deleteCommand)
	mainCommand.AddCommand(&startCommand)
	mainCommand.AddCommand(&showCommand)
	listCommand.AddCommand(&listBooksCommand)
	listCommand.AddCommand(&listUsersCommand)
	listCommand.AddCommand(&listPatternsCommand)
	listCommand.AddCommand(&listSuggestionsCommand)
	listCommand.AddCommand(&listSuspiciousCommand)
	listCommand.AddCommand(&listAdaptiveCommand)
	listCommand.AddCommand(&listELCommand)
	listCommand.AddCommand(&listRRDMCommand)
	listCommand.AddCommand(&listOCRModelsCommand)
	listCommand.AddCommand(&listCharsCommand)
	createCommand.AddCommand(&createUserCommand)
	createCommand.AddCommand(&createBookCommand)
	startCommand.AddCommand(&startProfileCommand)
	startCommand.AddCommand(&startELCommand)
	startCommand.AddCommand(&startRRDMCommand)
	startCommand.AddCommand(&startPredictCommand)
	startCommand.AddCommand(&startTrainCommand)
	deleteCommand.AddCommand(&deleteBooksCommand)
	deleteCommand.AddCommand(&deleteUsersCommand)
	mainCommand.SilenceUsage = true
	mainCommand.SilenceErrors = true
	mainCommand.PersistentFlags().BoolVarP(&jsonOutput, "json", "J", false,
		"output raw json")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "D", false,
		"enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U",
		getURL(), "set pocoweb url (env: POCOWEBC_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatString, "format", "F",
		"", "set output format")
	mainCommand.PersistentFlags().StringVarP(&authToken, "auth", "A",
		getAuth(), "set auth token (env: POCOWEBC_AUTH)")
}

func exactlyNIDs(n int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("expected exactly %d arguments", n)
		}
		for _, arg := range args {
			_, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("not an ID: %s", arg)
			}
		}
		return nil
	}
}

func parseIDs(id string, ids ...*int) int {
	split := strings.Split(id, ":")
	var i int
	for i = 0; i < len(ids) && i < len(split); i++ {
		id, err := strconv.Atoi(split[i])
		if err != nil {
			return 0
		}
		*ids[i] = id
	}
	return i
}

type command struct {
	client *api.Client
	data   []interface{}
	out    io.Writer
	err    error
}

func newCommand(out io.Writer) command {
	url, auth := getURL(), getAuth()
	if url == "" || auth == "" {
		return command{
			err: fmt.Errorf("missing login information: see login sub command"),
		}
	}
	log.Debugf("authenticating at %s with %s", url, auth)
	return command{client: api.Authenticate(url, auth), out: out}
}

func (cmd *command) do(f func(*api.Client) (interface{}, error)) {
	if cmd.err != nil {
		return
	}
	val, err := f(cmd.client)
	// cmd.print handles the possible error
	cmd.err = err
	cmd.print(val)
}

func (cmd *command) done() error {
	if cmd.err != nil {
		return cmd.err
	}
	if jsonOutput {
		return cmd.printJSON()
	}
	if formatString != "" {
		return cmd.printTemplate(formatString)
	}
	return nil
}

func (cmd *command) print(val interface{}) {
	if cmd.err != nil {
		return
	}
	// cache printout data for json or format output
	if jsonOutput || formatString != "" {
		cmd.data = append(cmd.data, val)
		return
	}
	if val != nil {
		cmd.printVal(val)
	}
}

func (cmd *command) printVal(val interface{}) {
	if cmd.err != nil {
		return
	}
	// just print the given value
	switch t := val.(type) {
	case *api.Page:
		cmd.printPage(t)
	case *api.Line:
		cmd.printLine(t)
	case *api.Token:
		cmd.printWord(t)
	case []api.Token:
		cmd.printWords(t)
	case []interface{}:
		cmd.printArray(t)
	case api.Session:
		cmd.printSession(t)
	case api.Version:
		cmd.printVersion(t)
	case *api.Book:
		cmd.printBook(t)
	case *api.Books:
		cmd.printBooks(t)
	case api.SplitPackages:
		cmd.printSplitPackages(t)
	case api.User:
		cmd.printUser(t)
	case api.Users:
		cmd.printUsers(t)
	case *api.SearchResults:
		cmd.printSearchResults(t)
	case gofiler.Profile:
		cmd.printProfile(t)
	case api.Suggestions:
		cmd.printSuggestions(t)
	case api.SuggestionCounts:
		cmd.printSuggestionCounts(t)
	case api.Patterns:
		cmd.printPatterns(t)
	case api.PatternCounts:
		cmd.printPatternCounts(t)
	case api.AdaptiveTokens:
		cmd.printAdaptiveTokens(t)
	case api.Models:
		cmd.printModels(t)
	case api.ExtendedLexicon:
		cmd.printExtendedLexicon(t)
	case *api.PostCorrection:
		cmd.printPostCorrection(t)
	case api.CharMap:
		cmd.printCharMap(t)
	default:
		panic(fmt.Sprintf("invalid type to print: %T", val))
	}
}

func (cmd *command) printJSON() error {
	var data interface{} = cmd.data
	if len(cmd.data) == 1 {
		data = cmd.data[0]
	}
	if err := json.NewEncoder(cmd.out).Encode(data); err != nil {
		return fmt.Errorf("error encoding to json: %v", err)
	}
	return nil
}

func (cmd *command) printTemplate(tmpl string) error {
	var data interface{} = cmd.data
	if len(cmd.data) == 1 {
		data = cmd.data[0]
	}
	t, err := template.New("pocwebc").Parse(
		strings.Replace(tmpl, "\\n", "\n", -1))
	if err != nil {
		return fmt.Errorf("invalid format string: %v", err)
	}
	if err = t.Execute(cmd.out, data); err != nil {
		return fmt.Errorf("error formatting string: %v", err)
	}
	return nil
}

func (cmd *command) printf(format string, args ...interface{}) {
	if cmd.err != nil {
		return
	}
	if _, err := fmt.Printf(format, args...); err != nil {
		cmd.err = err
	}
}

func (cmd *command) println(args ...interface{}) {
	if cmd.err != nil {
		return
	}
	if _, err := fmt.Println(args...); err != nil {
		cmd.err = err
	}
}

func (cmd *command) printColored(fc, pc bool, cor string) {
	if cmd.err != nil {
		return
	}
	if fc {
		_, err := green.Print(cor)
		cmd.err = err
		return
	}
	if pc {
		_, err := yellow.Print(cor)
		cmd.err = err
		return
	}
	_, err := fmt.Print(cor)
	cmd.err = err
}

func (cmd *command) printPage(page *api.Page) {
	for _, line := range page.Lines {
		cmd.printLine(&line)
	}
}

func (cmd *command) printLine(line *api.Line) {
	if !printWords {
		cmd.printf("%d:%d:%d ", line.ProjectID, line.PageID, line.LineID)
		cmd.printColored(line.IsFullyCorrected, line.IsPartiallyCorrected, line.Cor)
		cmd.println()
		return
	}
	cmd.printWords(line.Tokens)
}

func (cmd *command) printWords(words []api.Token) {
	for i := range words {
		cmd.printWord(&words[i])
	}
}

func (cmd *command) printWord(word *api.Token) {
	cmd.printf("%d:%d:%d:%d ", word.ProjectID, word.PageID, word.LineID, word.Offset)
	cmd.printColored(word.IsFullyCorrected, word.IsPartiallyCorrected, word.Cor)
	cmd.println()
}

func (cmd *command) printArray(xs []interface{}) {
	for _, x := range xs {
		cmd.printVal(x)
	}
}

func (cmd *command) printSession(s api.Session) {
	cmd.info("%d\t%s\t%s\t%s\t%s\n",
		s.User.ID, s.User.Email, s.User.Name, s.Auth,
		time.Unix(s.Expires, 0).Format(time.RFC3339))
}

func (cmd *command) printVersion(v api.Version) {
	cmd.println(v.Version)
}

func (cmd *command) printSearchResults(res *api.SearchResults) error {
	for _, ms := range res.Matches {
		for _, m := range ms {
			cmd.printColoredSearchMatches(m)
		}
	}
	return nil
}

func (cmd *command) printColoredSearchMatches(m api.Match) {
	cmd.printf("%d:%d:%d", m.Line.ProjectID, m.Line.PageID, m.Line.LineID)
	var epos int
	for _, t := range m.Tokens {
		if epos == 0 || t.Offset != epos {
			cmd.printf(" ")
		}
		if t.IsMatch {
			if _, err := red.Fprint(cmd.out, t.Cor); err != nil {
				cmd.err = err
				return
			}
		} else {
			cmd.printf(t.Cor)
		}
		corlen := len([]rune(t.Cor))
		ocrlen := len([]rune(t.OCR))
		maxlen := corlen
		if maxlen < ocrlen {
			maxlen = ocrlen
		}
		epos = t.Offset + maxlen
	}
	cmd.println()
}

func (cmd *command) printUsers(users api.Users) {
	for _, user := range users.Users {
		cmd.printUser(user)
	}
}

func (cmd *command) printUser(user api.User) {
	cmd.info("%d\t%s\t%s\t%s\t%t\n",
		user.ID, user.Name, user.Email, user.Institute, user.Admin)
}

func (cmd *command) printBooks(books *api.Books) {
	for _, book := range books.Books {
		cmd.printBook(&book)
	}
}

func (cmd *command) printSplitPackages(packages api.SplitPackages) {
	for _, pkg := range packages.Packages {
		cmd.info("%d\t%d\t%d\t%d\n",
			packages.BookID, pkg.ProjectID, pkg.Owner, len(pkg.PageIDs))
	}
}

func toStrings(xs []int) []string {
	res := make([]string, len(xs))
	for i, x := range xs {
		res[i] = fmt.Sprintf("%d", x)
	}
	return res
}

func (cmd *command) printBook(book *api.Book) {
	cmd.info("%d\t%d\t%s\t%s\t%d\t%s\t%s\t%d\t%s\t%s\t%t\n",
		book.BookID, book.ProjectID, book.Author, book.Title,
		len(book.PageIDs), book.Description, bookStatusString(book),
		book.Year, book.Language, book.ProfilerURL, book.IsBook)
}

func bookStatusString(book *api.Book) string {
	res := []byte("---")
	if book.Status["profiled"] {
		res[0] = 'p'
	}
	if book.Status["extended-lexicon"] {
		res[1] = 'l'
	}
	if book.Status["post-corrected"] {
		res[2] = 'c'
	}
	return string(res)
}

func (cmd *command) printModel(model api.Model) {
	cmd.info("%s\t%s\n", model.Name, model.Description)
}

func (cmd *command) printModels(models api.Models) {
	for _, model := range models.Models {
		cmd.printModel(model)
	}
}

func (cmd *command) printFreqMap(bid, pid int, freqs map[string]int, label string) {
	for k, v := range freqs {
		cmd.info("%d\t%d\t%s\t%d\t%s\n", bid, pid, k, v, label)
	}
}

func (cmd *command) printExtendedLexicon(el api.ExtendedLexicon) {
	cmd.printFreqMap(el.BookID, el.ProjectID, el.Yes, "yes")
	cmd.printFreqMap(el.BookID, el.ProjectID, el.No, "no")
}

func (cmd *command) printPostCorrection(pc *api.PostCorrection) {
	cmd.printFreqMap(pc.BookID, pc.ProjectID, pc.Always, "always")
	cmd.printFreqMap(pc.BookID, pc.ProjectID, pc.Sometimes, "sometimes")
	cmd.printFreqMap(pc.BookID, pc.ProjectID, pc.Never, "never")
}

func (cmd *command) printCharMap(cm api.CharMap) {
	for k, v := range cm.CharMap {
		cmd.info("%d\t%d\t%s\t%d\n", cm.BookID, cm.ProjectID, k, v)
	}
}

func (cmd *command) printProfile(profile gofiler.Profile) {
	for _, v := range profile {
		top := true
		for _, c := range v.Candidates {
			cmd.info("%s\t%s\t%d\t%f\t%t\n",
				v.OCR, c.Suggestion, c.Distance, c.Weight, top)
			top = false
		}
	}
}

func (cmd *command) printSuggestions(ss api.Suggestions) {
	for _, s := range ss.Suggestions {
		cmd.printSuggestionsArray("", s)
	}
}

func (cmd *command) printSuggestionCounts(counts api.SuggestionCounts) {
	for k, v := range counts.Counts {
		cmd.info("%d\t%d\t%s\t%d\n", counts.BookID, counts.ProjectID, k, v)
	}
}

func (cmd *command) printSuggestionsArray(pre string, suggestions []api.Suggestion) {
	for _, s := range suggestions {
		cmd.info("%s%s\t%s\t%s\t%s\t%s\t%s\t%d\t%f\t%t\n",
			pre, s.Token, s.Suggestion, s.Modern,
			strings.Join(s.HistPatterns, ","), strings.Join(s.OCRPatterns, ","),
			s.Dict, s.Distance, s.Weight, s.Top)
	}
}

func (cmd *command) printPatterns(patterns api.Patterns) {
	for p, v := range patterns.Patterns {
		cmd.printSuggestionsArray(p+"\t", v)
	}
}

func (cmd *command) printPatternCounts(counts api.PatternCounts) {
	for k, v := range counts.Counts {
		cmd.info("%d\t%d\t%s\t%d\t%t\n",
			counts.BookID, counts.ProjectID, k, v, counts.OCR)
	}
}

func (cmd *command) printAdaptiveTokens(at api.AdaptiveTokens) {
	for _, t := range at.AdaptiveTokens {
		cmd.info("%d\t%d\t%s\n", at.BookID, at.ProjectID, t)
	}
}

func (cmd *command) info(format string, args ...interface{}) {
	if cmd.err != nil {
		return
	}
	str := fmt.Sprintf(format, args...)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "\t", " ", -1)
	_, err := fmt.Fprint(cmd.out, str)
	cmd.err = err
}

func getURL() string {
	if pocowebURL != "" {
		return pocowebURL
	}
	return os.Getenv("POCOWEBC_URL")
}

func getAuth() string {
	if authToken != "" {
		return authToken
	}
	return os.Getenv("POCOWEBC_AUTH")
}

func unescape(args ...string) []string {
	res := make([]string, len(args))
	for i := range args {
		u, err := strconv.Unquote(`"` + args[i] + `"`)
		if err != nil {
			res[i] = args[i]
		} else {
			res[i] = u
		}
	}
	return res
}
