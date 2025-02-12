package main

import (
	"Blog_Aggregator/internal/config"
	"Blog_Aggregator/internal/database"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	defer db.Close()
	dbQueries := database.New(db)

	programState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", cmds.reset)
	cmds.register("users", cmds.users)
	cmds.register("agg", cmds.agg)
	cmds.register("addfeed", middlewareLoggedIn(cmds.addFeed))
	cmds.register("feeds", cmds.listFeeds)
	cmds.register("follow", middlewareLoggedIn(cmds.follow))
	cmds.register("following", middlewareLoggedIn(cmds.following))
	cmds.register("unfollow", middlewareLoggedIn(cmds.unfollow))
	cmds.register("browse", middlewareLoggedIn(cmds.browse))

	if len(os.Args) < 2 {
		log.Fatal("Usage: ./Blog_Aggregator <command> [args...]")
		return
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.run(programState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}

// TO DELETE THE DB BEFORE RUNNING TESTS:
// goose -dir sql/schema postgres "postgres://postgres:password@localhost:5432/gator" down
// goose -dir sql/schema postgres "postgres://postgres:password@localhost:5432/gator" up
