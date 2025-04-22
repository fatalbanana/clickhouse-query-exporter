package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fatalbanana/clickhouse-query-exporter/envcfg"

	"github.com/uptrace/go-clickhouse/ch"
	"go.uber.org/zap"
)

func mustWrite(w io.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		zap.L().Debug("error writing response", zap.Error(err))
		panic(http.ErrAbortHandler)
	}
}

func handleProbe(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		zap.L().Debug("received request without target",
			zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}

	dsn, ok := envcfg.Cfg.DSNMap[target]
	if !ok {
		zap.L().Debug("received request with unrecognised target",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("target", target))
		http.Error(w, "target is unrecognised", http.StatusBadRequest)
		return
	}

	queryName := r.URL.Query().Get("query")
	if queryName == "" {
		zap.L().Debug("received request without query",
			zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	query, ok := envcfg.Cfg.QueryMap[queryName]
	if !ok {
		zap.L().Debug("received request with unrecognised query",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("query", query))
		http.Error(w, "query is unrecognised", http.StatusBadRequest)
		return
	}

	db := ch.Connect(ch.WithDSN(dsn))

	var res uint64
	var queryOK bool
	err := db.QueryRowContext(r.Context(), query).Scan(&res)
	if err == nil {
		queryOK = true
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	mustWrite(w, []byte(fmt.Sprintf("# HELP chquery_%s_ok Status of ClickHouse query: %s\n", queryName, queryName)))
	mustWrite(w, []byte(fmt.Sprintf("# TYPE chquery_%s_ok gauge\n", queryName)))
	if !queryOK {
		zap.L().Error("error executing query",
			zap.String("queryName", queryName),
			zap.String("target", target),
			zap.Error(err))
		mustWrite(w, []byte(fmt.Sprintf("chquery_%s_ok 0\n", queryName)))
		return
	}
	mustWrite(w, []byte(fmt.Sprintf("chquery_%s_ok 1\n", queryName)))

	mustWrite(w, []byte(fmt.Sprintf("# HELP chquery_%s_result Result of ClickHouse query: %s\n", queryName, queryName)))
	mustWrite(w, []byte(fmt.Sprintf("# TYPE chquery_%s_result gauge\n", queryName)))
	mustWrite(w, []byte(fmt.Sprintf("chquery_%s_result %d\n", queryName, res)))

	zap.L().Debug("served probe results",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("target", target),
		zap.String("query", query))
}
