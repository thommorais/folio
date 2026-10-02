package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
)

const (
	exitFailure    = 1
	exitAuth       = 2
	exitNotFound   = 3
	exitValidation = 4
)

func exitCode(err error) int {
	if errors.Is(err, config.ErrNoToken) || errors.Is(err, client.ErrNoToken) {
		return exitAuth
	}
	if errors.Is(err, errUncommitted) {
		return exitValidation
	}

	switch status, _ := client.Status(err); status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return exitAuth
	case http.StatusNotFound:
		return exitNotFound
	case http.StatusBadRequest:
		return exitValidation
	}
	return exitFailure
}

func errorJSON(err error) string {
	report := struct {
		Error  string            `json:"error"`
		Status int               `json:"status,omitempty"`
		Fields map[string]string `json:"fields,omitempty"`
	}{Error: err.Error(), Fields: client.FieldErrors(err)}
	report.Status, _ = client.Status(err)

	out, _ := json.Marshal(report)
	return string(out)
}
