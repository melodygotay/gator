package main

import (
	"Blog_Aggregator/internal/database"
	"context"
	"fmt"
	"strconv"
)

func (c *commands) browse(s *state, cmd command, user database.User) error {
	// No need to get user again since middleware provides it
	limit := 2
	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("Invalid limit provided: %v\n", err)
		}
		limit = parsedLimit
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID, // Use the user provided by middleware
		Limit:  int32(limit),
	})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("\n%s\n", post.Title.String)
		fmt.Printf("Feed: %s\n", post.FeedName)
		fmt.Printf("Published: %v\n", post.PublishedAt.Format("2006-01-02 15:04:05"))
		if post.Url.Valid {
			fmt.Printf("URL: %s\n", post.Url.String)
		}
		if post.Description.Valid {
			fmt.Printf("Description: %s\n", post.Description.String)
		}
		fmt.Println("---")
	}

	return nil
}
