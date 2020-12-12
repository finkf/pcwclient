package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/UNO-SOFT/ulog"
	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

func chk(err error) {
	if err == nil {
		return
	}
	log.Fatalf("error: %v", err)
}

func exactArgs(allowed ...int) func(_ *cobra.Command, args []string) error {
	return func(_ *cobra.Command, args []string) error {
		n := len(args)
		for _, allowed := range allowed {
			if n == allowed {
				return nil
			}
		}
		return fmt.Errorf("invalid number of args: %d (allowed: %v)", n, args)
	}
}

func parseIDs(id string, ids ...*int) int {
	split := strings.Split(id, ":")
	var i int
	for i = 0; i < len(ids) && i < len(split); i++ {
		id, err := strconv.Atoi(split[i])
		if err != nil {
			return 0
		}
		*ids[i] = id
	}
	return i
}

func unescape(args ...string) []string {
	res := make([]string, len(args))
	for i := range args {
		u, err := strconv.Unquote(`"` + args[i] + `"`)
		if err != nil {
			res[i] = args[i]
		} else {
			res[i] = u
		}
	}
	return res
}

func getURL() string {
	if mainArgs.pocowebURL != "" {
		return mainArgs.pocowebURL
	}
	return os.Getenv("POCOWEB_URL")
}

func getAuth() string {
	if mainArgs.authToken != "" {
		return mainArgs.authToken
	}
	return os.Getenv("POCOWEB_AUTH")
}

func get(c *api.Client, url string, out interface{}) error {
	ulog.Write("get", "method", "GET", "url", url)
	return c.Get(url, out)
}

func post(c *api.Client, url string, payload, out interface{}) error {
	ulog.Write("post", "method", "POST", "url", url)
	return c.Post(url, payload, out)
}

func delete(c *api.Client, url string, out interface{}) error {
	ulog.Write("post", "method", "DELETE", "url", url)
	return c.Delete(url, nil)
}

func downloadZIP(c *api.Client, url string, out io.Writer) error {
	ulog.Write("download zip", "method", "GET", "url", url)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("bad status code: %s", res.Status)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/zip" {
		return fmt.Errorf("bad content type: %s", ct)
	}
	_, err = io.Copy(out, res.Body)
	return err
}
