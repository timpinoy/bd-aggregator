package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"time"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <feed_name> <url>", cmd.Name)
	}
	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	cfp := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      feedName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Url:       feedUrl,
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), cfp)
	if err != nil {
		return fmt.Errorf("creation of feed failed: %v", err)
	}
	fmt.Printf("Added feed %s\n", cfp)

	cffp := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), cffp)
	if err != nil {
		return fmt.Errorf("follow of feed failed: %v", err)
	}
	fmt.Printf("%s following %s\n", user.Name, feed.Name)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("retrieval of feeds failed: %v", err)
	}
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("retrieval of users failed: %v", err)
	}

	usersMap := make(map[uuid.UUID]string)
	for _, u := range users {
		usersMap[u.ID] = u.Name
	}

	for _, feed := range feeds {
		printFeed(feed, usersMap[feed.UserID])
		fmt.Println("=====================================")
	}

	return nil
}

func printFeed(feed database.Feed, username string) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", username)
	fmt.Printf("* LastFetchedAt: %v\n", feed.LastFetchedAt.Time)
}
