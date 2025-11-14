package main

import (
	"errors"
	"flag"
	"fmt"
)

// we have only one command-line option, but you never know in the future
type AppConfig struct {
	DatabaseName string
}

func (ac AppConfig) Print() {
	fmt.Println("APP CONFIG")
	fmt.Println("----------")
	fmt.Printf("DatabaseName: %v\n", ac.DatabaseName)
}

func NewAppConfig() (*AppConfig, error) {
	appConfig := AppConfig{}

	// parse command-line options
	flag.StringVar(&appConfig.DatabaseName, "dbname", "", "The Cloudant database name to write to")
	flag.StringVar(&appConfig.DatabaseName, "db", "", "The Cloudant database name to write to")
	flag.Parse()

	// if we don't have a database name after parsing
	if appConfig.DatabaseName == "" {
		return nil, errors.New("missing dbname/db")
	} else {
		return &appConfig, nil
	}
}
