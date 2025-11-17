package importer

import (
	"errors"
	"flag"
	"fmt"
)

// AppConfig contains the command-line options chosen by the user
type AppConfig struct {
	DatabaseName string
	Concurrency  int
}

func (ac AppConfig) Print() {
	fmt.Println("APP CONFIG")
	fmt.Println("----------")
	fmt.Printf("DatabaseName: %v\n", ac.DatabaseName)
	fmt.Printf("Concurrency: %v\n", ac.Concurrency)
}

func NewAppConfig() (*AppConfig, error) {
	appConfig := AppConfig{}

	// parse command-line options
	flag.StringVar(&appConfig.DatabaseName, "dbname", "", "The Cloudant database name to write to")
	flag.StringVar(&appConfig.DatabaseName, "db", "", "The Cloudant database name to write to")
	flag.IntVar(&appConfig.Concurrency, "concurrency", 1, "The number of concurrent HTTP write requests in flight")
	flag.IntVar(&appConfig.Concurrency, "c", 1, "The number of concurrent HTTP write requests in flight")
	flag.Parse()

	// if we don't have a database name after parsing
	if appConfig.DatabaseName == "" {
		return nil, errors.New("missing dbname/db")
	} else if appConfig.Concurrency < 1 || appConfig.Concurrency > 50 {
		return nil, errors.New("conccurrency must be between 1 and 50")
	} else {
		return &appConfig, nil
	}
}
