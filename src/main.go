package main

import (
	"flag"
	"github.com/benridley/wls_go/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"strconv"
)

// errorRegistry stores the number of seen errors for a host/port combo.
// On a successful scrape, the entry is deleted. Errors will be logged
// up to 3 times before no longer logging.
var errorRegistry = make(map[string]int)

// Number of times to log an error between successful scrapes.
const errLogCount = 10

func main() {
	configPath := flag.String("config-file", "config.yaml", "Configuration file path")
	host := flag.String("host", "127.0.0.1", "IP Address of the Weblogic Server instance to scrape")
	port := flag.String("port", "7001", "Port Address of the Weblogic Server instance to scrape")
	listenAddress := flag.String("web.listen-address", ":9601", "Address to listen on for web interface and telemetry.")
	userName := flag.String("username", getEnv("USERNAME", ""), "Username for Weblogic Server")
	password := flag.String("password", getEnv("PASSWORD", ""), "Password for Weblogic Server")

	flag.Parse()

	if *host == "" || *port == "" {
		log.Fatalf("Missing required parameter: Please provide host and port parameters.")
		return
	}

	configBytes, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err.Error())
	}

	config := exporter.Config{
		Host:          *host,
		Port:          *port,
		ListenAddress: *listenAddress,
		UserName:      *userName,
		Password:      *password,
	}
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	exporter, err := exporter.New(config)
	if err != nil {
		log.Fatalf("Unable to start exporter: %s", err.Error())
	}

	http.HandleFunc("/metrics", func(resp http.ResponseWriter, req *http.Request) {
		probeHandler(resp, req, &exporter)
	})

	if config.CertPath != "" {
		log.Fatal(http.ListenAndServeTLS(config.ListenAddress, config.CertPath, config.Keypath, nil))
	} else {
		log.Fatal(http.ListenAndServe(config.ListenAddress, nil))
	}
}

func probeHandler(resp http.ResponseWriter, req *http.Request, e *exporter.Exporter) {
	port := e.Config.Port
	host := e.Config.Host
	portInt, err := strconv.Atoi(port)

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "weblogic_probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	registry := prometheus.NewRegistry()
	metrics, err := e.DoQuery(host, portInt, e.Config.UserName, e.Config.Password)
	if err != nil {
		probeSuccessGauge.Set(0)
		// Check if we've seen this error already while failing scrapes. If not, log it.
		if numErrs, ok := errorRegistry[(host + port)]; ok {
			if numErrs < errLogCount {
				log.Printf("Failed to probe weblogic instance %s:%s: %v", host, port, err.Error())
				errorRegistry[host+port]++
				if errorRegistry[host+port] == errLogCount {
					log.Printf("Pausing logging of errors until a successful scrape occurs on %s:%s...", host, port)
				}
			}
		} else {
			// No errors seen yet
			log.Printf("Failed to probe weblogic instance %s:%s: %v", host, port, err.Error())
			errorRegistry[host+port] = 1
		}
		registry.MustRegister(probeSuccessGauge)
	} else {
		delete(errorRegistry, host+port)
		probeSuccessGauge.Set(1)
		registry.MustRegister(probeSuccessGauge)
		for _, metric := range metrics {
			registry.MustRegister(metric)
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(resp, req)
}

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}
