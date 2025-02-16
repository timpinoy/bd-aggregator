package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"log"
	"strings"
	"time"
)

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("parsing of duration failed: %w", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	log.Printf("Collecting feeds every %v\n", timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeNextFeed(s)
	}
}

func scrapeNextFeed(s *state) {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("retrieval of next feed to fetch failed: %v", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), feedToFetch.ID)
	if err != nil {
		log.Printf("marking of fetched feed failed: %v", err)
	}

	feed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		log.Printf("fetching of feed failed: %v", err)
	}

	for _, item := range feed.Channel.Item {
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("parsing of '%v' to time failed: %v", item.PubDate, err)
		}

		cpp := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feedToFetch.ID,
		}

		_, err = s.db.CreatePost(context.Background(), cpp)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("creation of post failed: %v", err)
			continue
		}
	}
	log.Printf("Aggregated %s, %v posts saved", feedToFetch.Name, len(feed.Channel.Item))
}
