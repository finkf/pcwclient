package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var listCommand = cobra.Command{
	Use:   "list",
	Short: "List various informations",
}

var listUserCommand = cobra.Command{
	Use:   "user",
	Short: "List user information",
	RunE:  doListUser,
}

func doListUser(cmd *cobra.Command, args []string) error {
	return listUser(os.Stdout)
}

func listUser(out io.Writer) error {
	cmd := newCommand(out)
	if userID != 0 {
		cmd.do(func() error {
			user, err := cmd.client.GetUser(int64(userID))
			cmd.data = user
			return err
		})
		return cmd.output(func() error {
			user := cmd.data.(api.User)
			return infoUser(cmd.out, user)
		})
	}
	cmd.do(func() error {
		user, err := cmd.client.GetUsers()
		cmd.data = user
		return err
	})
	return cmd.output(func() error {
		users := cmd.data.(api.Users)
		for _, user := range users.Users {
			if err := infoUser(cmd.out, user); err != nil {
				return err
			}
		}
		return nil
	})
}

var listBookCommand = cobra.Command{
	Use:   "book",
	Short: "List book information",
	RunE:  doListBook,
}

func doListBook(cmd *cobra.Command, args []string) error {
	return listBook(os.Stdout)
}

func listBook(out io.Writer) error {
	cmd := newCommand(out)
	if bookID != 0 {
		cmd.do(func() error {
			book, err := cmd.client.GetBook(bookID)
			cmd.data = book
			return err
		})
		return cmd.output(func() error {
			book := cmd.data.(*api.Book)
			return infoBook(cmd.out, *book)
		})
	}
	cmd.do(func() error {
		books, err := cmd.client.GetBooks()
		cmd.data = books
		return err
	})
	return cmd.output(func() error {
		books := cmd.data.(*api.Books)
		for _, book := range books.Books {
			if err := infoBook(cmd.out, book); err != nil {
				return err
			}
		}
		return nil
	})
}

func infoUser(out io.Writer, user api.User) error {
	return info(out, "%d\t%s\t%s\t%s\t%t\n",
		user.ID, user.Name, user.Email, user.Institute, user.Admin)
}

func infoBook(out io.Writer, book api.Book) error {
	return info(out, "%d\t%s\t%s\t%s\t%d\t%s\t%s\t%t\n",
		book.ProjectID, book.Author, book.Title, book.Description,
		book.Year, book.Language, book.ProfilerURL, book.IsBook)
}

func info(out io.Writer, format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "\t", " ", -1)
	_, err := fmt.Fprint(out, str)
	return err
}
