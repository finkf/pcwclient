package main // import "github.com/finkf/pcwclient"

import log "github.com/sirupsen/logrus"

func main() {
	if err := mainCommand.Execute(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
