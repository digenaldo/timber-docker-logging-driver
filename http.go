package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/go-plugins-helpers/sdk"
)

// NOTE: Fields for request and response structs must be public or else
// Docker will hang when using this log driver.

type startLoggingRequest struct {
	File string
	Info logger.Info
}

type stopLoggingRequest struct {
	File string
}

func handlers(h *sdk.Handler, driver *timberDriver) {
	h.HandleFunc("/LogDriver.StartLogging", func(w http.ResponseWriter, r *http.Request) {
		var req startLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Info.ContainerID == "" {
			respond(errors.New("must provide container id in log context"), w)
			return
		}

		err := driver.startLogging(req.File, req.Info)
		respond(err, w)
	})

	h.HandleFunc("/LogDriver.StopLogging", func(w http.ResponseWriter, r *http.Request) {
		var req stopLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := driver.stopLogging(req.File)
		respond(err, w)
	})
}

type response struct {
	Err string
}

func respond(err error, w http.ResponseWriter) {
	var res response
	if err != nil {
		res.Err = err.Error()
	}
	json.NewEncoder(w).Encode(&res)
}
