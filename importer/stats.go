package importer

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

// Stats stores the statistics for a CloudantImport operation. A log of the frequency of HTTP
// status codes / error messages and a running count of documents and batches written.
type Stats struct {
	StatusCodes    map[int]int    `json:"statusCodes"`
	ErrorMessages  map[string]int `json:"errors"`
	DocsWritten    int            `json:"docs"`
	BatchesWritten int            `json:"batches"`
}

// StatsDataPoint is the result of a single write API call
type StatsDataPoint struct {
	statusCode int
	result     []cloudantv1.DocumentResult
	latency    int
}

// NewStats creates a new empty Stats struct
func NewStats() *Stats {
	stats := Stats{
		StatusCodes:    make(map[int]int, 5),
		ErrorMessages:  make(map[string]int, 5),
		DocsWritten:    0,
		BatchesWritten: 0,
	}
	return &stats
}

// Save updates the Stats struct with the latest HTTP status code and error message
// and how many documents were written
func (s *Stats) Save(statsDataPoint *StatsDataPoint) {
	successCount := 0
	failureCount := 0
	s.StatusCodes[statsDataPoint.statusCode]++
	for _, v := range statsDataPoint.result {
		if v.Error != nil {
			s.ErrorMessages[*v.Error]++
			failureCount++
		} else {
			successCount++
		}
	}
	s.DocsWritten += len(statsDataPoint.result)
	s.BatchesWritten++

	// create and output a log line
	ll := NewLogLine(statsDataPoint.statusCode, statsDataPoint.latency, successCount, failureCount)
	ll.Output()
}

// Output turns the Stats struct into JSON and outputs it
func (s *Stats) Summary() {
	jsonStr, _ := json.Marshal(s)
	fmt.Println("-------")
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Println(string(jsonStr))
}
