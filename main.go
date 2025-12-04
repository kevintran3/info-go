package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	clouddetector "github.com/kevintran3/cloud-detector"
	ipinfo "github.com/kevintran3/ip-info"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	kt3InfoIP = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kt3_info_ip",
		Help: "KT3 info (1=running) labeled by the Kubernetes node's public IP.",
	}, []string{"ip", "country", "country_code", "region", "region_code", "city", "org"})
	kt3InfoIPv6 = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kt3_info_ipv6",
		Help: "KT3 info (1=running) labeled by the Kubernetes node's public IP.",
	}, []string{"ip"})
	myIPv4 = ipinfo.ParseMyIPv4()
	myIPv6 = ipinfo.ParseMyIPv6()
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	return port
}

func formatMapStringandSort(stringmap map[string]string, toSort bool) []string {
	str := []string{}
	for h, v := range stringmap {
		if len(string(v)) > 0 {
			str = append(str, fmt.Sprintf("  %s: %s", h, v))
		}
	}
	if toSort {
		sort.Strings(str)
	}
	return str
}

func requestToStr(req *http.Request) string {
	hostInfo := clouddetector.GetHostInfo()
	hostInfo["IP"] = myIPv4.Ip
	hostInfo["IPv6"] = myIPv6.Ip
	hostInfo["Location"] = fmt.Sprintf("%s, %s, %s (%s)", myIPv4.City, myIPv4.Region, myIPv4.Country, myIPv4.Organization)

	var headers = make(map[string]string)
	var logheaders = make(map[string]interface{})
	for h, v := range req.Header {
		headers[h] = strings.Join(v, ", ")
		logheaders[h] = strings.Join(v, ", ")
	}

	requests := map[string]string{
		"Client IP": ipinfo.ParseClientIP(req).Str,
		"Host":      req.Host,
		"Method":    req.Method,
		"Path":      req.URL.Path,
		"Protocol":  req.Proto,
	}

	infoStr := "<pre>"
	infoStr += "SERVER\n" + strings.Join(formatMapStringandSort(hostInfo, true), "\n")
	infoStr += "\nREQUEST\n" + strings.Join(formatMapStringandSort(requests, true), "\n")
	infoStr += "\nHEADERS\n" + strings.Join(formatMapStringandSort(headers, true), "\n")
	infoStr += "</pre>"
	return infoStr
}

func main() {
	log.SetFlags(0)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/info", infoHandler)

	kt3InfoIPMetric := kt3InfoIP.WithLabelValues(myIPv4.Ip, myIPv4.Country, myIPv4.CountryCode3, myIPv4.Region, myIPv4.RegionCode, myIPv4.City, myIPv4.Organization)
	kt3InfoIPMetric.Set(1.0)
	kt3InfoIPv6Metric := kt3InfoIPv6.WithLabelValues(myIPv6.Ip)
	kt3InfoIPv6Metric.Set(1.0)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":"+getPort(), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, ipinfo.ClientIP(r))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, requestToStr(r))
}
