package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/UNO-SOFT/ulog"
	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func init() {
	newBookCommand.Flags().StringVarP(&newBookArgs.author, "author", "a", "",
		"set book's author (required)")
	newBookCommand.Flags().StringVarP(&newBookArgs.title, "title", "t", "",
		"set book's title (required)")
	newBookCommand.Flags().StringVarP(&newBookArgs.description,
		"description", "d", "", "set book's description")
	newBookCommand.Flags().StringVarP(&newBookArgs.language, "language", "l", "",
		"set book's language")
	newBookCommand.Flags().StringVarP(&newBookArgs.profilerURL, "profilerurl", "u",
		"local", "set book's profiler url")
	newBookCommand.Flags().IntVarP(&newBookArgs.year, "year", "y", 1900,
		"set book's year")
	newBookCommand.Flags().StringVarP(&newBookArgs.histPatterns, "patters", "p", "",
		"set additional historical patterns for the book")
	_ = cobra.MarkFlagRequired(newBookCommand.Flags(), "author")
	_ = cobra.MarkFlagRequired(newBookCommand.Flags(), "title")
	_ = cobra.MarkFlagRequired(newBookCommand.Flags(), "language")
	newUserCommand.Flags().StringVarP(&newUserArgs.name, "name", "n", "",
		"set the user's name (required)")
	newUserCommand.Flags().StringVarP(&newUserArgs.email, "email", "e", "",
		"set the user's name (required)")
	newUserCommand.Flags().StringVarP(&newUserArgs.password, "password", "p",
		"", "set the user's password (required)")
	newUserCommand.Flags().StringVarP(&newUserArgs.institute, "institute",
		"i", "", "set the user's institute")
	newUserCommand.Flags().BoolVarP(&newUserArgs.admin, "admin", "a", false,
		"user has administrator permissions")
	_ = cobra.MarkFlagRequired(newUserCommand.Flags(), "name")
	_ = cobra.MarkFlagRequired(newUserCommand.Flags(), "email")
	_ = cobra.MarkFlagRequired(newUserCommand.Flags(), "password")
}

var newCommand = cobra.Command{
	Use:   "new",
	Short: "Create new books and users",
}

var newBookCommand = cobra.Command{
	Use:   "book [ZIP|DIR]",
	Short: "Create a new book",
	RunE:  newBook,
	Args:  cobra.ExactArgs(1),
}

var newBookArgs = struct {
	author       string
	title        string
	description  string
	language     string
	profilerURL  string
	histPatterns string
	year         int
}{}

func newBook(_ *cobra.Command, args []string) error {
	zip, err := openAsZIP(args[0])
	if err != nil {
		return fmt.Errorf("cannot create new book: open %s: %v", args[0], err)
	}
	defer zip.Close()
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	url := newBookURL(c)
	req, err := http.NewRequest(http.MethodPost, url, zip)
	if err != nil {
		return fmt.Errorf("cannot create new book: %v", err)
	}
	req.Header.Add("Content-Type", "application/zip")
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("cannot create new book: %v", err)
	}
	var book api.Book
	if err := api.UnmarshalResponse(res, &book); err != nil {
		return fmt.Errorf("cannot create new book: %v", err)
	}
	format(&book)
	return nil
}

func openAsZIP(p string) (io.ReadCloser, error) {
	fi, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return os.Open(p)
	}
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	prefix := len(path.Dir(p))
	if prefix > 0 { // increment prefix to include the slash if non empty prefix
		prefix++
	}
	err = filepath.Walk(p, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, e := zip.FileInfoHeader(fi)
		if e != nil {
			return e
		}
		internalPath := p[prefix:]
		if fi.IsDir() {
			internalPath += "/"
			header.Name = internalPath
			_, e := w.CreateHeader(header)
			ulog.Write("filepath.Walk", "path", p,
				"internal", internalPath, "prefix", prefix)
			return e
		}
		// copy file
		ulog.Write("filepath.Walk", "path", p,
			"internal", internalPath, "prefix", prefix)
		header.Method = zip.Deflate
		// open file
		in, e := os.Open(p)
		if e != nil {
			return e
		}
		defer in.Close()
		out, e := w.CreateHeader(header)
		if e != nil {
			return e
		}
		// write to archive
		_, e = io.Copy(out, in)
		return e
	})
	w.Close()
	if err := ioutil.WriteFile("/tmp/pocowebc.zip", buf.Bytes(), 0666); err != nil {
		return nil, err
	}
	return ioutil.NopCloser(&buf), err
}

func newBookURL(c *api.Client) string {
	return c.URL("books?author=%s&title=%s&language=%s"+
		"&description=%s&histPatterns=%s&profilerUrl=%s&year=%d",
		url.QueryEscape(newBookArgs.author),
		url.QueryEscape(newBookArgs.title),
		url.QueryEscape(newBookArgs.language),
		url.QueryEscape(newBookArgs.description),
		url.QueryEscape(newBookArgs.histPatterns),
		url.QueryEscape(newBookArgs.profilerURL),
		newBookArgs.year)
}

var newUserArgs = struct {
	email, password, institute, name string
	admin                            bool
}{}

var newUserCommand = cobra.Command{
	Use:   "user",
	Short: "Create a new user",
	RunE:  newUser,
}

func newUser(cmd *cobra.Command, args []string) error {
	if newUserArgs.email == "" || newUserArgs.password == "" {
		return fmt.Errorf("missing user email and/or password")
	}
	var newUser api.User
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	err := post(c, c.URL("users"), api.CreateUserRequest{
		User: api.User{
			Name:      newUserArgs.name,
			Email:     newUserArgs.email,
			Institute: newUserArgs.institute,
			Admin:     newUserArgs.admin,
		},
		Password: newUserArgs.password,
	}, &newUser)
	if err != nil {
		return fmt.Errorf("cannot create user %s: %v", newUserArgs.email, err)
	}
	format(&newUser)
	return nil
}
