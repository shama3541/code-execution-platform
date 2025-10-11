package main

import (
	"log"

	"github.com/shama3541/code-execution-platform/api"
)

func main() {

	newserver := &api.Server{}
	newserver.CreateServer()
	log.Print("Starting web server on port 3000")
	newserver.StartServer("0.0.0.0:3000")
}
