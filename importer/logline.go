package importer

import (
	"fmt"
	"time"
)

// LogLine stores the data collected for a single log
type LogLine struct {
	Date                time.Time `json:"date"`
	StatusCode          int       `json:"statusCode"`
	LatencyMilliseconds int       `json:"latency"`
	DocsSuccess         int       `json:"docsSuccess"`
	DocsFailed          int       `json:"docsFailed"`
}

func NewLogLine(statusCode int, latency int, success int, failed int) *LogLine {
	t := time.Now()
	ll := LogLine{
		Date:                t,
		StatusCode:          statusCode,
		LatencyMilliseconds: latency,
		DocsSuccess:         success,
		DocsFailed:          failed,
	}
	return &ll
}

// Output writes a single log line to stdout
func (ll *LogLine) Output() {
	fmt.Println(ll.Date.Format(time.RFC3339), ll.StatusCode, ll.LatencyMilliseconds, ll.DocsSuccess, ll.DocsFailed)
}
