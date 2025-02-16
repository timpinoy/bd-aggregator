package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"time"
)

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("retrieval of feed failed: %v", err)
	}

	cffp := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), cffp)
	if err != nil {
		return fmt.Errorf("following of feed failed: %v", err)
	}

	fmt.Println("Now following feed:")
	printFeedFollow(user.Name, feed.Name)
	return nil
}

func handlerFeedsFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("retrieval of feed follows failed: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("User is not following any feeds.")
		return nil
	}

	fmt.Printf("%s following %d feeds:\n", user.Name, len(feeds))
	for _, feed := range feeds {
		fmt.Printf("* %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("retrieval of feed failed: %v", err)
	}

	cufp := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), cufp)
	if err != nil {
		return fmt.Errorf("deletion of follow failed: %v", err)
	}

	fmt.Printf("User %s unfollowed feed %s", user.ID, cmd.Args[0])
	return nil
}

func printFeedFollow(username, feedName string) {
	fmt.Printf("* User:          %s\n", username)
	fmt.Printf("* Feed:          %s\n", feedName)
}
