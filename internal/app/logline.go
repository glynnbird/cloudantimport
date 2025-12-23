package importer

import (
	"log"
)

// LogLine stores the data collected for a single output log line
type LogLine struct {
	StatusCode          int `json:"statusCode"`
	LatencyMilliseconds int `json:"latency"`
	DocsSuccess         int `json:"docsSuccess"`
	DocsFailed          int `json:"docsFailed"`
}

// NewLogLine creates a ne LogLine struct, given the incoming attributes
func NewLogLine(statusCode int, latency int, success int, failed int) *LogLine {
	ll := LogLine{
		StatusCode:          statusCode,
		LatencyMilliseconds: latency,
		DocsSuccess:         success,
		DocsFailed:          failed,
	}
	return &ll
}

// Output writes a single log line to stdout
func (ll *LogLine) Output() {
	log.Println(ll.StatusCode, ll.LatencyMilliseconds, ll.DocsSuccess, ll.DocsFailed)
}
