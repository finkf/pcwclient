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

func reattach(cmd *command, jobID int) bool {
	var res bool
	if !startNoWait {
		cmd.do(func(client *api.Client) (interface{}, error) {
			status, err := cmd.client.GetJobStatus(jobID)
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

func waitForJobToFinish(cmd *command, jobID int) {
	cmd.do(func(client *api.Client) (interface{}, error) {
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
	cmd := newCommand(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(&cmd, jobID) {
		cmd.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostProfile(bid, args[1:]...)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(&cmd, jobID)
	return cmd.done()
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
	cmd := newCommand(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(&cmd, jobID) {
		cmd.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostExtendedLexicon(bid)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(&cmd, jobID)
	if !startNoWait {
		cmd.do(func(client *api.Client) (interface{}, error) {
			return client.GetExtendedLexicon(bid)
		})
	}
	return cmd.done()
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
	cmd := newCommand(os.Stdout)
	// start profiling
	jobID := bid
	if !reattach(&cmd, jobID) {
		cmd.do(func(client *api.Client) (interface{}, error) {
			job, err := client.PostPostCorrection(bid)
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(&cmd, jobID)
	return cmd.done()
}

var startPredictCommand = cobra.Command{
	Use:   "predict ID NAME",
	Short: "Run OCR model NAME on book ID",
	Args:  cobra.ExactArgs(2),
	RunE:  doPredict,
}

func doPredict(_ *cobra.Command, args []string) error {
	var bid, pid, lid int
	if parseIDs(args[0], &bid, &pid, &lid) == 0 {
		return fmt.Errorf("invalid book id: %s", args[0])
	}
	cmd := newCommand(os.Stdout)
	jobID := bid
	if !reattach(&cmd, jobID) {
		cmd.do(func(client *api.Client) (interface{}, error) {
			job, err := client.OCRPredict(bid, pid, lid, args[1])
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(&cmd, jobID)
	return cmd.done()
}

var startTrainCommand = cobra.Command{
	Use:   "train ID NAME",
	Short: "Train an OCR model on book ID using NAME as base model (not implemented yet)",
	Args:  cobra.ExactArgs(2),
	RunE:  doTrain,
}

func doTrain(_ *cobra.Command, args []string) error {
	var bid int
	if parseIDs(args[0], &bid) != 1 {
		return fmt.Errorf("invalid book id: %s", args[0])
	}
	cmd := newCommand(os.Stdout)
	jobID := bid
	if !reattach(&cmd, jobID) {
		cmd.do(func(client *api.Client) (interface{}, error) {
			job, err := cmd.client.OCRTrain(bid, args[1])
			jobID = job.ID
			return nil, err
		})
	}
	waitForJobToFinish(&cmd, jobID)
	return cmd.done()
}
