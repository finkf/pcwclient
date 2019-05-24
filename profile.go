package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/finkf/pcwgo/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	profileNoWait bool
	sleepS        int
)

func init() {
	profileCommand.Flags().BoolVarP(&profileNoWait, "nowait", "n", false,
		"do not wait for the profiling to finish")
	profileCommand.Flags().IntVarP(&sleepS, "sleeps", "s", 2,
		"set the number of seconds to sleep between checks")
}

var profileCommand = cobra.Command{
	Use:   "profile ID [QUERY [QUERY...]]",
	Short: "List user information",
	Args:  exactlyNIDs(1),
	RunE:  doProfile,
}

func doProfile(cmd *cobra.Command, args []string) error {
	var bid int
	if err := scanf(args[0], "%d", &bid); err != nil {
		return fmt.Errorf("invalid book id: %v", err)
	}
	return profile(os.Stdout, bid)
}

func profile(out io.Writer, id int) error {
	cmd := newCommand(out)
	// start profiling
	cmd.do(func() error {
		return cmd.client.PostProfile(id)
	})
	// wait
	cmd.do(func() error {
		for !profileNoWait {
			time.Sleep(time.Duration(sleepS) * time.Second)
			status, err := cmd.client.GetJobStatus(id)
			if err != nil {
				return err
			}
			log.Debugf("status: %s", status.StatusName)
			if status.StatusID == db.StatusIDFailed {
				return fmt.Errorf("job %d failed", status.JobID)
			}
			if status.StatusID == db.StatusIDDone {
				return nil
			}
		}
		return nil
	})
	return cmd.print()
}
