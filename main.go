package main

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/timpinoy/bd-aggregator/internal/database"
	"strconv"
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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("GetUserByName failed: %v", err)
		}

		return handler(s, cmd, user)
	}
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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("wrong number of arguments")
	}

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

	feed, err := s.db.CreateFeed(context.Background(), cfp)
	if err != nil {
		return fmt.Errorf("CreateFeed failed: %v", err)
	}
	fmt.Printf("Added feed %s\n", cfp)

	id, err = uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("UUID generation failed: %v", err)
	}
	cffp := database.CreateFeedFollowParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), cffp)
	if err != nil {
		return fmt.Errorf("CreateFeedFollow failed: %v", err)
	}
	fmt.Printf("%s following %s\n", user.Name, feed.Name)

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
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("parse time failed: %v", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Printf("Colleting feeds every %v\n", timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeed(s)
		if err != nil {
			return fmt.Errorf("scrape feed failed: %v", err)
		}
	}

	//return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("GetFeeds failed: %v", err)
	}
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("GetUsers failed: %v", err)
	}

	usersMap := make(map[uuid.UUID]string)
	for _, u := range users {
		usersMap[u.ID] = u.Name
	}

	for _, feed := range feeds {
		fmt.Printf("* Name: %s, URL: %s, User: %s", feed.Name, feed.Url, usersMap[feed.UserID])
	}

	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("GetFeed failed: %v", err)
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("UUID generation failed: %v", err)
	}
	cffp := database.CreateFeedFollowParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), cffp)
	if err != nil {
		return fmt.Errorf("CreateFeedFollow failed: %v", err)
	}
	fmt.Printf("%s following %s\n", user.Name, feed.Name)
	return nil
}

func handlerFeedsFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("GetFeedFollowsForUser failed: %v", err)
	}

	fmt.Printf("%s following %d feeds:\n", user.Name, len(feeds))
	for _, feed := range feeds {
		fmt.Printf("* %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("GetFeed failed: %v", err)
	}

	cufp := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), cufp)
	if err != nil {
		return fmt.Errorf("DeleteFeedFollow failed: %v", err)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	numberPosts := 2
	var err error
	if len(cmd.args) > 0 {
		numberPosts, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("could not parse number of posts: %v", err)
		}
	}

	p, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, Postlimit: int32(numberPosts)})
	if err != nil {
		return fmt.Errorf("GetPostsForUser failed: %v", err)
	}

	for _, post := range p {
		fmt.Printf("%s\n", post.Title)
		fmt.Printf("\tURL: %s\n\n", post.Url)
		fmt.Printf("%s\n\n\n", post.Description)
	}

	return nil
}

func scrapeFeed(s *state) error {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("GetNextFeedToFetch failed: %v", err)
	}

	feed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return fmt.Errorf("fetch feed: %v", err)
	}

	for _, item := range feed.Channel.Item {
		id, err := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("UUID generation failed: %v", err)
		}

		//fmt.Printf("* %s: %s\n", item.Title, time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate))

		publishedAt, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			fmt.Printf("Failed to parse published at %v: %v", item.PubDate, err)
		}

		cpp := database.CreatePostParams{
			ID:          id,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feedToFetch.ID,
		}

		_, err = s.db.CreatePost(context.Background(), cpp)
		if err != nil {
			fmt.Printf("CreatePost failed: %v", err)
		}

		err = s.db.MarkFeedFetched(context.Background(), feedToFetch.ID)
		if err != nil {
			return fmt.Errorf("MarkFeedFetched: %v", err)
		}
	}

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
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollowFeed))
	c.register("following", middlewareLoggedIn(handlerFeedsFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	c.register("browse", middlewareLoggedIn(handlerBrowse))

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
