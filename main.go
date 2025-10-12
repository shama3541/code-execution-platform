package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/shama3541/code-execution-platform/api"
)

const (
	dburl = "postgres://root:mysecret@localhost:5432/code-execution?sslmode=disable"
)

func main() {
	// inmem := database.NewInmemorystruct()
	dbconn, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatalf("Unable to connect to DB:%v", err)
	}

	newserver := api.NewServer(dbconn)
	newserver.CreateServer()
	newserver.StartServer("0.0.0.0:3000")

	log.Print("Starting web server on port 3000")

}
