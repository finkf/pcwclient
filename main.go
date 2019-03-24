package main // import "github.com/finkf/pocowebc"

import (
	"fmt"
	"os"
)

func main() {
	if err := mainCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
}
