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
	startNoWait bool
	startSleepS int
)

var startCommand = cobra.Command{
	Use:   "start",
	Short: "Start various jobs",
}

func init() {
	startCommand.PersistentFlags().BoolVarP(&startNoWait, "nowait", "n", false,
		"do not wait for the job to finish")
	startCommand.PersistentFlags().IntVarP(&startSleepS, "sleep", "s", 5,
		"set the number of seconds to sleep between checks if the job has finished")
}

var startProfileCommand = cobra.Command{
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
	var jobID int
	cmd.do(func() error {
		job, err := cmd.client.PostProfile(id)
		jobID = job.ID
		return err
	})
	// wait
	cmd.do(func() error {
		for !startNoWait {
			time.Sleep(time.Duration(startSleepS) * time.Second)
			status, err := cmd.client.GetJobStatus(jobID)
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
