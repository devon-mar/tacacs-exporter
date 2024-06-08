package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/devon-mar/tacacs-exporter/collector"
	"github.com/devon-mar/tacacs-exporter/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	showVer     = flag.Bool("version", false, "Show the version")
	configPath  = flag.String("config", "config.yml", "Path to the config file.")
	listenAddr  = flag.String("web.listen-address", ":9949", "HTTP server listen address")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	logLevel    = flag.String("log.level", "info", "The log level.")

	exporterConfig *config.Config

	exporterVersion = "development"
	exporterSha     = "123"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "no target specified", http.StatusBadRequest)
		return
	}

	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "no module specified", http.StatusBadRequest)
		return
	}

	module, ok := exporterConfig.Modules[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown module %q", moduleName), http.StatusBadRequest)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(target, r.RemoteAddr, &module))

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}

func configureLogging() {
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(*logLevel)); err != nil {
		slog.Error("invalid log level", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(logger)
	// functions from the log package will use this level.
	slog.SetLogLoggerLevel(slog.LevelInfo)
}

func main() {
	flag.Parse()

	if *showVer {
		fmt.Printf("Version: %s\n", exporterVersion)
		fmt.Printf("SHA: %s\n", exporterSha)
		os.Exit(0)
	}
	slog.Info("starting TACACS exporter")

	configureLogging()

	var err error
	exporterConfig, err = config.LoadFromFile(*configPath)
	if err != nil {
		slog.Error("error loading config", "err", err)
		os.Exit(1)
	}

	http.Handle(*metricsPath, http.HandlerFunc(metricsHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>TACACS Exporter</title></head>
		<body>
			<h1>TACACS Exporter</h1>
			<a href="` + *metricsPath + `">Metrics</a>
		</body>
		</html>`))
	})

	server := http.Server{Addr: *listenAddr}
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)

		signal.Notify(sigCh, os.Interrupt)
		sig := <-sigCh
		slog.Warn("received signal", "signal", sig)

		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("HTTP server Shutdown", "err", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("HTTP server ListenAndServe", "err", err)
		os.Exit(1)
	}

	<-idleConnsClosed
}
