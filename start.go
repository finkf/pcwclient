package main

import (
	"fmt"
	"time"

	"github.com/UNO-SOFT/ulog"
	"github.com/finkf/pcwgo/api"
	"github.com/finkf/pcwgo/db"
	"github.com/spf13/cobra"
)

var startArgs = struct {
	nowait bool
	sleep  int
}{}

var startCommand = cobra.Command{
	Use:   "start",
	Short: "Start various jobs",
}

func init() {
	startCommand.PersistentFlags().BoolVarP(&startArgs.nowait, "nowait", "n", false,
		"do not wait for the job to finish")
	startCommand.PersistentFlags().IntVarP(&startArgs.sleep, "sleep", "s", 5,
		"set the number of seconds to sleep between checks if the job has finished")
}

func reattach(c *api.Client, jobID int) (bool, error) {
	if !startArgs.nowait {
		var status api.JobStatus
		if err := get(c, c.URL("jobs/%d", jobID), &status); err != nil {
			return false, fmt.Errorf("reattach to job %d: %v",
				jobID, err)
		}
		if status.StatusID == db.StatusIDRunning {
			return true, nil
		}
	}
	return false, nil
}

func waitForJobToFinish(c *api.Client, jobID int) error {
	for !startArgs.nowait {
		var status api.JobStatus
		if err := get(c, c.URL("jobs/%d", jobID), &status); err != nil {
			return fmt.Errorf("get job status: %v", err)
		}
		if status.StatusID == db.StatusIDFailed {
			return fmt.Errorf("job %d failed", status.JobID)
		}
		if status.StatusID == db.StatusIDDone {
			return nil
		}
		time.Sleep(time.Duration(startArgs.sleep) * time.Second)
	}
	return nil
}

func start(c *api.Client, id int, fn func() error) error {
	re, err := reattach(c, id)
	ulog.Write("start", "re", re, "err", err)
	if err != nil {
		return err
	}
	if !re {
		if err := fn(); err != nil {
			ulog.Write("fn()", "err", err)
			return err
		}
	}
	ulog.Write("waitForJobToFinish", "id", id)
	return waitForJobToFinish(c, id)
}

var startProfileCommand = cobra.Command{
	Use:   "profile ID [ALEX-TOKENS...]",
	Short: "Start to profile book ID",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doProfile,
}

func doProfile(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("start profile: invalid book ID: %q", args[0])
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	jobID := bid
	err := start(c, jobID, func() error {
		var job api.Job
		return post(c, c.URL("/profile/books/%d", bid), nil, &job)
	})
	if err != nil {
		return fmt.Errorf("start profile book %d: %v", bid, err)
	}
	return nil
}

var startELCommand = cobra.Command{
	Use:   "el ID",
	Short: "Create extended lexicon for book ID",
	Args:  cobra.ExactArgs(1),
	RunE:  doEL,
}

func doEL(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("start el: invalid book ID: %q",
			args[0])
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	jobID := bid
	err := start(c, jobID, func() error {
		var job api.Job
		return post(c, c.URL("postcorrect/le/books/%d", bid), nil, &job)
	})
	if err != nil {
		return fmt.Errorf("start el for book %d: %v",
			bid, err)
	}
	return nil
}

var startRRDMCommand = cobra.Command{
	Use:   "rrdm ID",
	Short: "Start automatic post-correction on book ID",
	Args:  cobra.ExactArgs(1),
	RunE:  doRRDM,
}

func doRRDM(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("start rrdm: invalid book ID: %q", args[0])
	}
	c := api.Authenticate(getURL(), getAuth(), mainArgs.skipVerify)
	jobID := bid
	err := start(c, jobID, func() error {
		var job api.Job
		return post(c, c.URL("postcorrect/books/%d", bid), nil, &job)
	})
	if err != nil {
		return fmt.Errorf("start rrdm for book %d: %v", bid, err)
	}
	return nil
}
