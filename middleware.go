package main

import (
	"fmt"
	"blogaggregator/internal/database"
	"os"
	"context"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			fmt.Println("user not found")
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}