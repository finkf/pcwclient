package main

import (
	"fmt"
	"io"
	"os"

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
		cmd.add(user)
		return err
	})
	return cmd.print()
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
		users, err := cmd.client.GetUsers()
		cmd.add(users)
		return err
	})
	return cmd.print()
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
		cmd.add(book)
		return err
	})
	return cmd.print()
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
		cmd.add(books)
		return err
	})
	return cmd.print()
}
