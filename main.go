package main

import (
	"context"
	"flag"
	"net/http"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samkirsch10/intel-gpu-exporter/internal/linux"
	"github.com/samkirsch10/intel-gpu-exporter/internal/windows"
	log "github.com/sirupsen/logrus"
)

func main() {

	device := flag.String("device", "", "Specify device for intel_gpu_top")
	refresh := flag.String("refresh", "5s", "Refresh period for metrics updates.")
	port := flag.String("port", "9091", "Port to serve metrics")
	loglvl := flag.String("log-level", "INFO", "Log level")
	args := flag.String("additional-args", "", "Additional args to pass to gatherer command.")
	flag.Parse()

	switch strings.ToUpper(*loglvl) {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	default:
		panic("unknown log level. Options are 'TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR'")
	}

	if runtime.GOOS == "windows" {
		windows.NewGatherer(*device, *refresh).Start(context.Background())
	} else {
		linux.NewGatherer(*device, *refresh, *args).Start(context.Background())
	}

	log.Infof("Starting GPU metrics exporter on port %s", *port)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
