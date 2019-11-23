package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/finkf/gocropus"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var snippetNoGT = false

func init() {
	snippetGetCommand.Flags().BoolVarP(&snippetNoGT, "nogt", "n",
		false, "do not crate gt file for snippet")
}

var snippetCommand = cobra.Command{
	Use:   "snippet",
	Short: "work with line snippets",
}

var snippetGetCommand = cobra.Command{
	Use:   "get DIR IDS...",
	Short: "download line snippets IDS into directory DIR",
	Run:   doGetSnippets,
	Args:  cobra.MinimumNArgs(1),
}

func doGetSnippets(cmd *cobra.Command, args []string) {
	c := api.Authenticate(getURL(), getAuth(), skipVerify)
	var line api.Line
	for _, id := range args[1:] {
		var bid, pid, lid int
		if parseIDs(id, &bid, &pid, &lid) != 3 {
			log.Fatalf("invalid line id: %s", id)
		}
		line.ProjectID = bid
		line.PageID = pid
		line.LineID = lid
		downloadLineSnippet(c, &line, args[0])
	}
}

func downloadLineSnippet(c *api.Client, line *api.Line, dir string) {
	handle(c.GetLineX(line), "cannot download line %s: %v", line.ID())
	// image
	path := filepath.Join(dir, line.ID()+".png")
	iout, err := os.Create(path)
	handle(err, "cannot write image %s: %v", path)
	defer iout.Close()
	handle(c.DownloadLinePNG(iout, line),
		"cannot download line %s: %v", line.ID())
	// gt
	if !snippetNoGT {
		path = filepath.Join(dir, line.ID()+".gt.txt")
		handle(ioutil.WriteFile(path, []byte(line.Cor), 0666),
			"cannot write ground-truth %s: %v", path)
	}
	// ocr
	path = filepath.Join(dir, line.ID()+".txt")
	handle(ioutil.WriteFile(path, []byte(line.OCR), 0666),
		"cannot write ocr %s: %v", path)
}

var snippetPutCommand = cobra.Command{
	Use:   "put FILES...",
	Short: "update snipptes FILES to the servcer",
	Run:   doPutSnippets,
}

func doPutSnippets(_ *cobra.Command, args []string) {
	c := api.Authenticate(getURL(), getAuth(), skipVerify)
	for _, file := range args {
		var img, llocs string
		switch filepath.Ext(file) {
		case gocropus.PngExt:
			img = file
			llocs, _ = gocropus.LLocsFromStripped(file, false)
		case gocropus.LLocsExt:
			llocs = file
			img, _ = gocropus.ImageFromStripped(file)
		default:
			log.Fatalf("error: bad filename: %s", file)
		}
		putSnippet(c, img, llocs)
	}
}

func putSnippet(c *api.Client, img, llocs string) {

}
