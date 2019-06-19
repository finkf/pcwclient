package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showCommand = cobra.Command{
	Use:   "show ID [IDS...]",
	Short: "Show image files for the given line or page IDS",
	RunE:  show,
	Args:  cobra.MinimumNArgs(1),
}

func show(_ *cobra.Command, args []string) error {
	cmd := newCommand(os.Stdout)
	for _, id := range args {
		var bid, pid, lid int
		if n := parseIDs(id, &bid, &pid, &lid); n < 2 {
			return fmt.Errorf("invalid id: %s", id)
		}
		cmd.do(func() error {
			if lid == 0 {
				p, err := cmd.client.GetPage(bid, pid)
				if err != nil {
					return err
				}
				return showImage(os.Stdout, p.ImgFile)
			}
			l, err := cmd.client.GetLine(bid, pid, lid)
			if err != nil {
				return err
			}
			return showImage(os.Stdout, l.ImgFile)
		})
	}
	return cmd.err
}

func showImage(out io.Writer, imgpath string) error {
	u, err := url.Parse(getURL())
	if err != nil {
		return err
	}
	u.Path = imgpath
	log.Debugf("GET %s", u.String())
	req, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer req.Body.Close()
	log.Debugf("response from server: %s", req.Status)
	if req.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status: %s", req.Status)
	}
	_, err = io.Copy(out, req.Body)
	return err
}
