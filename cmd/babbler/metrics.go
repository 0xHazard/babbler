package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	udpRTT = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "udp_rtt",
			Help: "UDP round trip time, nanoseconds",
		},
		[]string{"source", "destination", "location"},
	)
	tcpRTT = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tcp_rtt",
			Help: "TCP round trip time, nanoseconds",
		},
		[]string{"source", "destination", "location"},
	)
)

func updateRTT(proto, src, dst string, rtt int64) {
	switch proto {
	case "udp":
		udpRTT.With(prometheus.Labels{"source": src, "destination": dst, "location": location}).Set(float64(rtt))
	case "tcp":
		tcpRTT.With(prometheus.Labels{"source": src, "destination": dst, "location": location}).Set(float64(rtt))
	}
}

func exporter() {
	prometheus.MustRegister(udpRTT)
	prometheus.MustRegister(tcpRTT)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", mport), nil))
}
