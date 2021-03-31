package config

import (
	"io"
	"os"
	"strings"

	"github.com/coreos/go-systemd/journal"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *Config) SetupLog() (err error) {
	level, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	logger, err := getLogger(c.LogDriver)
	if err != nil {
		return err
	}
	log.SetOutput(logger)
	return
}

func getLogger(driver string) (io.Writer, error) {
	switch {
	case strings.HasPrefix(driver, "journal://"):
		if !journal.Enabled() {
			return nil, errors.Errorf("failed to set logger: journal not enabled")
		}
	case strings.HasPrefix(driver, "file://"):
		return newFileLogger(strings.TrimPrefix(driver, "file://"))
	}
	return newFileLogger("/dev/null")
}

type journalLogger struct{}

func (l *journalLogger) Write(p []byte) (int, error) {
	return len(p), journal.Print(journal.PriInfo, string(p))
}

type fileLogger struct {
	*os.File
}

func newFileLogger(filePath string) (*fileLogger, error) {
	file, err := os.Open(filePath)
	return &fileLogger{file}, errors.WithStack(err)
}

func (l *fileLogger) Write(p []byte) (int, error) {
	return l.File.Write(p)
}
