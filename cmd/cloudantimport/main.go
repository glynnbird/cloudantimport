package main

import importer "github.com/glynnbird/cloudantimport/internal/app"

func main() {

	// create a new importer
	cloudantImport, err := importer.New()
	if err != nil {
		panic(err)
	}

	// run it
	err = cloudantImport.Run()
	if err != nil {
		panic(err)
	}
}
