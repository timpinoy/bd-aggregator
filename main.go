package main

import (
	"fmt"
	"github.com/timpinoy/bd-aggregator/internal/config"
	"log"
	"os"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	function, ok := c.commands[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return function(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	return s.cfg.SetUser(cmd.args[0])
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	s := &state{cfg: &cfg}

	c := commands{commands: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)

	if osArgs := os.Args; len(osArgs) < 2 {
		fmt.Println("Usage: cli <command> [args...]")
		return
	}

	cmd := command{name: os.Args[1], args: os.Args[2:]}
	err = c.commands[os.Args[1]](s, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
