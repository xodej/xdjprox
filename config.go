package main

import (
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
)

// Config is needed for configuration from json
type Config struct {
	TargetURL   string
	EntryURL    string
	TimeFormat  string
	LogRequest  bool
	LogResponse bool
	EnableWrite bool
	LogFile     string
	LogAll      bool
}

// SetupLogFile enables logging into file
func (c *Config) SetupLogFile() func() {
	// no log file given
	if c.LogFile == "" {
		return nil
	}

	// create file handle for log file
	fh, err := os.OpenFile(c.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Printf("failed to create log file: %v", err)
		return nil
	}

	log.SetOutput(io.MultiWriter(os.Stdout, fh))

	return func() {
		err := fh.Close()
		if err != nil && fh != nil {
			log.Fatalf("failed to close log file: %#v", err)
		}
	}
}
