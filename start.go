package main

import (
	"fmt"
	"os"
	"time"

	"github.com/finkf/pcwgo/api"
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

func reattach(c *client, jobID int) bool {
	var res bool
	if !startNoWait {
		c.do(func(client *api.Client) (interface{}, error) {
			status, err := c.client.GetJobStatus(jobID)
			if err != nil {
				return nil, err
			}
			if status.StatusID == db.StatusIDRunning {
				res = true
				return nil, nil
			}
			return nil, nil
		})
	}
	return res
}

func waitForJobToFinish(c *client, jobID int) {
	c.do(func(client *api.Client) (interface{}, error) {
		for !startNoWait {
			status, err := client.GetJobStatus(jobID)
			if err != nil {
				return nil, fmt.Errorf("cannot get job status: %v", err)
			}
			if status.StatusID == db.StatusIDFailed {
				return nil, fmt.Errorf("job %d failed", status.JobID)
			}
			if status.StatusID == db.StatusIDDone {
				return nil, nil
			}
			log.Infof("job %d: %s", jobID, status.JobName)
			time.Sleep(time.Duration(startSleepS) * time.Second)
		}
		return nil, nil
	})
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
		return fmt.Errorf("invalid book ID: %q", args[0])
	}
	c := newClient(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(c, jobID) {
		c.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostProfile(bid, args[1:]...)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(c, jobID)
	return c.done()
}

var startELCommand = cobra.Command{
	Use:   "el ID",
	Short: "Create extended lexicon for book ID",
	Args:  exactlyNIDs(1),
	RunE:  doEL,
}

func doEL(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book ID: %q", args[0])
	}
	c := newClient(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(c, jobID) {
		c.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostExtendedLexicon(bid)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(c, jobID)
	if !startNoWait {
		c.do(func(client *api.Client) (interface{}, error) {
			return client.GetExtendedLexicon(bid)
		})
	}
	return c.done()
}

var startRRDMCommand = cobra.Command{
	Use:   "rrdm ID",
	Short: "Start automatic post-correction on book ID",
	Args:  exactlyNIDs(1),
	RunE:  doRRDM,
}

func doRRDM(_ *cobra.Command, args []string) error {
	var bid int
	if n := parseIDs(args[0], &bid); n != 1 {
		return fmt.Errorf("invalid book ID: %q", args[0])
	}
	c := newClient(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(c, jobID) {
		c.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostPostCorrection(bid)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(c, jobID)
	return c.done()
}
