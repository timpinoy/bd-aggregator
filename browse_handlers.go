package main

import (
	"context"
	"fmt"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"strconv"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	numberPosts := 2
	var err error
	if len(cmd.Args) > 0 {
		numberPosts, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("could not parse number of posts: %v", err)
		}
	}

	p, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID:    user.ID,
		Postlimit: int32(numberPosts),
	})
	if err != nil {
		return fmt.Errorf("rettrieving of posts failed: %v", err)
	}

	fmt.Printf("Found %d posts for user %s:\n------\n", len(p), user.Name)

	for _, post := range p {
		fmt.Printf("%s\n", post.Title)
		fmt.Printf("\tURL: %s\n\n", post.Url)
		fmt.Printf("%s\n\n\n", post.Description)
		fmt.Println("=====================================")
	}

	return nil
}
