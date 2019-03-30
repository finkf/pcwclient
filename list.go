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
	Use:   "user ID",
	Short: "List user information",
	Args:  cobra.ExactArgs(1),
	RunE:  doListUser,
}

func doListUser(cmd *cobra.Command, args []string) error {
	return listUser(os.Stdout, args[0])
}

func listUser(out io.Writer, id string) error {
	var uid int
	if err := scanf(id, "%d", &uid); err != nil {
		return fmt.Errorf("invalid user id: %v", err)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		user, err := cmd.client.GetUser(int64(uid))
		cmd.data = user
		return err
	})
	return cmd.output(func() error {
		user := cmd.data.(api.User)
		return infoUser(cmd.out, user)
	})
}

var listUsersCommand = cobra.Command{
	Use:   "users",
	Short: "List user information",
	RunE:  doListUsers,
}

func doListUsers(cmd *cobra.Command, args []string) error {
	return listUsers(os.Stdout)
}

func listUsers(out io.Writer) error {
	cmd := newCommand(out)
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
	Use:   "book ID",
	Short: "List book information",
	Args:  cobra.ExactArgs(1),
	RunE:  doListBook,
}

func doListBook(cmd *cobra.Command, args []string) error {
	return listBook(os.Stdout, args[0])
}

func listBook(out io.Writer, id string) error {
	var bid int
	if err := scanf(id, "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	cmd := newCommand(out)
	cmd.do(func() error {
		book, err := cmd.client.GetBook(bid)
		cmd.data = book
		return err
	})
	return cmd.output(func() error {
		book := cmd.data.(*api.Book)
		return infoBook(cmd.out, *book)
	})
}

var listBooksCommand = cobra.Command{
	Use:   "books",
	Short: "List information about all books",
	RunE:  doListBooks,
}

func doListBooks(cmd *cobra.Command, args []string) error {
	return listBooks(os.Stdout)
}

func listBooks(out io.Writer) error {
	cmd := newCommand(out)
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
