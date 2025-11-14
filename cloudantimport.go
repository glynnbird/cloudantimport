package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

const bufferSize = 500 // the maximum size of our internal buffer of unwritten documents

type CloudantImport struct {
	appConfig *AppConfig             // our command-line options
	buffer    []cloudantv1.Document  // the buffer of documents that aven't been saved to Cloudant yet
	service   *cloudantv1.CloudantV1 // the Cloudant SDK client
	bufferLen int                    // how many strings are in our buffer
	reader    *bufio.Reader          // the input stream
	stats     *Stats                 // running statistics
}

// NewCloudantImport creates a new CloudantImport struct, loading the CLI parameters,
// instantiating the Cloudant SDK client and creating a buffer of strings.
func NewCloudantImport() (*CloudantImport, error) {
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

	// setup the buffer
	buffer := make([]cloudantv1.Document, bufferSize)

	// setup IO reader
	reader := bufio.NewReader(os.Stdin)

	// create a stats placeholder
	stats := NewStats()

	ci := CloudantImport{
		appConfig: appConfig,
		buffer:    buffer,
		service:   service,
		bufferLen: 0,
		reader:    reader,
		stats:     stats,
	}

	return &ci, nil
}

// writeBuffer saves the stored Cloudant documents to Cloudant
func (ci *CloudantImport) writeBuffer() {
	// this is the slice of data to write
	documents := ci.buffer[:ci.bufferLen]

	// write to Cloudant with POST /{db}/_bulk_docs
	postBulkDocsOptions := ci.service.NewPostBulkDocsOptions(ci.appConfig.DatabaseName)
	bulkDocs, err := ci.service.NewBulkDocs(documents)
	if err != nil {
		panic(err)
	}
	postBulkDocsOptions.SetBulkDocs(bulkDocs)
	result, response, err := ci.service.PostBulkDocs(postBulkDocsOptions)
	if err != nil {
		panic(err)
	}
	ci.stats.Save(response.StatusCode, result)
	ci.stats.Output()
}

// Run executes a CloudantImport job, reading lines of data from stdin,
// parsing them as JSON and then turning the resultant map into a
// Cloudant document suitable for the SDKs. Up to bufferSize documents
// are bufferred in memory and written to Cloudant in bulk.
func (ci *CloudantImport) Run() {
	// loop until we run out of data
	for {
		// read a line
		text, err := ci.reader.ReadString('\n')

		// if this is the last line
		if err != nil {

			// flush the buffer
			if ci.bufferLen > 0 {
				ci.writeBuffer()
			}
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
				// write it to Cloudant and reset the buffer
				ci.writeBuffer()
				ci.bufferLen = 0
			}
		}
	}
}
