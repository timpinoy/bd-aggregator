package main

import (
	"fmt"
	"github.com/timpinoy/bd-aggregator/internal/config"
	"log"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	err = cfg.SetUser("tim")
	if err != nil {
		panic(err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	fmt.Println(cfg)
}
