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

	"github.com/finkf/gofiler"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// various command line flags
var (
	debug      = false
	skipVerify = false
	authToken  = ""
	pocowebURL = ""
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
	mainCommand.AddCommand(&pkgCommand)
	mainCommand.AddCommand(&poolCommand)
	pkgCommand.AddCommand(&assignCommand)
	pkgCommand.AddCommand(&takeBackCommand)
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
	listCommand.AddCommand(&listCharsCommand)
	createCommand.AddCommand(&createUserCommand)
	createCommand.AddCommand(&createBookCommand)
	createCommand.AddCommand(&createPackagesCommand)
	startCommand.AddCommand(&startProfileCommand)
	startCommand.AddCommand(&startELCommand)
	startCommand.AddCommand(&startRRDMCommand)
	deleteCommand.AddCommand(&deleteBooksCommand)
	deleteCommand.AddCommand(&deleteUsersCommand)
	poolCommand.AddCommand(&downloadPoolCommand)
	poolCommand.AddCommand(&runPoolCommand)
	mainCommand.AddCommand(&snippetCommand)
	snippetCommand.AddCommand(&snippetGetCommand)
	snippetCommand.AddCommand(&snippetPutCommand)

	mainCommand.SilenceUsage = true
	mainCommand.SilenceErrors = true
	mainCommand.PersistentFlags().BoolVarP(&formatJSON, "json", "J", false,
		"output raw json")
	mainCommand.PersistentFlags().BoolVarP(&skipVerify, "skip-verify", "S", false,
		"ignore invalid ssl certificates")
	mainCommand.PersistentFlags().BoolVarP(&debug, "debug", "D", false,
		"enable debug output")
	mainCommand.PersistentFlags().StringVarP(&pocowebURL, "url", "U",
		getURL(), "set pocoweb url (env: PCWCLIENT_URL)")
	mainCommand.PersistentFlags().StringVarP(&formatTemplate, "format", "F",
		"", "set output format")
	mainCommand.PersistentFlags().StringVarP(&authToken, "auth", "A",
		getAuth(), "set auth token (env: PCWCLIENT_AUTH)")
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

type client struct {
	client *api.Client
	data   []interface{}
	out    io.Writer
	err    error
}

func newClient(out io.Writer) *client {
	url, auth := getURL(), getAuth()
	if url == "" || auth == "" {
		return &client{
			err: fmt.Errorf("missing login information: see login sub command"),
		}
	}
	log.Debugf("authenticating at %s with %s", url, auth)
	return &client{client: api.Authenticate(url, auth, skipVerify), out: out}
}

func (c *client) do(f func(*api.Client) (interface{}, error)) {
	if c.err != nil {
		return
	}
	val, err := f(c.client)
	// c.print handles the possible error
	c.err = err
	c.print(val)
}

func (c *client) done() error {
	log.Debugf("done")
	if c.err != nil {
		return c.err
	}
	if formatJSON {
		return c.printJSON()
	}
	if formatTemplate != "" {
		return c.printTemplate(formatTemplate)
	}
	return nil
}

func (c *client) print(val interface{}) {
	if c.err != nil {
		return
	}
	// cache printout data for json or format output
	if formatJSON || formatTemplate != "" {
		c.data = append(c.data, val)
		return
	}
	if val != nil {
		c.printVal(val)
	}
}

func (c *client) printVal(val interface{}) {
	if c.err != nil {
		return
	}
	// just print the given value
	switch t := val.(type) {
	case []interface{}:
		c.printArray(t)
	case api.Session:
		c.printSession(t)
	case api.Version:
		c.printVersion(t)
	case *api.Book:
		c.printBook(t)
	case *api.Books:
		c.printBooks(t)
	case api.SplitPackages:
		c.printSplitPackages(t)
	case api.User:
		c.printUser(t)
	case api.Users:
		c.printUsers(t)
	case gofiler.Profile:
		c.printProfile(t)
	case api.Suggestions:
		c.printSuggestions(t)
	case api.SuggestionCounts:
		c.printSuggestionCounts(t)
	case api.Patterns:
		c.printPatterns(t)
	case api.PatternCounts:
		c.printPatternCounts(t)
	case api.AdaptiveTokens:
		c.printAdaptiveTokens(t)
	case api.Models:
		c.printModels(t)
	case api.ExtendedLexicon:
		c.printExtendedLexicon(t)
	case *api.PostCorrection:
		c.printPostCorrection(t)
	case api.CharMap:
		c.printCharMap(t)
	default:
		log.Fatalf("error: invalid type to print: %T", val)
	}
}

func (c *client) printJSON() error {
	var data interface{} = c.data
	if len(c.data) == 1 {
		data = c.data[0]
	}
	if err := json.NewEncoder(c.out).Encode(data); err != nil {
		return fmt.Errorf("error encoding to json: %v", err)
	}
	return nil
}

func (c *client) printTemplate(tmpl string) error {
	var data interface{} = c.data
	if len(c.data) == 1 {
		data = c.data[0]
	}
	t, err := template.New("pocwebc").Parse(
		strings.Replace(tmpl, "\\n", "\n", -1))
	if err != nil {
		return fmt.Errorf("invalid format string: %v", err)
	}
	if err = t.Execute(c.out, data); err != nil {
		return fmt.Errorf("error formatting string: %v", err)
	}
	return nil
}

func (c *client) println(args ...interface{}) {
	if c.err != nil {
		return
	}
	if _, err := fmt.Println(args...); err != nil {
		c.err = err
	}
}

func (c *client) printArray(xs []interface{}) {
	for _, x := range xs {
		c.printVal(x)
	}
}

func (c *client) printSession(s api.Session) {
	c.info("%d\t%s\t%s\t%s\t%s\n",
		s.User.ID, s.User.Email, s.User.Name, s.Auth,
		time.Unix(s.Expires, 0).Format(time.RFC3339))
}

func (c *client) printVersion(v api.Version) {
	c.println(v.Version)
}

func (c *client) printUsers(users api.Users) {
	for _, user := range users.Users {
		c.printUser(user)
	}
}

func (c *client) printUser(user api.User) {
	c.info("%d\t%s\t%s\t%s\t%t\n",
		user.ID, user.Name, user.Email, user.Institute, user.Admin)
}

func (c *client) printBooks(books *api.Books) {
	for _, book := range books.Books {
		c.printBook(&book)
	}
}

func (c *client) printSplitPackages(packages api.SplitPackages) {
	for _, pkg := range packages.Packages {
		c.info("%d\t%d\t%d\t%d\n",
			packages.BookID, pkg.ProjectID, pkg.Owner, len(pkg.PageIDs))
	}
}

func (c *client) printBook(book *api.Book) {
	c.info("%d\t%d\t%s\t%s\t%d\t%s\t%s\t%d\t%s\t%s\t%t\n",
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
		res[1] = 'e'
	}
	if book.Status["post-corrected"] {
		res[2] = 'c'
	}
	return string(res)
}

func (c *client) printModel(model api.Model) {
	c.info("%s\t%s\n", model.Name, model.Description)
}

func (c *client) printModels(models api.Models) {
	for _, model := range models.Models {
		c.printModel(model)
	}
}

func (c *client) printFreqMap(bid, pid int, freqs map[string]int, label string) {
	for k, v := range freqs {
		c.info("%d\t%d\t%s\t%d\t%s\n", bid, pid, k, v, label)
	}
}

func (c *client) printExtendedLexicon(el api.ExtendedLexicon) {
	c.printFreqMap(el.BookID, el.ProjectID, el.Yes, "yes")
	c.printFreqMap(el.BookID, el.ProjectID, el.No, "no")
}

func (c *client) printPostCorrection(pc *api.PostCorrection) {
	for k, v := range pc.Corrections {
		c.info("%s\t%s\t%s\t%f\t%t\n", k, v.OCR, v.Cor, v.Confidence, v.Taken)
	}
}

func (c *client) printCharMap(cm api.CharMap) {
	for k, v := range cm.CharMap {
		c.info("%d\t%d\t%s\t%d\n", cm.BookID, cm.ProjectID, k, v)
	}
}

func (c *client) printProfile(profile gofiler.Profile) {
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
		for _, cand := range v.Candidates {
			c.printSuggestion("", api.Suggestion{
				Token:        k,
				Suggestion:   cand.Suggestion,
				Modern:       cand.Modern,
				Distance:     cand.Distance,
				Weight:       float64(cand.Weight),
				HistPatterns: pats(cand.HistPatterns),
				OCRPatterns:  pats(cand.OCRPatterns),
				Top:          top,
			})
			// c.info("%s\t%s\t%d\t%f\t%t\n",
			// 	v.OCR, c.Suggestion, c.Distance, c.Weight, top)
			top = false
		}
	}
}

func (c *client) printSuggestions(ss api.Suggestions) {
	for _, s := range ss.Suggestions {
		c.printSuggestionsArray("", s)
	}
}

func (c *client) printSuggestionCounts(counts api.SuggestionCounts) {
	for k, v := range counts.Counts {
		c.info("%d\t%d\t%s\t%d\n", counts.BookID, counts.ProjectID, k, v)
	}
}

func (c *client) printSuggestionsArray(pre string, suggestions []api.Suggestion) {
	for _, s := range suggestions {
		c.printSuggestion("", s)
	}
}

func (c *client) printSuggestion(pre string, s api.Suggestion) {
	c.info("%s%s\t%s\t%s\t%s\t%s\t%s\t%d\t%f\t%t\n",
		pre, s.Token, s.Suggestion, s.Modern,
		strings.Join(s.HistPatterns, ","), strings.Join(s.OCRPatterns, ","),
		s.Dict, s.Distance, s.Weight, s.Top)
}

func (c *client) printPatterns(patterns api.Patterns) {
	for p, v := range patterns.Patterns {
		c.printSuggestionsArray(p+"\t", v)
	}
}

func (c *client) printPatternCounts(counts api.PatternCounts) {
	for k, v := range counts.Counts {
		c.info("%d\t%d\t%s\t%d\t%t\n",
			counts.BookID, counts.ProjectID, k, v, counts.OCR)
	}
}

func (c *client) printAdaptiveTokens(at api.AdaptiveTokens) {
	for _, t := range at.AdaptiveTokens {
		c.info("%d\t%d\t%s\n", at.BookID, at.ProjectID, t)
	}
}

func (c *client) info(format string, args ...interface{}) {
	if c.err != nil {
		return
	}
	str := fmt.Sprintf(format, args...)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "\t", " ", -1)
	_, err := fmt.Fprint(c.out, str)
	c.err = err
}

func getURL() string {
	if pocowebURL != "" {
		return pocowebURL
	}
	return os.Getenv("PCWCLIENT_URL")
}

func getAuth() string {
	if authToken != "" {
		return authToken
	}
	return os.Getenv("PCWCLIENT_AUTH")
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
