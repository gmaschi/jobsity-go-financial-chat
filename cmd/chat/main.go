package main

import (
	"database/sql"
	chatfactory "github.com/gmaschi/jobsity-go-financial-chat/internal/factories/chat"
	usersdb "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/postgresql/users"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"log"
)

func main() {
	config, err := env.NewConfig()
	if err != nil {
		log.Fatalln("cannot load env variables")
	}

	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatalln("could not connect to database:", err)
	}

	store := usersdb.NewStore(conn)
	server, err := chatfactory.New(config, store)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatalln("cannot start server", err)
	}
}
