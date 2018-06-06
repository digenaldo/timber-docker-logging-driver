package main

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/logger"
	"github.com/pkg/errors"
	"github.com/timberio/timber-go/batch"
	"github.com/timberio/timber-go/forward"
	"github.com/tonistiigi/fifo"
)

type timberDriver struct {
	loggers map[string]*timberLogger
	mu      sync.Mutex
}

func newTimberDriver() *timberDriver {
	return &timberDriver{
		loggers: make(map[string]*timberLogger),
	}
}

func (driver *timberDriver) startLogging(file string, logCtx logger.Info) error {
	driver.mu.Lock()
	if _, exists := driver.loggers[file]; exists {
		driver.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	driver.mu.Unlock()

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
	logger := &timberLogger{
		batcher:   batcher,
		forwarder: forwarder,
		info:      logCtx,
		stream:    logFifo,
	}
	driver.loggers[file] = logger
	driver.mu.Unlock()

	// Start consuming container logs
	logger.startLogging()

	return nil
}

func (driver *timberDriver) stopLogging(file string) error {
	logrus.WithField("file", file).Debugf("Stop logging")

	driver.mu.Lock()
	logger, ok := driver.loggers[file]
	if ok {
		//	Stop logger and remove reference from driver state
		logger.stopLogging()
		delete(driver.loggers, file)
	}
	driver.mu.Unlock()

	return nil
}
