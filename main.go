package main

import (
	"fmt"
	"github.com/timpinoy/bd-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	err = cfg.SetUser("tim")
	if err != nil {
		panic(err)
	}

	cfg, err = config.Read()
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
