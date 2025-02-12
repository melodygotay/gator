package main

import (
	"Blog_Aggregator/internal/database"
	"context"
	"fmt"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(s *state, cmd command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("user is not logged in: %w", err)
		}

		if user.ID == [16]byte{} {
			return fmt.Errorf("user does not exist or is not logged in")
		}

		return handler(s, cmd, user)
	}
}
