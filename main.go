package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fatalbanana/clickhouse-query-exporter/envcfg"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	sigChan chan os.Signal
)

func runServer(bindAddr string, bindPort int) *http.Server {
	ws := &http.Server{
		Addr:              net.JoinHostPort(bindAddr, strconv.Itoa(bindPort)),
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       5 * time.Second,
	}

	go func() {
		err := ws.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("error starting webserver", zap.Error(err))
		}
	}()

	return ws
}

func main() {

	log := zap.L()

	http.HandleFunc("/probe", handleProbe)
	ws := runServer(envcfg.Cfg.BindAddr, envcfg.Cfg.BindPort)

	<-sigChan
	if ws == nil {
		return
	}
	shutCtx, cancelShutCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutCtx()

	err := ws.Shutdown(shutCtx)
	if err != nil {
		log.Error("error shutting down webserver", zap.Error(err))
	}
}

func init() {
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	config := zap.NewProductionConfig()
	if envcfg.Cfg.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	log, err := config.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(log)
}
