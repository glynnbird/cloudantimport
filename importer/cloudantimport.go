package importer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

const bufferSize = 500 // the maximum size of our internal buffer of unwritten documents

type CloudantImport struct {
	appConfig   *AppConfig                 // our command-line options
	buffer      []cloudantv1.Document      // the buffer of documents that haven't been saved to Cloudant yet
	service     *cloudantv1.CloudantV1     // the Cloudant SDK client
	bufferLen   int                        // how many strings are in our buffer
	reader      *bufio.Reader              // the input stream
	stats       *Stats                     // running statistics
	wgWorker    sync.WaitGroup             // to keep track of running goroutines
	wgCollector sync.WaitGroup             // to keep track of the collector goroutine
	resultsChan chan StatsDataPoint        // channel to carry results of API calls
	jobsChan    chan []cloudantv1.Document // channel to carry jobs, slices of Cloudant documents to write
	errorsChan  chan error                 // channel to carry errors that occurred when writing to Cloudant
}

// New creates a new CloudantImport struct, loading the CLI parameters,
// instantiating the Cloudant SDK client and creating a buffer of strings.
func New() (*CloudantImport, error) {
	// load the CLI parameters
	appConfig, err := NewAppConfig()
	if err != nil {
		return nil, err
	}

	// set up the Cloudant service
	service, err := cloudantv1.NewCloudantV1UsingExternalConfig(&cloudantv1.CloudantV1Options{})
	if err != nil {
		return nil, err
	}
	service.EnableRetries(3, 5*time.Second)

	// setup the buffer
	buffer := make([]cloudantv1.Document, bufferSize)

	// setup IO reader
	reader := bufio.NewReader(os.Stdin)

	// create a stats placeholder
	stats := NewStats()

	ci := CloudantImport{
		appConfig:   appConfig,
		buffer:      buffer,
		service:     service,
		bufferLen:   0,
		reader:      reader,
		stats:       stats,
		wgWorker:    sync.WaitGroup{},
		wgCollector: sync.WaitGroup{},
		resultsChan: make(chan StatsDataPoint),
		jobsChan:    make(chan []cloudantv1.Document, appConfig.Concurrency),
		errorsChan:  make(chan error),
	}

	return &ci, nil
}

// writeBuffer saves the stored Cloudant documents to Cloudant. It is a
// goroutine, so there are N workers - 1 per "concurrency". Each work
// loops on the jobsChan waiting to be sent batches of data.
// When the channel is closed, the workers will exit. Response data is
// transmitted back on the resultsChan, errors on the errorsChan.
func (ci *CloudantImport) writeBufferWorker() {
	// make sure we release our slot
	defer ci.wgWorker.Done()

	for job := range ci.jobsChan {
		start := time.Now()

		// write to Cloudant with POST /{db}/_bulk_docs
		postBulkDocsOptions := ci.service.NewPostBulkDocsOptions(ci.appConfig.DatabaseName)
		bulkDocs, err := ci.service.NewBulkDocs(job)
		if err != nil {
			ci.errorsChan <- err
			return
		}
		postBulkDocsOptions.SetBulkDocs(bulkDocs)
		result, response, err := ci.service.PostBulkDocs(postBulkDocsOptions)
		if err != nil {
			ci.errorsChan <- err
			return
		}
		latency := time.Since(start)

		// save the stats
		statsDataPoint := StatsDataPoint{
			statusCode: response.StatusCode,
			result:     result,
			latency:    int(latency.Milliseconds()),
		}
		ci.resultsChan <- statsDataPoint
	}
}

// statsCollector waits for data arriving back on resultsChan and
// errorsChan, aggregating results and panicking if an error occurs
func (ci *CloudantImport) statsCollector() {
	defer ci.wgCollector.Done()
	for {
		select {
		// <- returns the value of the channel and boolean ok,
		// that indicates whether the channel is open or not.
		// If ok == false, we can return - nothing more to do
		case r, ok := <-ci.resultsChan:
			if !ok {
				return
			}
			ci.stats.Save(&r)
		case err, ok := <-ci.errorsChan:
			if !ok {
				return
			}
			panic(fmt.Sprintf("ERROR: %v", err))
		}
	}
}

// checkTargetDatabase checks whether the database to be written to exists. It returns
// an error if it doesn't
func (ci *CloudantImport) checkTargetDatabase() error {
	opts := ci.service.NewGetDatabaseInformationOptions(ci.appConfig.DatabaseName)
	_, _, err := ci.service.GetDatabaseInformation(opts)
	return err
}

// Run executes a CloudantImport job, reading lines of data from stdin,
// parsing them as JSON and then turning the resultant map into a
// Cloudant document suitable for the SDKs. Up to bufferSize documents
// are bufferred in memory and written to Cloudant in bulk.
func (ci *CloudantImport) Run() error {

	// check that the target database exists
	err := ci.checkTargetDatabase()
	if err != nil {
		return errors.New("database does not exist")
	}

	// Start worker pool
	for i := 0; i < ci.appConfig.Concurrency; i++ {
		ci.wgWorker.Add(1)
		go ci.writeBufferWorker()
	}

	// spin up a goroutine to handle the results and errors
	ci.wgCollector.Add(1)
	go ci.statsCollector()

	// loop until we run out of data
	for {
		// read a line
		text, err := ci.reader.ReadString('\n')

		// if this is the last line
		if err != nil {

			// flush the buffer
			if ci.bufferLen > 0 {
				// last write
				ci.jobsChan <- ci.buffer[:ci.bufferLen]
			}

			// close the jobs channel - we're finished
			close(ci.jobsChan)
			break
		}

		// strip the line break
		text = strings.TrimSuffix(text, "\n")
		text = strings.TrimSuffix(text, "\r")

		// if we have more than a blank line
		if len(text) > 0 {
			// parse the document and turn it into a format suitable for the SDKs
			var dataMap map[string]interface{}
			err := json.Unmarshal([]byte(text), &dataMap)
			if err != nil {
				// if it doesn't parse as JSON, skip to the next line
				continue
			}

			// generate a Cloudant doc
			doc := cloudantv1.Document{}
			doc.SetProperties(dataMap)

			// add it to the buffer
			ci.buffer[ci.bufferLen] = doc
			ci.bufferLen++

			// if the buffer is full
			if ci.bufferLen == bufferSize {
				// write to the jobs channel
				// note to self - we have to clone the slice here because we will go on to
				// reuse the underlying buffer which if we didn't clone, would  modify
				// the data that the goroutine at the other end of the channel will see
				clone := make([]cloudantv1.Document, ci.bufferLen)
				copy(clone, ci.buffer[:ci.bufferLen])
				ci.jobsChan <- clone
				ci.bufferLen = 0
			}
		}
	}

	// wait for the in-flight requests to complete
	ci.wgWorker.Wait()
	close(ci.resultsChan)
	close(ci.errorsChan)
	ci.wgCollector.Wait()

	// generate final summary
	ci.stats.Summary()

	return nil
}
