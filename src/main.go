package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/benridley/wls_go/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

// Config represents the main application config
type Config struct {
	CertPath   string              `yaml:"tls_cert_path"` // Certificate used for TLS, should include CA chain if its signed.
	Keypath    string              `yaml:"tls_key_path"`  // Private Key used for TLS
	ListenPort string              `yaml:"listen_port"`   // Port used to listen for scrape requests
	Queries    exporter.MbeanQuery `yaml:"queries"`       // Queries of mBeans the exporter tries to scrape
}

// Available log levels
const (
	LOG_INFO = iota
	LOG_DEBUG
)

func main() {
	configPath := flag.String("config-file", "config.yaml", "Configuration file path")
	logLevelConfig := flag.String("log-level", "info", "Log level to use. Info or Debug.")
	flag.Parse()

	configBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err.Error())
	}

	var logLevel int
	if strings.ToLower(*logLevelConfig) == "info" {
		logLevel = LOG_INFO
	} else if strings.ToLower(*logLevelConfig) == "debug" {
		logLevel = LOG_DEBUG
	} else {
		log.Fatalf("Unknown log level %s", *logLevelConfig)
	}

	config := Config{}
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	if config.ListenPort == "" {
		config.ListenPort = "9325"
	}

	e, err := exporter.New(config.Queries, logLevel)
	if err != nil {
		log.Fatalf("Unable to start exporter: %s", err.Error())
	} else {
		log.Printf("Started exporter on port %s", config.ListenPort)
	}

	http.HandleFunc("/probe", func(resp http.ResponseWriter, req *http.Request) {
		probeHandler(resp, req, &e)
	})

	if config.CertPath != "" {
		log.Fatal(http.ListenAndServeTLS(":"+config.ListenPort, config.CertPath, config.Keypath, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+config.ListenPort, nil))
	}
}

func probeHandler(resp http.ResponseWriter, req *http.Request, e *exporter.Exporter) {
	params := req.URL.Query()
	host := params.Get("host")
	port := params.Get("port")

	portInt, err := strconv.Atoi(port)
	if host == "" || port == "" {
		http.Error(resp, "Missing required parameter: Please provide host and port parameters.", 400)
		return
	} else if err != nil {
		http.Error(resp, "Unable to convert port to integer, please provide a valid value for port", 400)
		return
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		http.Error(resp, "Missing authentication information. Please provide basic authentication credentials.", 400)
		return
	}

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "weblogic_probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	registry := prometheus.NewRegistry()
	metrics, err := e.DoQuery(host, portInt, username, password)
	if err != nil {
		probeSuccessGauge.Set(0)
		if e.LogLevel >= LOG_DEBUG {
			log.Printf("Failed to probe weblogic instance %s:%s: %v", host, port, err.Error())
		}
		registry.MustRegister(probeSuccessGauge)
	} else {
		probeSuccessGauge.Set(1)
		for _, metric := range metrics {
			registry.MustRegister(metric)
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(resp, req)
}
