package main

import (
	"flag"
	"fmt"
	"github.com/janosgyerik/portping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

type Probes struct {
	Probes []probe
}

type probe struct {
	Address string
	Network string
	Timeout int
}

func (p probe) String() string {
	return fmt.Sprintf("Address: %s, Network: %s, Timeout %d", p.Address, p.Network, p.Timeout)
}

func (p probe) GetNameForVector() string {
	forbidden := [...]string{".", ":"} //ellipses
	replaceWith := "_"
	for _, char := range forbidden {
		p.Address = strings.ReplaceAll(p.Address, char, replaceWith)

	}

	return p.Address
}

type runtimeConfStruct struct {
	debug  bool
	probes []probe
}

var prometheusConfig = prometheusConfigStruct{
	gaugeVectors:   make(map[string]*prometheus.GaugeVec),
	registry:       prometheus.NewRegistry(),
	httpServerPort: 9101,
}

var runtimeConf = runtimeConfStruct{
	debug:  false,
	probes: []probe{},
}

func initParams() {
	probeString := ""
	configFile := ""
	flag.UintVar(&prometheusConfig.httpServerPort, "prometheusServerPort", prometheusConfig.httpServerPort, "Prometheus Exporter server port.")
	flag.BoolVar(&runtimeConf.debug, "debug", runtimeConf.debug, "Log Level Debug")
	flag.StringVar(&probeString, "probes", "", "List of hosts and ports to probe like 127.0.0.1:80;tcp,127.0.0.1:443;tcp,127.0.0.1:8990;tcp <host>:<port>;<Network>;<Timeout in seconds>")
	flag.StringVar(&configFile, "configFile", "", "Pass a config file with probes")
	flag.Parse()
	logLvl := log.InfoLevel
	if runtimeConf.debug {
		logLvl = log.DebugLevel
	}
	log.SetLevel(logLvl)

	if strings.TrimSpace(configFile) != "" {
		var probes = Probes{}
		yamlFile, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(yamlFile, &probes)
		if err != nil {
			log.Fatalf("error: %v", err)

		}

		for _, probe := range probes.Probes {
			runtimeConf.probes = append(runtimeConf.probes, probe)
		}
		log.Debug("Probes from File", probes)
	}

	if strings.TrimSpace(probeString) != "" {

		rawProbesFromFlag := strings.Split(probeString, ",")
		for _, rawProbe := range rawProbesFromFlag {
			hostPortNetwork := strings.Split(rawProbe, ";")
			aProbe := probe{}
			aProbe.Address = hostPortNetwork[0]
			aProbe.Network = hostPortNetwork[1]
			aProbe.Timeout, _ = strconv.Atoi(hostPortNetwork[2])
			runtimeConf.probes = append(runtimeConf.probes, aProbe)
		}
	}

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
	initParams()
	setupWebserver()

	log.Info(runtimeConf.probes)
	// declare vectors
	for _, probe := range runtimeConf.probes {
		name := probe.GetNameForVector()
		prometheusConfig.gaugeVectors[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "port_checker",
			Name:      name,
			Help:      name}, []string{"Address", "Network"})
		prometheusConfig.registry.MustRegister(prometheusConfig.gaugeVectors[name])
	}

	// loop
	for {
		for _, probe := range runtimeConf.probes {
			name := probe.GetNameForVector()

			r := portping.Ping(probe.Network, probe.Address, time.Duration(probe.Timeout)*time.Second)
			log.Debug(probe, r)

			if r == nil {
				prometheusConfig.gaugeVectors[name].WithLabelValues(probe.Address, probe.Network).Set(1)

			} else {
				prometheusConfig.gaugeVectors[name].WithLabelValues(probe.Address, probe.Network).Set(0)
			}
		}

		time.Sleep(time.Duration(1) * time.Second)

	}
}
