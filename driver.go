package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/jsonfilelog"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/timberio/timber-go/batch"
	"github.com/timberio/timber-go/forward"
	"github.com/tonistiigi/fifo"
)

type timberDriver struct {
	idx  map[string]*timberLoggerCollection
	logs map[string]*timberLoggerCollection
	mu   sync.Mutex
}

type timberLoggerCollection struct {
	jsonLogger   logger.Logger
	timberLogger *timberLogger
}

func newTimberDriver() *timberDriver {
	return &timberDriver{
		logs: make(map[string]*timberLoggerCollection),
		idx:  make(map[string]*timberLoggerCollection),
	}
}

func (tlc *timberLoggerCollection) Close() {
	tlc.timberLogger.Close()
	tlc.jsonLogger.Close()
}

func (driver *timberDriver) startLogging(file string, logCtx logger.Info) error {
	driver.mu.Lock()
	if _, exists := driver.logs[file]; exists {
		driver.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	driver.mu.Unlock()

	// Create a logger handler
	if logCtx.LogPath == "" {
		logCtx.LogPath = filepath.Join("/var/log/docker", logCtx.ContainerID)
	}

	if err := os.MkdirAll(filepath.Dir(logCtx.LogPath), 0755); err != nil {
		return errors.Wrap(err, "error setting up logger dir")
	}

	// Create json logger
	jsonLogger, err := jsonfilelog.New(logCtx)
	if err != nil {
		return errors.Wrap(err, "error creating jsonfile logger")
	}

	// Validate apikey
	apiKey := logCtx.Config["timber-api-key"]
	if apiKey == "" {
		return fmt.Errorf("api key not found. log-opt timber-api-key is required")
	}

	// Create input channel for messages from fifo and pass to batcher
	inputChan := make(chan []byte)
	batcher := batch.NewBatcher(inputChan, batch.Config{
		Logger: logrus.StandardLogger(),
	})

	// Create httpForwarder to logs to Timber backend
	forwarder, err := forward.NewHTTPForwarder(apiKey,
		forward.Config{
			Logger: logrus.StandardLogger(),
		},
	)

	// Open fifo for reading container logs
	logrus.WithField("id", logCtx.ContainerID).WithField("file", file).WithField("logpath", logCtx.LogPath).Debugf("Start logging")
	logFifo, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	// Save current state
	driver.mu.Lock()
	tl := &timberLogger{
		batcher:   batcher,
		forwarder: forwarder,
		info:      logCtx,
		stream:    logFifo,
	}
	tlc := &timberLoggerCollection{jsonLogger: jsonLogger, timberLogger: tl}
	driver.logs[file] = tlc
	driver.idx[logCtx.ContainerID] = tlc
	driver.mu.Unlock()

	// Start consuming container logs
	tl.Start(jsonLogger)

	return nil
}

func (driver *timberDriver) StopLogging(file string) error {
	logrus.WithField("file", file).Debugf("Stop logging")

	driver.mu.Lock()
	tlc, ok := driver.logs[file]
	if ok {
		//	Stop logger and remove reference from driver state
		tlc.Close()
		delete(driver.logs, file)
	}
	driver.mu.Unlock()

	return nil
}

func (driver *timberDriver) ReadLogs(info logger.Info, config logger.ReadConfig) (io.ReadCloser, error) {
	driver.mu.Lock()
	tlc, exists := driver.idx[info.ContainerID]
	driver.mu.Unlock()
	if !exists {
		return nil, fmt.Errorf("logger does not exist for %s", info.ContainerID)
	}

	reader, writer := io.Pipe()
	logReader, ok := tlc.jsonLogger.(logger.LogReader)
	if !ok {
		return nil, fmt.Errorf("logger does not support reading")
	}

	go func() {
		watcher := logReader.ReadLogs(config)

		enc := protoio.NewUint32DelimitedWriter(writer, binary.BigEndian)
		defer enc.Close()
		defer watcher.Close()

		var buf logdriver.LogEntry
		for {
			select {
			case msg, ok := <-watcher.Msg:
				if !ok {
					writer.Close()
					return
				}

				buf.Line = msg.Line
				buf.Partial = msg.Partial
				buf.TimeNano = msg.Timestamp.UnixNano()
				buf.Source = msg.Source

				if err := enc.WriteMsg(&buf); err != nil {
					writer.CloseWithError(err)
					return
				}
			case err := <-watcher.Err:
				writer.CloseWithError(err)
				return
			}

			buf.Reset()
		}
	}()

	return reader, nil
}
