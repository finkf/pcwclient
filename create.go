package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createBookCommand.Flags().StringVarP(&bookAuthor, "author", "a", "",
		"set book's author (required)")
	createBookCommand.Flags().StringVarP(&bookTitle, "title", "t", "",
		"set book's title (required)")
	createBookCommand.Flags().StringVarP(&bookDescription, "description", "d", "",
		"set book's description")
	createBookCommand.Flags().StringVarP(&bookLanguage, "language", "l", "",
		"set book's language")
	createBookCommand.Flags().StringVarP(&bookProfilerURL, "profilerurl", "p", "local",
		"set book's profiler url")
	createBookCommand.Flags().IntVarP(&bookYear, "year", "y", 1900,
		"set book's year")
	_ = cobra.MarkFlagRequired(createBookCommand.Flags(), "author")
	_ = cobra.MarkFlagRequired(createBookCommand.Flags(), "title")
	createUserCommand.Flags().StringVarP(&userName, "name", "n", "",
		"set the user's name (required)")
	createUserCommand.Flags().StringVarP(&userEmail, "email", "e", "",
		"set the user's name (required)")
	createUserCommand.Flags().StringVarP(&userPassword, "password", "p", "",
		"set the user's password (required)")
	createUserCommand.Flags().StringVarP(&userInstitute, "institute", "i", "",
		"set the user's institute")
	createUserCommand.Flags().BoolVarP(&userAdmin, "admin", "a", false,
		"user has administrator permissions")
	_ = cobra.MarkFlagRequired(createUserCommand.Flags(), "name")
	_ = cobra.MarkFlagRequired(createUserCommand.Flags(), "email")
	_ = cobra.MarkFlagRequired(createUserCommand.Flags(), "password")
	createPackagesCommand.Flags().BoolVarP(&splitRandom, "random", "r",
		false, "create random packages")
}

var createCommand = cobra.Command{
	Use:   "new",
	Short: "Create books and users",
}

var createBookCommand = cobra.Command{
	Use:   "book [ZIP|DIR]",
	Short: "Create a new book",
	RunE:  doCreateBook,
	Args:  cobra.ExactArgs(1),
}

var (
	bookAuthor      = ""
	bookTitle       = ""
	bookDescription = ""
	bookLanguage    = ""
	bookProfilerURL = "local"
	bookYear        = 1900
	userName        = ""
	userEmail       = ""
	userInstitute   = ""
	userPassword    = ""
	userAdmin       = false
)

func doCreateBook(cmd *cobra.Command, args []string) error {
	return createBook(args[0], os.Stdout)
}

func createBook(p string, out io.Writer) error {
	if bookAuthor == "" || bookTitle == "" || bookLanguage == "" {
		return fmt.Errorf("missing title, author and/or language")
	}
	zip, err := openAsZIP(p)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", p, err)
	}
	defer func() {
		zip.Close()
	}()
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.PostBook(zip, api.Book{
			Title:       bookTitle,
			Author:      bookAuthor,
			Description: bookDescription,
			Year:        bookYear,
			Language:    bookLanguage,
			ProfilerURL: bookProfilerURL,
		})
	})
	return c.done()
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
			log.Debugf("filepath.Walk: %s (%s - %d)", p, internalPath, prefix)
			return e
		}
		// copy file
		log.Debugf("filepath.Walk: %s (%s - %d)", p, internalPath, prefix)
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

var createUserCommand = cobra.Command{
	Use:   "user",
	Short: "Create a new user",
	RunE:  doCreateUser,
}

func doCreateUser(cmd *cobra.Command, args []string) error {
	return createUser(os.Stdout)
}

func createUser(out io.Writer) error {
	if userEmail == "" || userPassword == "" {
		return fmt.Errorf("missing user email and/or password")
	}
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return client.PostUser(api.CreateUserRequest{
			User: api.User{
				Name:      userName,
				Email:     userEmail,
				Institute: userInstitute,
				Admin:     userAdmin,
			},
			Password: userPassword,
		})
	})
	return c.done()
}

var createPackagesCommand = cobra.Command{
	Use:   "pkgs ID USERID [USERID...]",
	Short: "Split the project ID into multiple packages",
	Long: `
Split the project ID into multiple packages.  The project is split
into N packages where N is the number of given USERIDs.  Each project
is assigned to the given users in order.

E.g. "pocowebc new pkgs 13 1 2 3" splits the project 13 into 3
packages.  The first package is owned by user 1, the second by user 2
and the third by user 3.`,
	RunE: doSplit,
	Args: cobra.MinimumNArgs(2),
}

func doSplit(cmd *cobra.Command, args []string) error {
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("split: invalid id: %s", arg)
		}
		ids = append(ids, id)
	}
	if err := split(os.Stdout, ids[0], ids[1:]); err != nil {
		return fmt.Errorf("split: %v", err)
	}
	return nil
}

func split(out io.Writer, bid int, userids []int) error {
	c := newClient(out)
	c.do(func(client *api.Client) (interface{}, error) {
		return c.client.Split(bid, splitRandom, userids[0], userids[1:]...)
	})
	return c.done()
}
