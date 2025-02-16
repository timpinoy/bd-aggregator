package main

import (
	"context"
	"fmt"
)

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("deletion of users failed: %v", err)
	}
	fmt.Println("Users deleted")
	return nil
}
