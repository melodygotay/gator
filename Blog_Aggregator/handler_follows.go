package main

import (
	"Blog_Aggregator/internal/database"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func (c *commands) follow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: %s <feed_id>", cmd.Name)
	}

	feedID, err := uuid.Parse(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid feed id: %w", err)
	}

	feed, err := s.db.GetFeed(context.Background(), feedID)
	if err != nil {
		return fmt.Errorf("couldn't find feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	fmt.Printf("Following feed successfully!\n")
	fmt.Printf(" * User: %v\n", feedFollow.UserName)
	fmt.Printf(" * Feed: %v\n", feed.Name)

	return nil
}

func (c *commands) following(s *state, cmd command, user database.User) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no arguments required")
	}

	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error fetching the user's followed feeds: %w", err)
	}

	for _, follow := range feedFollows {
		fmt.Println(follow.FeedName)
	}

	return nil
}

func (c *commands) unfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	urlInput := os.Args[2]

	feed, err := s.db.GetFeedByURL(context.Background(), urlInput)
	if err != nil {
		return fmt.Errorf("couldn't find feed: %w", err)
	}

	feedFollow, err := s.db.RemoveFeedFollow(context.Background(), database.RemoveFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})

	if err != nil {
		return fmt.Errorf("error unfollowing feed: %w", err)
	}

	fmt.Printf("%v has unfollowed %s successfully!\n", feedFollow.UserName, feed.Name)

	return nil
}
