package main

import (
	"fmt"
	"blogaggregator/internal/config"
	"blogaggregator/internal/database"
	"errors"
	"os"
	"database/sql"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	// runs a given command with the provided state if it exists
	cmdToRun, ok := c.commandMap[cmd.name]
	if !ok {
		return errors.New("command not found")
	}
	err := cmdToRun(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	// registers a new handler function for a command name
	c.commandMap[name] = f
}

func main() {
	c := config.Read()
	db, err := sql.Open("postgres", c.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s := state{
		cfg: &c,
		db: dbQueries,
	}
	cmds := commands{
		commandMap: map[string]func(*state, command)error{},
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		fmt.Println("not enough arguments provided")
		os.Exit(1)
	}

	comm := command{
		name: os.Args[1], 
		args: os.Args[2:],
	}
	err = cmds.run(&s, comm)
	if err != nil {
		fmt.Println(err)
	}
}
