package main

import (
	"fmt"
	"strconv"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func init() {
	pkgSplitCommand.Flags().BoolVarP(&pkgSplitArgs.random, "random", "r",
		false, "create random packages")
}

var pkgCommand = cobra.Command{
	Use:   "pkg",
	Short: "Assign and reassign packages.",
}

var pkgAssignCommand = cobra.Command{
	Use:   "assign ID [USERID]",
	Short: "Assign the package ID to the user USERID",
	RunE:  doAssign,
	Args:  exactArgs(1, 2),
	Long: `
Assign the package ID to a user.  If USERID is omitted, the package is
assigned back to its original owner.  Otherwise it is assigned to the
user with the given USERID.`,
}

func doAssign(_ *cobra.Command, args []string) error {
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("cannot assign package: invalid id: %s", arg)
		}
		ids = append(ids, id)
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var err error
	switch len(ids) {
	case 2:
		err = get(c, c.URL("pkg/assign/books/%d?assignto=%d", ids[0], ids[1]), nil)
	default:
		err = get(c, c.URL("pkg/assign/books/%d", ids[0]), nil)
	}
	if err != nil {
		return fmt.Errorf("cannot assign package %d: %v", ids[0], err)
	}
	return nil
}

var pkgReassignCommand = cobra.Command{
	Use:   "reassign ID",
	Short: "Reassign packages of book ID to its owner",
	RunE:  doReassign,
	Args:  cobra.ExactArgs(1),
	Long: `
Reassign all packages of the book ID that are owned by different users
are to the owner of the project.`,
}

func doReassign(cmd *cobra.Command, args []string) error {
	var pid int
	if n := parseIDs(args[0], &pid); n != 1 {
		return fmt.Errorf("cannot reassign: invalid id: %s", args[0])
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	if err := get(c, c.URL("pkg/takeback/books/%d", pid), nil); err != nil {
		return fmt.Errorf("cannot reassign package %d: %v", pid, err)
	}
	return nil
}

var pkgSplitArgs = struct {
	random bool
}{}

var pkgSplitCommand = cobra.Command{
	Use:   "split ID USERID [USERID...]",
	Short: "Split the project ID into multiple packages",
	RunE:  doSplit,
	Args:  cobra.MinimumNArgs(2),
	Long: `
Split the project ID into multiple packages.  The project is split
into N packages where N is the number of given USERIDs.  Each project
is assigned to the given users in order.

E.g. "pocowebc new pkgs 13 1 2 3" splits the project 13 into 3
packages.  The first package is owned by user 1, the second by user 2
and the third by user 3.`,
}

func doSplit(cmd *cobra.Command, args []string) error {
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("cannot split: invalid id: %s", arg)
		}
		ids = append(ids, id)
	}
	url := c.URL("pkg/split/books/%d", ids[0])
	err := post(c, url, api.SplitRequest{
		UserIDs: ids[1:],
		Random:  pkgSplitArgs.random,
	}, nil)
	if err != nil {
		return fmt.Errorf("cannot split %d: %v", ids[0], err)
	}
	return nil
}
