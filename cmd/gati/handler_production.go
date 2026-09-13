//go:build !simulation

package main

import (
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/httpapi"
	"github.com/jegati/jegati/internal/store"
	"net/http"
)

func configureHandler(c config.Config, backend *store.Store, roads []byte) (http.Handler, error) {
	return httpapi.Handler(c, backend, roads), nil
}
