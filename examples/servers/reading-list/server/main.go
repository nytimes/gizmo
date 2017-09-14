package main

import (
	readinglist "github.com/NYTimes/gizmo/examples/servers/reading-list"
	"github.com/NYTimes/gizmo/server/kit"
)

// a tiny main package that simply initializes and initiates the server.
func main() {
	db, err := readinglist.NewDB()
	if err != nil {
		panic(err)
	}
	svc, err := readinglist.NewService(db)
	if err != nil {
		panic(err)
	}
	errors := make(chan error)
	kit.Run(svc, errors)
	err = <-errors
	if err != nil {
		panic(err)
	}
}
