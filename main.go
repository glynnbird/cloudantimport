package main

import "github.com/glynnbird/cloudantimport/importer"

func main() {

	// create a new importer
	cloudantImport, err := importer.New()
	if err != nil {
		panic(err)
	}

	// run it
	cloudantImport.Run()
}
