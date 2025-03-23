package windows

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type GPUData struct {
	Name        string
	Utilization float64
	MemoryUsage float64
}

type MetricGatherer struct {
	Device        string
	RefreshPeriod time.Duration
}

//go:embed files/windows_gpu.ps1
var powershellFile string

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

func CreateTempFile(prefix, content string) (string, error) {
	// Get the default temp directory
	tempDir := os.TempDir()

	// Create a temporary file with the given prefix
	tempFile, err := os.CreateTemp(tempDir, prefix+"*.ps1")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write content to the file
	_, err = tempFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil // Return the full path of the file
}

func (g *MetricGatherer) Start(ctx context.Context) error {
	tmpl := template.Must(template.New("gpu").Parse(powershellFile))
	var buff bytes.Buffer
	tmpl.Execute(&buff, struct {
		Device   string
		Interval int
	}{
		Device:   g.Device,
		Interval: int(g.RefreshPeriod.Seconds()),
	})

	path, err := CreateTempFile("gpu", buff.String())
	if err != nil {
		log.Errorf("failed to create tmp powershell file: %v", err)
		return err
	}

	cmd := fmt.Sprintf("powershell -ExecutionPolicy Bypass -File %s", path)
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
	igpuMemoryUsage.WithLabelValues(data.Name).Set(data.MemoryUsage)
	igpuUtilization.WithLabelValues(data.Name).Set(data.Utilization)
}
