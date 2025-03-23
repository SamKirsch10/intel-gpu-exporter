package linux

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type GPUData struct {
	Engines   map[string]map[string]any `json:"engines"`
	Frequency struct {
		Requested float64 `json:"requested"`
		Actual    float64 `json:"actual"`
		Unit      string  `json:"unit"`
	} `json:"frequency"`
	IMCBandwidth struct {
		Reads  float64 `json:"reads"`
		Writes float64 `json:"writes"`
		Unit   string  `json:"unit"`
	} `json:"imc-bandwidth"`
	Interrupts struct {
		Count float64 `json:"count"`
		Unit  string  `json:"unit"`
	} `json:"interrupts"`
}

type MetricGatherer struct {
	Device        string
	RefreshPeriod time.Duration
}

func NewGatherer(device string, refresh string) *MetricGatherer {
	d, err := time.ParseDuration(refresh)
	if err != nil {
		log.Fatalf("bad duration passed to gatherer: %v", err)
	}
	return &MetricGatherer{
		Device:        device,
		RefreshPeriod: d,
	}
}

func (g *MetricGatherer) Start(ctx context.Context) error {
	cmd := fmt.Sprintf("intel_gpu_top -J -s %d", g.RefreshPeriod.Milliseconds())
	if g.Device != "" {
		cmd += " -d " + g.Device
	}
	log.Debugf("Executing command: '%s'", cmd)

	cmdParts := strings.Split(cmd, " ")
	process := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
	stdout, err := process.StdoutPipe()
	if err != nil {
		log.Fatal("Error creating stdout pipe: ", err)
		return err
	}

	err = process.Start()
	if err != nil {
		log.Fatal("Error starting process: ", err)
		return err
	}
	go g.do(ctx, stdout)

	return nil
}

func (g *MetricGatherer) do(ctx context.Context, procStdout io.ReadCloser) {
	defer procStdout.Close()
	scanner := bufio.NewScanner(procStdout)

	t := time.NewTicker(g.RefreshPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			var data GPUData
			var buffer bytes.Buffer
			for scanner.Scan() {
				line := scanner.Text()
				buffer.WriteString(line + "\n")
				if json.Valid(buffer.Bytes()) {
					err := json.Unmarshal([]byte(buffer.Bytes()), &data)
					if err != nil {
						log.Errorf("JSON decode error: %v", err)
						continue
					}
					buffer.Reset()
					updateMetrics(data)
				}
			}
		}
	}
}

func updateMetrics(data GPUData) {
	for engine, metrics := range data.Engines {
		if _, ok := igpuEnginesMetrics[engine]; !ok {
			igpuEnginesMetrics[engine] = make(map[string]prometheus.Gauge)
		}
		for m, val := range metrics {
			if m == "unit" {
				continue
			}
			if _, ok := igpuEnginesMetrics[engine][m]; !ok {
				igpuEnginesMetrics[engine][m] = prometheus.NewGauge(prometheus.GaugeOpts{
					Name: fmt.Sprintf("engine_%s_%s_percent", engine, m),
				})
				prometheus.MustRegister(igpuEnginesMetrics[engine][m])
			}
			igpuEnginesMetrics[engine][m].Set(val.(float64))
		}
	}

	igpuFrequencyActual.WithLabelValues(data.Frequency.Unit).Set(data.Frequency.Actual)
	igpuFrequencyRequested.WithLabelValues(data.Frequency.Unit).Set(data.Frequency.Requested)
	igpuBandwidthReads.WithLabelValues(data.IMCBandwidth.Unit).Set(data.IMCBandwidth.Reads)
	igpuBandwidthWrites.WithLabelValues(data.IMCBandwidth.Unit).Set(data.IMCBandwidth.Writes)
	igpuInterrupts.WithLabelValues(data.Interrupts.Unit).Set(data.Interrupts.Count)
}
