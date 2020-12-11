package main

import (
	"fmt"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var deleteCommand = cobra.Command{
	Use:   "delete",
	Short: "Delete users or books",
}

var deleteBooksCommand = cobra.Command{
	Use:   "books IDS...",
	Short: "Delete a books, pages or lines",
	Args:  cobra.MinimumNArgs(1),
	RunE:  deleteBooks,
}

func deleteBooks(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for _, id := range args {
		var bid, pid, lid int
		var url string
		switch n := parseIDs(id, &bid, &pid, &lid); n {
		case 3:
			url = c.URL("books/%d/pages/%d/lines/%d", bid, pid, lid)
		case 2:
			url = c.URL("books/%d/pages/%d", bid, pid)
		case 1:
			url = c.URL("books/%d", bid)
		default:
			return fmt.Errorf("delete book: invalid id: %q", id)
		}
		if err := delete(c, url, nil); err != nil {
			return fmt.Errorf("delete book %s: %v", id, err)
		}
	}
	return nil
}

var deleteUsersCommand = cobra.Command{
	Use:   "users IDS...",
	Short: "delete users",
	Args:  cobra.MinimumNArgs(1),
	RunE:  deleteUsers,
}

func deleteUsers(_ *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	for _, id := range args {
		var uid int
		if n := parseIDs(id, &uid); n != 1 {
			return fmt.Errorf("delete user: invalid user id: %s", id)
		}
		if err := delete(c, c.URL("users/%d", uid), nil); err != nil {
			return fmt.Errorf("delete user: %v", err)
		}
	}
	return nil
}
