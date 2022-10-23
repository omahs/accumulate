// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/tendermint/tendermint/config"
)

// NewConsoleWriter parses the log format and creates an appropriate writer.
// It is based on part of Tendermint's NewTendermintLogger.
func NewConsoleWriter(format string) (io.Writer, error) {
	return NewConsoleWriterWith(os.Stderr, format)
}

func NewConsoleWriterWith(w io.Writer, format string) (io.Writer, error) {
	switch strings.ToLower(format) {
	case config.LogFormatPlain:
		return newConsoleWriter(w), nil

	case config.LogFormatJSON:
		return w, nil

	default:
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}
}

// newConsoleWriter creates a zerolog console writer that formats log messages
// as plain text for the console. It is based on part of Tendermint's NewTendermintLogger.
func newConsoleWriter(w io.Writer) *zerolog.ConsoleWriter {
	return &zerolog.ConsoleWriter{
		Out: w,
		// NoColor:    true,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			if ll, ok := i.(string); ok {
				return strings.ToUpper(ll)
			}
			return "????"
		},
	}
}

type ServiceLogger struct {
	Service service.Logger

	fmt *zerolog.ConsoleWriter
	buf *bytes.Buffer
	mu  *sync.Mutex
}

var _ zerolog.LevelWriter = (*ServiceLogger)(nil)

func NewServiceLogger(svc service.Service, format string) (*ServiceLogger, error) {
	logger := new(ServiceLogger)
	var err error
	logger.Service, err = svc.Logger(nil)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(format) {
	case config.LogFormatPlain:
		logger.buf = new(bytes.Buffer)
		logger.mu = new(sync.Mutex)
		logger.fmt = newConsoleWriter(logger.buf)

	case config.LogFormatJSON:

	default:
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}

	return logger, nil
}

func (l *ServiceLogger) Write(b []byte) (int, error) {
	return l.WriteLevel(zerolog.NoLevel, b)
}

func (l *ServiceLogger) WriteLevel(level zerolog.Level, b []byte) (int, error) {
	// Use zerolog's console writer to format the log message
	if l.fmt != nil {
		l.mu.Lock()
		l.buf.Reset()
		_, _ = l.fmt.Write(b)
		b = make([]byte, l.buf.Len())
		copy(b, l.buf.Bytes())
		l.mu.Unlock()
	}

	switch level {
	case zerolog.PanicLevel, zerolog.FatalLevel, zerolog.ErrorLevel:
		_ = l.Service.Error(string(b))
	case zerolog.WarnLevel:
		_ = l.Service.Warning(string(b))
	default:
		_ = l.Service.Info(string(b))
	}
	return len(b), nil
}
