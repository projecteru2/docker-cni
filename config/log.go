package config

import (
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *Config) SetupLog() (err error) {
	level, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		return errors.WithStack(err)
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
	case strings.HasPrefix(driver, "file://"):
		return newFileLogger(strings.TrimPrefix(driver, "file://"))
	}
	return newFileLogger("/dev/null")
}

type fileLogger struct {
	*os.File
}

func newFileLogger(filePath string) (*fileLogger, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	return &fileLogger{file}, errors.WithStack(err)
}

func (l *fileLogger) Write(p []byte) (int, error) {
	written, err := l.File.Write(p)
	return written, errors.WithStack(err)
}
