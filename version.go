package main

import (
	"fmt"

	"github.com/finkf/pcwgo/api"
	"github.com/spf13/cobra"
)

var versionCommand = cobra.Command{
	Use:   "version",
	Short: "get version information",
	RunE:  runVersion,
}

func runVersion(_ *cobra.Command, args []string) error {
	url := getURL()
	if url == "" {
		return fmt.Errorf("missing url: use --url, or set POCOWEBC_URL")
	}
	var version api.Version
	c := api.NewClient(url, mainArgs.skipVerify)
	if err := get(c, c.URL("api-version"), &version); err != nil {
		return fmt.Errorf("get api version: %v", err)
	}
	return nil
}
