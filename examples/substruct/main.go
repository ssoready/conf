package main

import (
	"fmt"
	"time"

	"github.com/ucarion/conf"
)

func main() {
	type dbConfig struct {
		DSN     string        `conf:"dsn"`
		Timeout time.Duration `conf:"timeout,noredact"`
	}

	config := struct {
		PrimaryDB   dbConfig `conf:"primary-db"`
		SecondaryDB dbConfig `conf:"secondary-db"`
	}{}

	conf.Load(&config)
	fmt.Println("raw config", config)
	fmt.Println("redacted config", conf.Redact(config))
}
