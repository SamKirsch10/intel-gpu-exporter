package linux

import "github.com/prometheus/client_golang/prometheus"

var (
	igpuEnginesMetrics map[string]map[string]prometheus.Gauge

	igpuFrequencyRequested = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_frequency_requested",
	}, []string{"unit"})
	igpuFrequencyActual = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_frequency_actual_mhz",
	}, []string{"unit"})
	igpuInterrupts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_interrupts",
	}, []string{"unit"})
	igpuBandwidthReads = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_imc_bandwidth_reads",
	}, []string{"unit"})
	igpuBandwidthWrites = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_imc_bandwidth_writes",
	}, []string{"unit"})
)

func init() {
	igpuEnginesMetrics = make(map[string]map[string]prometheus.Gauge)
	prometheus.MustRegister(igpuFrequencyRequested)
	prometheus.MustRegister(igpuFrequencyActual)
	prometheus.MustRegister(igpuInterrupts)
	prometheus.MustRegister(igpuBandwidthReads)
	prometheus.MustRegister(igpuBandwidthWrites)
}
