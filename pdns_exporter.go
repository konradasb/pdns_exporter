package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Statistic ...
type Statistic struct {
	Value json.RawMessage `json:"value"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Size  string          `json:"size,omitempty"`
}

// StatisticItem ...
type StatisticItem struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// MapStatisticItem ...
type MapStatisticItem struct {
	Value []SimpleStatisticItem `json:"value"`
	Name  string                `json:"name"`
	Type  string                `json:"type"`
}

// RingStatisticItem ...
type RingStatisticItem struct {
	Value []SimpleStatisticItem `json:"value"`
	Name  string                `json:"name"`
	Type  string                `json:"type"`
	Size  string                `json:"size"`
}

// SimpleStatisticItem ...
type SimpleStatisticItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Exporter ...
type Exporter struct {
	metrics []*prometheus.Desc
	apiURL  string
	apiKey  string
}

var (
	client = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, 5*time.Second)
				if err != nil {
					return nil, err
				}
				if err := c.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return nil, err
				}
				return c, nil
			},
		},
	}
)

var (
	listenAddress = flag.String("listen-address", ":9120", "Address to listen on for incoming connections.")
	apiURL        = flag.String("api-url", "http://localhost:8081/api/v1/servers/localhost/statistics", "PowerDNS statistics endpoint URL.")
	apiKey        = flag.String("api-key", "", "PowerDNS API Key")
)

// Describe ...
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range e.metrics {
		ch <- metric
	}
}

func (e *Exporter) scrapeStatistics() []Statistic {
	request, err := http.NewRequest("GET", e.apiURL, nil)
	if err != nil {
		log.Fatalln("Failed to create new request context:", err)
	}
	request.Header.Add("X-API-Key", e.apiKey)

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln("Failed to send request:", err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Failed to read response:", err)
	}

	// Unmarshall into abstract Statistic
	stat := []Statistic{}
	if err := json.Unmarshal(content, &stat); err != nil {
		fmt.Printf("Failed to decode JSON: %s", err)
	}

	return stat
}

func newStatisticItemMetric(name string) *prometheus.Desc {
	return prometheus.NewDesc(
		fmt.Sprintf("pdns_auth_%s", strings.ReplaceAll(name, "-", "_")),
		fmt.Sprintf("See PowerDNS statistic '%s'.", name),
		nil,
		nil,
	)
}

func newMapStatisticItemMetric(name string) *prometheus.Desc {
	return prometheus.NewDesc(
		fmt.Sprintf("pdns_auth_map_%s", strings.ReplaceAll(name, "-", "_")),
		fmt.Sprintf("See PowerDNS statistic '%s'", name),
		nil,
		nil,
	)
}

func formatSimpleStatisticItemName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), " ", ""), "-", "_")
}

// Collect ...
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	statistics := e.scrapeStatistics()

	statisticItems := []StatisticItem{}
	mapStatisticItems := []MapStatisticItem{}
	ringStatisticItems := []RingStatisticItem{}

	unmarshallStatistics(statistics, &statisticItems, &mapStatisticItems, &ringStatisticItems)

	for _, statisticItem := range statisticItems {
		value, err := strconv.ParseFloat(statisticItem.Value, 64)
		if err != nil {
			log.Fatalf("Failed to convert string into float: %s", err)
		}
		ch <- prometheus.MustNewConstMetric(newStatisticItemMetric(statisticItem.Name), prometheus.CounterValue, value)
	}

	for _, mapStatisticItem := range mapStatisticItems {
		for _, simpleStatisticItem := range mapStatisticItem.Value {
			fmt.Println("Simple")
			value, err := strconv.ParseFloat(simpleStatisticItem.Value, 64)
			if err != nil {
				log.Fatalf("Failed to convert string into float: %s", err)
			}
			ch <- prometheus.MustNewConstMetric(
				newMapStatisticItemMetric(
					fmt.Sprintf("%s_%s", mapStatisticItem.Name, formatSimpleStatisticItemName(simpleStatisticItem.Name)),
				),
				prometheus.CounterValue,
				value,
			)
		}
	}
}

// AuthExporter ...
func AuthExporter() *Exporter {
	return &Exporter{
		apiURL: *apiURL,
		apiKey: *apiKey,
	}
}

func main() {
	flag.Parse()
	run()
}

func registerExporters() {
	// Auth
	prometheus.MustRegister(AuthExporter())
}

func run() {
	registerExporters()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*listenAddress, nil)
}

func unmarshallStatistics(statistics []Statistic, statisticItems *[]StatisticItem,
	mapStatisticItems *[]MapStatisticItem, ringStatisticItems *[]RingStatisticItem) {
	for _, statistic := range statistics {
		switch statistic.Type {
		case "StatisticItem":
			statisticItem := StatisticItem{Name: statistic.Name, Type: statistic.Type}
			json.Unmarshal(statistic.Value, &statisticItem.Value)
			*statisticItems = append(*statisticItems, statisticItem)
		case "MapStatisticItem":
			mapStatisticItem := MapStatisticItem{Name: statistic.Name, Type: statistic.Type}
			json.Unmarshal(statistic.Value, &mapStatisticItem.Value)
			*mapStatisticItems = append(*mapStatisticItems, mapStatisticItem)
		case "RingStatisticItem":
			ringStatisticItem := RingStatisticItem{Name: statistic.Name, Type: statistic.Type, Size: statistic.Size}
			json.Unmarshal(statistic.Value, &ringStatisticItem.Value)
			*ringStatisticItems = append(*ringStatisticItems, ringStatisticItem)
		}
	}
}
