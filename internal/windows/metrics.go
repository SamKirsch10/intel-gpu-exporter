package windows

import "github.com/prometheus/client_golang/prometheus"

var (
	igpuUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_utilization_percent",
	}, []string{"gpu"})
	igpuMemoryUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "igpu_memory_usage_mb",
	}, []string{"gpu"})
)

func init() {
	prometheus.MustRegister(igpuUtilization)
	prometheus.MustRegister(igpuMemoryUsage)
}
