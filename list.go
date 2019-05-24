package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var histPatterns bool

func init() {
	listPatternsCommand.Flags().BoolVarP(&histPatterns, "hist", "H", false,
		"list historical rewrite patterns")
}

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

var listPatternsCommand = cobra.Command{
	Use:   "patterns ID [QUERY [QUERY...]]",
	Short: "List patterns for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListPatterns,
}

func doListPatterns(cmd *cobra.Command, args []string) error {
	var bid int
	if err := scanf(args[0], "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	switch len(args) {
	case 1:
		return listAllPatterns(os.Stdout, bid)
	case 2:
		return listPatterns(os.Stdout, bid, args[1])
	default:
		return listPatterns(os.Stdout, bid, args[1], args[2:]...)
	}
}

func listAllPatterns(out io.Writer, id int) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		ps, err := cmd.client.GetPatterns(id, !histPatterns)
		cmd.add(ps)
		return err
	})
	return cmd.print()
}

func listPatterns(out io.Writer, id int, q string, qs ...string) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		ps, err := cmd.client.QueryPatterns(id, !histPatterns, q, qs...)
		cmd.add(ps)
		return err
	})
	return cmd.print()
}

var listSuggestionsCommand = cobra.Command{
	Use:   "suggestions ID [QUERY [QUERY...]]",
	Short: "List profiler suggestions for the given book",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doListSuggestions,
}

func doListSuggestions(cmd *cobra.Command, args []string) error {
	var bid int
	if err := scanf(args[0], "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	switch len(args) {
	case 1:
		return listAllSuggestions(os.Stdout, bid)
	case 2:
		return listSuggestions(os.Stdout, bid, args[1])
	default:
		return listSuggestions(os.Stdout, bid, args[1], args[2:]...)
	}
}

func listAllSuggestions(out io.Writer, id int) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		profile, err := cmd.client.GetProfile(id)
		cmd.add(profile)
		return err
	})
	return cmd.print()
}

func listSuggestions(out io.Writer, id int, q string, qs ...string) error {
	cmd := newCommand(out)
	cmd.do(func() error {
		suggestions, err := cmd.client.QueryProfile(id, q, qs...)
		cmd.add(suggestions)
		return err
	})
	return cmd.print()
}

var listSuspiciousCommand = cobra.Command{
	Use:   "suspicious ID",
	Short: "List suspicous words for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListSuspicious,
}

func doListSuspicious(cmd *cobra.Command, args []string) error {
	var bid int
	if err := scanf(args[0], "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	cmdx := newCommand(os.Stdout)
	cmdx.do(func() error {
		counts, err := cmdx.client.GetSuspicious(bid)
		cmdx.add(counts)
		return err
	})
	return cmdx.print()
}

var listAdaptiveCommand = cobra.Command{
	Use:   "adaptive ID",
	Short: "List adaptive tokens for the given book",
	Args:  exactlyNIDs(1),
	RunE:  doListAdaptive,
}

func doListAdaptive(cmd *cobra.Command, args []string) error {
	var bid int
	if err := scanf(args[0], "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	cmdx := newCommand(os.Stdout)
	cmdx.do(func() error {
		at, err := cmdx.client.GetAdaptiveTokens(bid)
		cmdx.add(at)
		return err
	})
	return cmdx.print()
}
