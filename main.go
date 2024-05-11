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
	ip := ipinfo.ParseMyIPv4()
	hostInfo["IP"] = ip.Ip
	hostInfo["IPv6"] = ipinfo.ParseMyIPv6().Ip
	hostInfo["Location"] = fmt.Sprintf("%s, %s, %s (%s)", ip.City, ip.Region, ip.Country, ip.Organization)

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
