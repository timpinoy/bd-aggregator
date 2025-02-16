package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/timpinoy/bd-aggregator/internal/config"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"log"
	"os"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to retrieve user: %v", err)
		}

		return handler(s, cmd, user)
	}
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	s := &state{
		cfg: &cfg,
		db:  dbQueries,
	}

	c := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAggregate)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollowFeed))
	c.register("following", middlewareLoggedIn(handlerFeedsFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	c.register("browse", middlewareLoggedIn(handlerBrowse))

	if osArgs := os.Args; len(osArgs) < 2 {
		fmt.Println("Usage: cli <command> [Args...]")
		return
	}

	cmd := command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = c.run(s, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
