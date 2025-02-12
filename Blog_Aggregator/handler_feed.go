package main

import (
	"Blog_Aggregator/internal/database"
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (c *commands) addFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}
	nameInput := os.Args[2]
	urlInput := os.Args[3]
	re := regexp.MustCompile(`^(.*)$`)
	nameMatch := re.FindStringSubmatch(nameInput)
	urlMatch := re.FindStringSubmatch(urlInput)
	var name, url string
	if len(nameMatch) > 1 {
		name = nameMatch[1]
	} else {
		fmt.Printf("length: %v", len(nameMatch))
		return fmt.Errorf("invalid name format")
	}
	if len(urlMatch) > 1 {
		url = urlMatch[1]
	} else {
		return fmt.Errorf("invalid url format")
	}

	feed, err := s.db.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}

	fmt.Printf("Feed created successfully!\n")
	fmt.Printf(" * ID: %v\n", feed.ID)
	fmt.Printf(" * Name: %v\n", feed.Name)
	fmt.Printf(" * URL: %v\n", feed.Url)
	fmt.Printf(" * User ID: %v\n", feed.UserID)
	fmt.Printf("Feed followed successfully!\n")
	fmt.Printf(" * User: %v\n", feedFollow.UserName)
	fmt.Printf(" * Feed: %v\n", feed.Name)

	return nil
}

func (c *commands) listFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	for _, feed := range feeds {
		printFeed(feed, s)
	}

	return nil
}

func printFeed(feed database.Feed, s *state) {
	fmt.Printf(" * ID:          %s\n", feed.ID)
	fmt.Printf(" * Name:	%v\n", feed.Name)
	fmt.Printf(" * URL: 	%v\n", feed.Url)
	userName, err := s.db.GetUserById(context.Background(), feed.UserID)
	if err != nil {
		fmt.Println("username not found")
		return
	}
	fmt.Printf(" * User:	%v\n", userName)
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println("Couldn't get next feeds to fetch", err)
		return
	}
	fmt.Println("Found a feed to fetch!")
	scrapeFeed(s.db, feed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}
	for _, item := range feedData.Channel.Item {
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// If that format doesn't work, try RFC822
			publishedAt, err = time.Parse(time.RFC822, item.PubDate)
			if err != nil {
				// If we can't parse the date, use current time as fallback
				publishedAt = time.Now().UTC()
			}
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       sql.NullString{String: item.Title, Valid: item.Title != ""},
			Url:         sql.NullString{String: item.Link, Valid: item.Link != ""},
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					continue
				}
			}
			fmt.Printf("Failed to create post %s: %v\n", item.Title, err)
			continue
		}
		fmt.Printf("Found post: %s\n", item.Title)
	}

	fmt.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}
