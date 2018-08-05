package sslcert

import (
	"strconv"

	"github.com/czerwonk/atlas_exporter/exporter"

	"github.com/DNS-OARC/ripeatlas/measurement"
	"github.com/czerwonk/atlas_exporter/probe"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ns  = "atlas"
	sub = "sslcert"
)

var (
	labels               []string
	rttDesc              *prometheus.Desc
	sslVerDesc           *prometheus.Desc
	successDesc          *prometheus.Desc
	alertLevelDesc       *prometheus.Desc
	alertDescriptionDesc *prometheus.Desc
)

func init() {
	labels = []string{"measurement", "probe", "dst_addr", "asn", "ip_version", "country_code", "lat", "long"}

	successDesc = prometheus.NewDesc(prometheus.BuildFQName(ns, sub, "success"), "Destination was reachable", labels, nil)
	sslVerDesc = prometheus.NewDesc(prometheus.BuildFQName(ns, sub, "version"), "SSL/TLS version used for the request", labels, nil)
	rttDesc = prometheus.NewDesc(prometheus.BuildFQName(ns, sub, "rtt"), "Round trip time in ms", labels, nil)
	alertLevelDesc = prometheus.NewDesc(prometheus.BuildFQName(ns, sub, "alert_level"), "Status of the SSL/TLS certificate (0 = valid)", labels, nil)
	alertDescriptionDesc = prometheus.NewDesc(prometheus.BuildFQName(ns, sub, "alert_description"), "Description for the alert level (see RIPIE Atlas documentation)", labels, nil)
}

type sslCertMetricExporter struct {
	id string
}

// NewExporter creates a exporter for SSL measurement results
func NewExporter(id string) exporter.MetricExporter {
	return &sslCertMetricExporter{id}
}

// Export exports a prometheus metric
func (m *sslCertMetricExporter) Export(res *measurement.Result, probe *probe.Probe, ch chan<- prometheus.Metric) {
	labelValues := []string{
		m.id,
		strconv.Itoa(probe.ID),
		res.DstAddr(),
		strconv.Itoa(probe.ASNForIPVersion(res.Af())),
		strconv.Itoa(res.Af()),
		probe.CountryCode,
		probe.Latitude(),
		probe.Longitude(),
	}

	ver, _ := strconv.ParseFloat(res.Ver(), 64)
	ch <- prometheus.MustNewConstMetric(sslVerDesc, prometheus.GaugeValue, ver, labelValues...)

	var alertLevel, alertDescription float64
	if res.SslcertAlert() != nil {
		alertLevel = float64(res.SslcertAlert().Level())
		alertDescription = float64(res.SslcertAlert().Description())
	}
	ch <- prometheus.MustNewConstMetric(alertLevelDesc, prometheus.GaugeValue, alertLevel, labelValues...)
	ch <- prometheus.MustNewConstMetric(alertDescriptionDesc, prometheus.GaugeValue, alertDescription, labelValues...)

	if res.Rt() > 0 {
		ch <- prometheus.MustNewConstMetric(successDesc, prometheus.GaugeValue, 1, labelValues...)
		ch <- prometheus.MustNewConstMetric(rttDesc, prometheus.GaugeValue, res.Rt(), labelValues...)
	} else {
		ch <- prometheus.MustNewConstMetric(successDesc, prometheus.GaugeValue, 0, labelValues...)
	}
}

// ExportHistograms exports aggregated metrics for the measurement
func (m *sslCertMetricExporter) ExportHistograms(res []*measurement.Result, ch chan<- prometheus.Metric) {

}

// Describe exports metric descriptions for Prometheus
func (m *sslCertMetricExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- successDesc
	ch <- rttDesc
	ch <- sslVerDesc
	ch <- alertLevelDesc
	ch <- alertDescriptionDesc
}

// IsValid returns whether an result is valid or not (e.g. IPv6 measurement and Probe does not support IPv6)
func (m *sslCertMetricExporter) IsValid(res *measurement.Result, probe *probe.Probe) bool {
	return probe.ASNForIPVersion(res.Af()) > 0
}
