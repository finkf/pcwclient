package main

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createBookCommand.Flags().StringVarP(&bookAuthor, "author", "A", "", "set book's author (required)")
	createBookCommand.Flags().StringVarP(&bookTitle, "title", "T", "", "set book's title (required)")
	createBookCommand.Flags().StringVarP(&bookDescription, "description", "D", "", "set book's description")
	createBookCommand.Flags().StringVarP(&bookLanguage, "language", "L", "", "set book's language")
	createBookCommand.Flags().StringVarP(&bookProfilerURL, "profilerurl", "P", "local", "set book's profiler url")
	createBookCommand.Flags().IntVarP(&bookYear, "year", "Y", 1900, "set book's year")
	cobra.MarkFlagRequired(createBookCommand.Flags(), "author")
	cobra.MarkFlagRequired(createBookCommand.Flags(), "title")
	createUserCommand.Flags().StringVarP(&userName, "name", "N", "", "set the user's name (required)")
	createUserCommand.Flags().StringVarP(&userEmail, "email", "E", "", "set the user's name (required)")
	createUserCommand.Flags().StringVarP(&userPassword, "password", "P", "", "set the user's password (required)")
	createUserCommand.Flags().StringVarP(&userInstitute, "institute", "I", "", "set the user's institute")
	createUserCommand.Flags().BoolVarP(&userAdmin, "admin", "R", false, "user has administrator permissions")
	cobra.MarkFlagRequired(createUserCommand.Flags(), "name")
	cobra.MarkFlagRequired(createUserCommand.Flags(), "email")
	cobra.MarkFlagRequired(createUserCommand.Flags(), "password")
}

var createCommand = cobra.Command{
	Use:   "create",
	Short: "Create various things",
}

var createBookCommand = cobra.Command{
	Use:   "book",
	Short: "Create a new book",
	RunE:  doCreateBook,
	Args:  cobra.ExactArgs(1),
}

var (
	bookAuthor      = ""
	bookTitle       = ""
	bookDescription = ""
	bookLanguage    = ""
	bookProfilerURL = ""
	bookYear        = 0
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
	var zip io.ReadCloser
	var err error
	cmd := newCommand(out)
	defer func() {
		if zip != nil {
			zip.Close()
		}
	}()
	cmd.do(func() error {
		zip, err = openAsZIP(p)
		return err
	})
	cmd.do(func() error {
		book, err := cmd.client.PostZIP(zip)
		cmd.data = book
		return err
	})
	cmd.do(func() error {
		newBook := api.Book{
			ProjectID:   cmd.data.(api.Book).ProjectID,
			Title:       bookTitle,
			Author:      bookAuthor,
			Description: bookDescription,
			Year:        bookYear,
			Language:    bookLanguage,
			ProfilerURL: bookProfilerURL,
		}
		book, err := cmd.client.PostBook(newBook)
		cmd.data = book
		return err
	})
	return cmd.output(func() error {
		return infoBook(out, *cmd.data.(*api.Book))
	})
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
			log.Printf("filepath.Walk: %s (%s - %d)", p, internalPath, prefix)
			return e
		}
		// copy file
		log.Printf("filepath.Walk: %s (%s - %d)", p, internalPath, prefix)
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
	ioutil.WriteFile("/tmp/pocowebc.zip", buf.Bytes(), 0666)
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
	cmd := newCommand(out)
	cmd.do(func() error {
		req := api.CreateUserRequest{
			User: api.User{
				Name:      userName,
				Email:     userEmail,
				Institute: userInstitute,
				Admin:     userAdmin,
			},
			Password: userPassword,
		}
		u, err := cmd.client.PostUser(req)
		cmd.data = u
		return err
	})
	return cmd.output(func() error {
		return infoUser(out, cmd.data.(api.User))
	})
}
