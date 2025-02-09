package main

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"time"
)

import (
	"fmt"
	"github.com/timpinoy/bd-aggregator/internal/config"
	"log"
	"os"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
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

	u, err := s.db.GetUserByName(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user does not exist: %w", err)
	}

	return s.cfg.SetUser(u.Name)
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("UUID generation failed: %v", err)
	}

	cup := database.CreateUserParams{
		ID:        id,
		Name:      cmd.args[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.db.CreateUser(context.Background(), cup)
	if err != nil {
		return fmt.Errorf("CreateUser failed: %v", err)
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("SetUser failed: %v", err)
	}

	fmt.Printf("Registered user: %s\n", cmd.args[0])
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("wrong number of arguments")
	}

	user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)

	id, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("UUID generation failed: %v", err)
	}
	cfp := database.CreateFeedParams{
		ID:        id,
		Name:      cmd.args[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Url:       cmd.args[1],
		UserID:    user.ID,
	}

	_, err = s.db.CreateFeed(context.Background(), cfp)
	if err != nil {
		return fmt.Errorf("CreateFeed failed: %v", err)
	}
	fmt.Printf("Added feed %s\n", cfp)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("DeleteUsers failed: %v", err)
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("GetUsers failed: %v", err)
	}

	for _, u := range users {
		if u.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("fetch feed: %v", err)
	}
	fmt.Println(feed)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	dbQueries := database.New(db)

	s := &state{cfg: &cfg, db: dbQueries}

	c := commands{commands: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAggregate)
	c.register("addfeed", handlerAddFeed)

	if osArgs := os.Args; len(osArgs) < 2 {
		fmt.Println("Usage: cli <command> [args...]")
		return
	}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	f, ok := c.commands[os.Args[1]]
	if !ok {
		log.Fatalf("command not found\n")
	}
	err = f(s, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
