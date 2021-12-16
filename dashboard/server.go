package main

// Serve metrics to calling process (authentication coming soon)
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pbnjay/memory"
)

type MetricResponse struct {
	Timestamp string `json:"timestamp"`
	Metric    string `json:"metric"`
	Value     uint64 `json:"value"`
}

func server() {
	r := mux.NewRouter()
	r.HandleFunc("/metrics", getMetrics)
	http.Handle("/metrics", r)
	fmt.Println("Starting up on port 7100")
	log.Fatal(http.ListenAndServe(":7100", nil))
}

func getMetrics(w http.ResponseWriter, req *http.Request) {
	system_memory := memory.TotalMemory()
	dt := time.Now().String()

	resultsMetrics := []MetricResponse{}
	setDataRow := MetricResponse{dt, "memory", system_memory}

	resultsMetrics = append(resultsMetrics, setDataRow)

	b, _ := json.Marshal(resultsMetrics)
	w.Write(b)
}
