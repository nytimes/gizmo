package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/config/mysql"
	"github.com/NYTimes/gizmo/examples/servers/mysql-saved-items/service"
	"github.com/NYTimes/gizmo/server"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var cfg struct {
		Server *server.Config
		MySQL  *mysql.Config
	}
	// load from the local JSON file into a config.Config struct
	config.LoadJSONFile("./config.json", &cfg)
	// SetConfigOverrides will allow us to override some of the values in
	// the JSON file with CLI flags.
	server.SetConfigOverrides(cfg.Server)

	// initialize Gizmo’s server with given configs
	server.Init("nyt-saved-items", cfg.Server)

	// instantiate a new ‘saved items service’ with our MySQL credentials
	svc, err := service.NewSavedItemsService(cfg.MySQL)
	if err != nil {
		server.Log.Fatal("unable to create saved items service: ", err)
	}

	// register our saved item service with the Gizmo server
	err = server.Register(svc)
	if err != nil {
		server.Log.Fatal("unable to register saved items service: ", err)
	}

	// run the Gizmo server
	err = server.Run()
	if err != nil {
		server.Log.Fatal("unable to run saved items service: ", err)
	}
}
