package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"

	"git.sr.ht/~flobar/gocropus"
	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	snippetGetNoGT  = false
	snippetPutImage = false
	snippetPutNoOCR = false
)

func init() {
	snippetGetCommand.Flags().BoolVarP(&snippetGetNoGT, "nogt", "n",
		false, "do not crate gt file for snippets")
	snippetPutCommand.Flags().BoolVarP(&snippetPutImage, "image", "i",
		false, "upload image files")
	snippetPutCommand.Flags().BoolVarP(&snippetPutNoOCR, "noocr", "n",
		false, "do not upload ocr data")
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
	if !snippetGetNoGT {
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
	Short: "update snipptes FILES to the server",
	Run:   doPutSnippets,
}

func doPutSnippets(_ *cobra.Command, args []string) {
	c := api.Authenticate(getURL(), getAuth(), skipVerify)
	for _, file := range args {
		var img, llocs string
		switch filepath.Ext(file) {
		case gocropus.PngExt:
			img = file
			llocs, _ = gocropus.LLocsFromFile(file, false)
		case gocropus.LLocsExt:
			llocs = file
			img, _ = gocropus.ImageFromFile(file)
		default:
			log.Fatalf("error: bad filename: %s", file)
		}
		putSnippet(c, img, llocs)
	}
}

func putSnippet(c *api.Client, img, llocs string) {
	var data api.PostOCR
	var line api.Line
	addID(&line, gocropus.Strip(llocs))
	if snippetPutImage {
		addImageData(&data, img)
	}
	if !snippetPutNoOCR {
		addOCRData(&data, llocs)
	}
	handle(c.PutLineOCR(&line, &data), "cannot put snippet %s: %v", llocs)
}

func addImageData(data *api.PostOCR, path string) {
	in, err := os.Open(path)
	handle(err, "cannot open image %s: %v", path)
	defer in.Close()
	img, err := png.Decode(in)
	handle(err, "cannot decode image %s: %v", path)
	var buf bytes.Buffer
	handle(png.Encode(&buf, img), "cannot convert image to bytes: %v")
	data.ImageData = base64.StdEncoding.EncodeToString(buf.Bytes())
}

func addOCRData(data *api.PostOCR, path string) {
	llocs, err := gocropus.OpenLLocsFile(path)
	handle(err, "cannot open llocs %s: %v", path)
	data.OCR = llocs.String()
	data.Cuts = llocs.Cuts()
	data.Confs = llocs.Confs()
}

func addID(line *api.Line, stripped string) {
	// Try filename: id:id:id
	var bid, pid, lid int
	if n := parseIDs(filepath.Base(stripped), &bid, &pid, &lid); n == 3 {
		line.ProjectID = bid
		line.PageID = pid
		line.LineID = lid
		return
	}
	// Try file path: .../id/id/id
	elems := filepath.SplitList(stripped)
	if len(elems) < 3 {
		log.Fatalf("error: cannot determine line ID for %s", stripped)
	}
	elems = elems[len(elems)-3:]
	if _, err := fmt.Sscanf(elems[0], "%d", &line.ProjectID); err != nil {
		log.Fatalf("error: cannot determine line ID for %s", stripped)
	}
	if _, err := fmt.Sscanf(elems[1], "%d", &line.PageID); err != nil {
		log.Fatalf("error: cannot determine line ID for %s", stripped)
	}
	if _, err := fmt.Sscanf(elems[2], "%d", &line.LineID); err != nil {
		log.Fatalf("error: cannot determine line ID for %s", stripped)
	}
}
