package main

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"time"

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

func (tl *timberLogger) Start(fileLogger logger.Logger) {
	go tl.consumeLogs(fileLogger)
	go forward.Forward(tl.batcher.BufferChan, tl.forwarder)
}

func (tl *timberLogger) Close() {
	tl.stream.Close()
	// Closing the batcher input channel gracefully shuts down batcher and forwarder
	close(tl.batcher.ByteChan)
}

func (tl *timberLogger) consumeLogs(fileLogger logger.Logger) {
	dec := protoio.NewUint32DelimitedReader(tl.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer tl.Close()

	var log logdriver.LogEntry

	for {
		if err := dec.ReadMsg(&log); err != nil {
			// exit consumer loop if reader reaches EOF or the fifo is closed by the writer
			// https://github.com/splunk/docker-logging-plugin/blob/release/2.0.0/message_processor.go#L62-L63
			if err == io.EOF || err == os.ErrClosed || strings.Contains(err.Error(), "file already closed") {
				logrus.WithField("id", tl.info.ContainerID).WithError(err).Debug("shutting down log consumer")
				return
			}

			logrus.WithField("id", tl.info.ContainerID).WithError(err).Error("received unexpected error. retrying...")
			dec = protoio.NewUint32DelimitedReader(tl.stream, binary.BigEndian, 1e6)
		}

		if len(log.Line) > 0 {
			// Build log message to send to jsonLogger
			var msg logger.Message

			msg.Line = log.Line
			msg.Source = log.Source
			msg.Partial = log.Partial // NOTE: Ignored for now
			msg.Timestamp = time.Unix(0, log.TimeNano)

			if err := fileLogger.Log(&msg); err != nil {
				logrus.WithField("source", log.Source).WithError(err).WithField("message",
					msg).Error("Error writing log message")
			}

			// Send to timber after we have attempted to log to local disk
			tl.batcher.ByteChan <- log.Line
		}

		log.Reset()
	}
}
