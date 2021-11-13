package main

import (
	"fmt"
	"github.com/janosgyerik/portping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type probe struct {
	address string
	network string
	timeout int
}
type prometheusConfigStruct struct {
	registry       *prometheus.Registry
	httpServerPort uint
	httpServ       *http.Server
	updateInterval time.Duration
	debug          bool
	configFile     string
	currency       string
	gaugeVectors   map[string]*prometheus.GaugeVec
}

var prometheusConfig = prometheusConfigStruct{
	gaugeVectors:   make(map[string]*prometheus.GaugeVec),
	registry:       prometheus.NewRegistry(),
	httpServerPort: 9101,
}
func setupWebserver() {

	// Register prom metrics path in http serv
	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		prometheusConfig.registry,
		promhttp.HandlerFor(prometheusConfig.registry, promhttp.HandlerOpts{}),
	))

	// Init & start serv
	prometheusConfig.httpServ = &http.Server{
		Addr:    fmt.Sprintf(":%d", prometheusConfig.httpServerPort),
		Handler: httpMux,
	}
	go func() {
		log.Infof("> Starting HTTP server at %s", prometheusConfig.httpServ.Addr)
		err := prometheusConfig.httpServ.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP Server errored out %v", err)
		}
	}()

}

func main() {
	setupWebserver()
	var probes = []probe{
		{timeout: 1, address: "127.0.0.1:80", network: "tcp"},
		{timeout: 1, address: "127.0.0.1:443", network: "tcp"},
		{timeout: 1, address: "127.0.0.1:8990", network: "tcp"},
	}

	for _, probe := range probes {

		name := strings.ReplaceAll(fmt.Sprintf("%s", probe.address), ":", "_")
		name = strings.ReplaceAll(name, ".", "_")
		prometheusConfig.gaugeVectors[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "port_checker",
			Name:      name,
			Help:      name}, []string{"address", "network"})
		prometheusConfig.registry.MustRegister(prometheusConfig.gaugeVectors[name])
	}

	for {

		for _, probe := range probes {
			name := strings.ReplaceAll(fmt.Sprintf("%s", probe.address), ":", "_")
			name = strings.ReplaceAll(name, ".", "_")
			r := portping.Ping(probe.network, probe.address, time.Duration(probe.timeout)*time.Second)
			log.Debug(probe, r)

			if r == nil {
				prometheusConfig.gaugeVectors[name].WithLabelValues(probe.address, probe.network).Set(1)

			} else {
				prometheusConfig.gaugeVectors[name].WithLabelValues(probe.address, probe.network).Set(0)

			}
		}

		time.Sleep(time.Duration(1) * time.Second)

	}
}
