package main

import (
	"fmt"
	"blogaggregator/internal/database"
	"os"
	"context"
	"time"
	"strconv"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("username is required")
		os.Exit(1)
	}
	//check if user exists in db
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println("user not found")
		os.Exit(1)
	}

	s.cfg.SetUser(cmd.args[0])
	fmt.Println("User has been set as " + cmd.args[0])
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("name is required")
		os.Exit(1)
	}

	//create user in db
	userCreateParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.args[0],
	}
	user, err := s.db.CreateUser(context.Background(), userCreateParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//set curr user in config to given name
	s.cfg.SetUser(cmd.args[0])
	
	fmt.Println("User has been created in db, name=" + user.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Println("error resetting users table")
		os.Exit(1)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("time_between_reqs os required")
		os.Exit(1)
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		fmt.Println(err)
		return err
	}

	ticker := time.NewTicker(duration)
	fmt.Printf("Collecting feeds every %v\n", duration)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		fmt.Println("name and url are required")
		os.Exit(1)
	}
	
	//create the feed
	newFeed := database.CreateFeedParams {
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.args[0],
		Url: cmd.args[1],
		UserID: user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), newFeed)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(feed)

	//connect user to the feed
	feedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUserNames(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Printf("Feed name: %v\nFeed URL: %v\nCreated by: %v\n\n", feed.Name, feed.Url, feed.UserName.String)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		fmt.Println("feed url is required")
		os.Exit(1)
	}

	//get feed info
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//subscribe user to feed
	feedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}
	feedUser, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%v followed feed %v\n", feedUser.UserName, feedUser.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		fmt.Println("feed url is required")
		os.Exit(1)
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	deleteParams := database.DeleteFeedForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedForUser(context.Background(), deleteParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Unfollowed feed for user: %v\n", user.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) > 0 {
		parsedLimit, err2 := strconv.Atoi(cmd.args[0])
		if err2 != nil {
			fmt.Println(err2)
			os.Exit(1)
		}
		limit = parsedLimit
	}

	getParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	}

	posts, err := s.db.GetPostsForUser(context.Background(), getParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Printf("Published at: %v\n--- %v ---\n%v\n\nLink: %s\n\n=================\n", post.PublishedAt.Time, post.Title.String, post.Description.String, post.Url.String)
	}
	return nil
}
