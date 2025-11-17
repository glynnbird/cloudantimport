# cloudantimport

## Introduction

When populating Cloudant databases, often the source of the data is initially some JSON documents in a file.

*cloudantimport* is designed to assist with importing such data into Cloudant efficiently. Simply pipe a file full of JSON documents into *cloudantimport*, telling it the database to send the data to and it will group the documents into batches and employ Cloudant's bulk import API.

## Installation

You will need to [download and install the Go compiler](https://go.dev/doc/install). Clone this repo then:

```sh
go build
```

The copy the resultant binary `cloudantimport` (or `cloudantimport.exe` in Windows systems) into your path.

## Configuration

`cloudantimport` authenticates with your chosen Cloudant service using environment variables as documented [here](https://github.com/IBM/cloudant-go-sdk/blob/v0.10.8/docs/Authentication.md#authentication-with-environment-variables) e.g.

```sh
CLOUDANT_URL=https://xxxyyy.cloudantnosqldb.appdomain.cloud
CLOUDANT_APIKEY="my_api_key"
```

## Usage

Pipe a JSON file (one document per line) into _cloudantimport_ and supply the database you want to write to using the `--dbname`/`--db` parameter:

```sh
cat myfile.json | cloudantimport --db mydb
```

## Generating random data

_cloudantimport_ can be paired with [datamaker](https://www.npmjs.com/package/datamaker) to generate any amount of sample data:

```sh
# template ---> datamaker ---> 100 JSON docs ---> cloudantimport ---> Cloudant
echo '{"_id":"{{uuid}}","name":"{{name}}","email":"{{email true}}","dob":"{{date 1950-01-01}}"}' | datamaker -f json -i 100 | cloudantimport --db people
written {"docCount":100,"successCount":1,"failCount":0,"statusCodes":{"201":1}}
written {"batch":1,"batchSize":100,"docSuccessCount":100,"docFailCount":0,"statusCodes":{"201":1},"errors":{}}
Import complete
```

or with the template as a file:

```sh
cat template.json | datamaker -f json -i 10000 | cloudantimport --db people
```

## Understanding the output

The output comes in two parts. Firstly, one line per bulk write request made:

```
2025-11-14T15:27:55Z 201 174 500 0
2025-11-14T15:27:55Z 201 186 500 0
2025-11-14T15:27:56Z 201 173 500 0
```

This shows the date/time, HTTP status code, latency (ms), number of documents successfully written and the number that failed.

Then at the end comes a summary:

```
-------
Summary
-------
{"statusCodes":{"201":20},"errors":{"conflict":10},"docs":9990,"batches":20}
```

which lists a counts of each HTTP status code, counts of document write errors, total docs written and total number of write
