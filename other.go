package main

import (
	"net/http"

	"go.uber.org/zap"
)

func handleOther(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("received request with bad method or endpoint",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()))
	http.Error(w, "bad method or endpoint", http.StatusBadRequest)
}
