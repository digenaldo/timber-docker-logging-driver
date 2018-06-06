package main

import (
	"encoding/binary"
	"io"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	protoio "github.com/gogo/protobuf/io"

	"github.com/timberio/timber-go/batch"
	"github.com/timberio/timber-go/forward"
)

type timberLogger struct {
	batcher   *batch.Batcher
	forwarder *forward.HTTPForwarder
	info      logger.Info
	stream    io.ReadCloser
}

func (logger *timberLogger) startLogging() {
	go logger.consumeLogs()
	go forward.Forward(logger.batcher.BufferChan, logger.forwarder)
}

func (logger *timberLogger) stopLogging() {
	logger.stream.Close()
	// Closing the batcher input channel gracefully shuts down batcher and forwarder
	close(logger.batcher.ByteChan)
}

func (logger *timberLogger) consumeLogs() {
	dec := protoio.NewUint32DelimitedReader(logger.stream, binary.BigEndian, 1e6)
	defer dec.Close()

	var log logdriver.LogEntry

	for {
		if err := dec.ReadMsg(&log); err != nil {
			// exit consumer loop if reader reaches EOF or the fifo is closed by the writer
			// https://github.com/splunk/docker-logging-plugin/blob/release/2.0.0/message_processor.go#L62-L63
			if err == io.EOF || err == os.ErrClosed || strings.Contains(err.Error(), "file already closed") {
				logrus.WithField("id", logger.info.ContainerID).WithError(err).Debug("shutting down log consumer")
				return
			}

			logrus.WithField("id", logger.info.ContainerID).WithError(err).Error("received unexpected error. retrying...")
			dec = protoio.NewUint32DelimitedReader(logger.stream, binary.BigEndian, 1e6)
		}

		if len(log.Line) > 0 {
			logger.batcher.ByteChan <- log.Line
		}

		log.Reset()
	}
}
