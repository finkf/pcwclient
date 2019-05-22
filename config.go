package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type config struct {
	URL  string
	Auth string
}

func (c *config) save(p string) (err error) {
	if err := os.MkdirAll(path.Dir(p), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err == nil {
			err = e
		}
	}()
	if _, err := fmt.Fprintf(out, "# Automtically created by pocoweb at %s\n",
		time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	return toml.NewEncoder(out).Encode(c)
}

func (c *config) load(p string) error {
	_, err := toml.DecodeFile(p, c)
	return err
}

// load config from configpath if noconfig is false
// or load the config from auth or url or load the config
// using POCOWEBC_URL and POCOWEBC_AUTH environment variables.
func loadConfig() *config {
	if noconfig {
		return &config{URL: getURL(), Auth: getAuth()}
	}
	c := &config{}
	var path string
	if configpath != "" {
		path = configpath
	}
	if path == "" {
		path = os.Getenv("POCOWEBC_CONFIG")
	}
	if err := c.load(path); err != nil {
		log.Errorf("cannot open config: %v", err)
		return &config{URL: getURL(), Auth: getAuth()}
	}
	return c
}

func saveConfig(c *config) {
	if noconfig { // do not save config
		return
	}
	var path string
	if configpath != "" {
		path = configpath
	}
	if path == "" {
		path = os.Getenv("POCOWEBC_CONFIG")
	}
	if err := c.save(path); err != nil {
		log.Errorf("cannot save config: %v", err)
	}
}

func getURL() string {
	if pocowebURL != "" {
		return pocowebURL
	}
	return os.Getenv("POCOWEBC_URL")
}

func getAuth() string {
	if authToken != "" {
		return authToken
	}
	return os.Getenv("POCOWEBC_AUTH")
}
