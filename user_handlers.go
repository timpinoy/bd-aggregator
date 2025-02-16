package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"time"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}
	username := cmd.Args[0]

	_, err := s.db.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user does not exist: %w", err)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("setting of current user failed: %w", err)
	}
	fmt.Printf("Logged in as %s\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {

		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}
	username := cmd.Args[0]

	cup := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err := s.db.CreateUser(context.Background(), cup)
	if err != nil {
		return fmt.Errorf("creation of user failed: %w", err)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("setting of current user failed: %w", err)
	}

	fmt.Printf("Registered user: %s\n", username)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("retrieving user list failed: %v", err)
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
