package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/finkf/pcwgo/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	userPool       bool
	persistentPool bool
	poolCommands   []string
	poolBaseDir    string
)

func init() {
	poolCommand.PersistentFlags().BoolVarP(&userPool,
		"user", "u", false, "download user pool")
	runPoolCommand.Flags().BoolVarP(&persistentPool,
		"persistent", "p", false, "do not remove temporary files")
	runPoolCommand.Flags().StringArrayVarP(&poolCommands,
		"command", "c", nil, "run command")
	runPoolCommand.Flags().StringVarP(&poolBaseDir,
		"base", "b", "", "set base dir for data (uses a random temp dir if empty)")
}

var poolCommand = cobra.Command{
	Use:   "pool",
	Short: "Handle user/global pool",
}

var downloadPoolCommand = cobra.Command{
	Use:   "download",
	Short: "Download zipped user/global pool and write it to stdout",
	Args:  cobra.ExactArgs(0),
	RunE:  downloadPool,
}

func downloadPool(_ *cobra.Command, _ []string) error {
	c := newClient(os.Stdout)
	c.do(func(client *api.Client) (interface{}, error) {
		out := os.Stdout
		return nil, downloadGlobalOrUserPool(out, client, userPool)
	})
	return c.done()
}

func downloadGlobalOrUserPool(out io.Writer, c *api.Client, user bool) error {
	if user {
		return c.DownloadUserPool(out)
	}
	return c.DownloadGlobalPool(out)
}

var runPoolCommand = cobra.Command{
	Use:   "run",
	Short: "Run commands on dowloaded pool",
	Args:  cobra.ExactArgs(0),
	RunE:  runPool,
}

func runPool(_ *cobra.Command, _ []string) error {
	r := poolRunner{
		client:     newClient(os.Stdout),
		baseDir:    poolBaseDir,
		persistent: persistentPool,
		user:       userPool,
	}
	r.run(poolCommands...)
	return r.done()
}

type poolRunner struct {
	client           *client
	zipr             *zip.Reader
	baseDir          string
	persistent, user bool
}

func (r *poolRunner) run(cmds ...string) {
	r.download()
	r.extract()
	for _, cmd := range cmds {
		r.runCmd(cmd)
	}
}

func (r *poolRunner) runCmd(cmdstr string) {
	log.Debugf("runCmd(%q)", cmdstr)
	r.client.do(func(client *api.Client) (interface{}, error) {
		cmd := exec.Command("sh", "-c", cmdstr)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PCWCLIENT_AUTH=%s", getAuth()),
			fmt.Sprintf("PCWCLIENT_URL=%s", getURL()),
		)
		cmd.Dir = filepath.Join(r.baseDir, "corpus")
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		go func() {
			defer stderr.Close()
			_, _ = io.Copy(os.Stderr, stderr)
		}()
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		go func() {
			defer stdout.Close()
			_, _ = io.Copy(os.Stdout, stdout)
		}()
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if err := cmd.Wait(); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (r *poolRunner) extract() {
	log.Debugf("extract")
	r.client.do(func(client *api.Client) (interface{}, error) {
		if r.baseDir == "" {
			dir, err := ioutil.TempDir("", "pcwclient_pool_")
			if err != nil {
				return nil, err
			}
			r.baseDir = dir
		}
		for _, file := range r.zipr.File {
			if _, err := r.extractFile(file); err != nil {
				return nil, err
			}
		}
		log.Debugf("extracted %d files to %s", len(r.zipr.File), filepath.Join(r.baseDir))
		// no need to close zipr (its just a wrapper around bytes.Buffer)
		r.zipr = nil
		return nil, nil
	})
}

func (r *poolRunner) extractFile(file *zip.File) (opath string, oerr error) {
	opath = filepath.Join(r.baseDir, file.Name)
	if err := os.MkdirAll(filepath.Dir(opath), os.ModePerm); err != nil {
		return "", err
	}
	in, err := file.Open()
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(opath)
	if err != nil {
		return "", err
	}
	defer func() {
		x := out.Close()
		if oerr != nil {
			oerr = x
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return opath, nil
}

func (r *poolRunner) download() {
	r.client.do(func(client *api.Client) (interface{}, error) {
		var buf bytes.Buffer
		if err := downloadGlobalOrUserPool(&buf, client, r.user); err != nil {
			return nil, err
		}
		n := int64(buf.Len())
		log.Debugf("downloaded %.2fMB", float64(n)/float64(1024*1024))
		zipr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), n)
		if err != nil {
			return nil, err
		}
		r.zipr = zipr
		return nil, nil
	})
}

func (r *poolRunner) done() error {
	log.Debugf("done")
	var err error
	if !r.persistent {
		err = os.RemoveAll(r.baseDir)
	}
	if r.client.err != nil {
		return r.client.err
	}
	return err
}
