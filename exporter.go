package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	spamhausScore = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "spamhaus_score",
			Help: "Spamhaus score for the given domain",
		},
		[]string{"domain"},
	)
	spamhausScoreDimension = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "spamhaus_score_dimension",
			Help: "Spamhaus dimension scores for the given domain",
		},
		[]string{"domain", "dimension"},
	)
	mutex sync.Mutex
)

func init() {
	prometheus.MustRegister(spamhausScore)
	prometheus.MustRegister(spamhausScoreDimension)
}

type SpamhausResponse struct {
	Score      float64 `json:"score"`
	Dimensions struct {
		Human    float64 `json:"human"`
		Identity float64 `json:"identity"`
		Infra    float64 `json:"infra"`
		Malware  float64 `json:"malware"`
		SMTP     float64 `json:"smtp"`
	} `json:"dimensions"`
}

func fetchSpamhausData(domain string) (*SpamhausResponse, error) {
	url := "https://www.spamhaus.org/api/v1/sia-proxy/api/intel/v2/byobject/domain/" + domain + "/overview"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if the response is JSON or HTML (in case of errors)
	if body[0] == '<' {
		return nil, fmt.Errorf("unexpected HTML response for domain %s", domain)
	}

	var spamhausResponse SpamhausResponse
	err = json.Unmarshal(body, &spamhausResponse)
	if err != nil {
		return nil, err
	}

	return &spamhausResponse, nil
}

func processDomains(domains []string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Reset metrics before processing
	spamhausScore.Reset()
	spamhausScoreDimension.Reset()

	for _, domain := range domains {
		data, err := fetchSpamhausData(domain)
		if err != nil {
			log.Printf("Error fetching data for domain %s: %v", domain, err)
			continue
		}
		spamhausScore.With(prometheus.Labels{"domain": domain}).Set(data.Score)
		spamhausScoreDimension.With(prometheus.Labels{"domain": domain, "dimension": "human"}).Set(data.Dimensions.Human)
		spamhausScoreDimension.With(prometheus.Labels{"domain": domain, "dimension": "identity"}).Set(data.Dimensions.Identity)
		spamhausScoreDimension.With(prometheus.Labels{"domain": domain, "dimension": "infra"}).Set(data.Dimensions.Infra)
		spamhausScoreDimension.With(prometheus.Labels{"domain": domain, "dimension": "malware"}).Set(data.Dimensions.Malware)
		spamhausScoreDimension.With(prometheus.Labels{"domain": domain, "dimension": "smtp"}).Set(data.Dimensions.SMTP)
	}
}

func domainsHandler(w http.ResponseWriter, r *http.Request) {
	targets := r.URL.Query()["target"]
	if len(targets) == 0 {
		http.Error(w, "Missing 'target' parameter", http.StatusBadRequest)
		return
	}

	processDomains(targets)
	w.Write([]byte("Domains processed"))
}

func main() {
	port := flag.String("web.listen-address", "8080", "Address to listen on for web interface and telemetry.")
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", domainsHandler)

	log.Printf("Beginning to serve on port :%s", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
