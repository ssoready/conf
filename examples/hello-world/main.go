package main

import (
	"fmt"

	"github.com/ucarion/conf"
)

func main() {
	config := struct {
		Username string `conf:"name,noredact" usage:"who to log in as"`
		Password string `conf:"password"`
	}{
		Username: "jdoe",
	}

	conf.Load(&config)
	fmt.Println("raw config", config)
	fmt.Println("redacted config", conf.Redact(config))
}
